package bunny

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/objectstorage"
)

const apiBase = "https://api.bunny.net"

type Backend struct {
	inner  objectstorage.Backend
	apiKey string
	domain string
	http   *http.Client
}

func New(inner objectstorage.Backend, apiKey, domain string) *Backend {
	return &Backend{
		inner:  inner,
		apiKey: apiKey,
		domain: domain,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

var _ objectstorage.Backend = (*Backend)(nil)

func (b *Backend) CreateBucket(ctx context.Context, req objectstorage.BucketRequest) (objectstorage.BucketConnection, error) {
	return b.inner.CreateBucket(ctx, req)
}

func (b *Backend) DeleteBucket(ctx context.Context, region, physicalName string) error {
	return b.inner.DeleteBucket(ctx, region, physicalName)
}

func (b *Backend) Stats(ctx context.Context) (map[string]objectstorage.BucketStats, error) {
	return b.inner.Stats(ctx)
}

func (b *Backend) Credentials(ctx context.Context, req objectstorage.CredentialsRequest) (objectstorage.BucketConnection, error) {
	return b.inner.Credentials(ctx, req)
}

func (b *Backend) SetPublic(ctx context.Context, req objectstorage.SetPublicRequest) (objectstorage.SetPublicResult, error) {
	if !req.Public {
		if err := b.unpublish(ctx, req); err != nil {
			return objectstorage.SetPublicResult{}, err
		}

		if _, err := b.inner.Credentials(ctx, objectstorage.CredentialsRequest{
			Workspace:  req.Workspace,
			Permission: objectstorage.ReadOnly,
			ReadScope:  req.PublicSet,
		}); err != nil {
			return objectstorage.SetPublicResult{}, fmt.Errorf("sync read credentials: %w", err)
		}

		return objectstorage.SetPublicResult{}, nil
	}

	origin, err := b.inner.Credentials(ctx, objectstorage.CredentialsRequest{
		Workspace:  req.Workspace,
		Permission: objectstorage.ReadOnly,
		ReadScope:  req.PublicSet,
	})

	if err != nil {
		return objectstorage.SetPublicResult{}, fmt.Errorf("read-only credentials: %w", err)
	}

	pullZoneID, err := b.ensurePullZone(ctx, req, origin)

	if err != nil {
		return objectstorage.SetPublicResult{}, err
	}

	hostname := req.Slug + "." + b.domain

	if err := b.addHostname(ctx, pullZoneID, hostname); err != nil {
		return objectstorage.SetPublicResult{}, err
	}

	go b.loadFreeCertificate(hostname)

	return objectstorage.SetPublicResult{
		PullZoneID:     pullZoneID,
		PublicEndpoint: "https://" + hostname,
	}, nil
}

func (b *Backend) ensurePullZone(ctx context.Context, req objectstorage.SetPublicRequest, origin objectstorage.BucketConnection) (string, error) {
	if req.PullZoneID != "" {
		return req.PullZoneID, nil
	}

	existing, err := b.findPullZone(ctx, req.Slug)

	if err != nil {
		return "", err
	}

	if existing != "" {
		return existing, nil
	}

	var created struct {
		ID int64 `json:"Id"`
	}

	if err := b.do(ctx, http.MethodPost, "/pullzone", map[string]any{
		"Name":                 req.Slug,
		"OriginUrl":            req.OriginURL,
		"OriginType":           0,
		"AWSSigningEnabled":    true,
		"AWSSigningKey":        origin.AccessKeyID,
		"AWSSigningSecret":     origin.SecretAccessKey,
		"AWSSigningRegionName": req.Region,
	}, &created); err != nil {
		return "", fmt.Errorf("create pull zone: %w", err)
	}

	return strconv.FormatInt(created.ID, 10), nil
}

func (b *Backend) findPullZone(ctx context.Context, name string) (string, error) {
	var page struct {
		Items []struct {
			ID   int64  `json:"Id"`
			Name string `json:"Name"`
		} `json:"Items"`
	}

	if err := b.do(ctx, http.MethodGet, "/pullzone?perPage=100&search="+url.QueryEscape(name), nil, &page); err != nil {
		return "", fmt.Errorf("search pull zones: %w", err)
	}

	for _, item := range page.Items {
		if item.Name == name {
			return strconv.FormatInt(item.ID, 10), nil
		}
	}

	return "", nil
}

func (b *Backend) addHostname(ctx context.Context, pullZoneID, hostname string) error {
	if err := b.do(ctx, http.MethodPost, "/pullzone/"+pullZoneID+"/addHostname", map[string]string{"Hostname": hostname}, nil); err != nil {
		if isAlreadyExists(err) {
			return nil
		}

		return fmt.Errorf("add hostname: %w", err)
	}

	return nil
}

func (b *Backend) loadFreeCertificate(hostname string) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	for attempt := 0; attempt < 12; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				slog.WarnContext(ctx, "bunny free certificate not issued within budget", "hostname", hostname)
				return
			case <-time.After(30 * time.Second):
			}
		}

		if err := b.do(ctx, http.MethodGet, "/pullzone/loadFreeCertificate?hostname="+url.QueryEscape(hostname), nil, nil); err == nil {
			slog.InfoContext(ctx, "bunny free certificate issued", "hostname", hostname)
			return
		}
	}

	slog.WarnContext(ctx, "bunny free certificate not issued within budget", "hostname", hostname)
}

func (b *Backend) unpublish(ctx context.Context, req objectstorage.SetPublicRequest) error {
	id := req.PullZoneID

	if id == "" {
		found, err := b.findPullZone(ctx, req.Slug)

		if err != nil {
			return err
		}

		id = found
	}

	if id == "" {
		return nil
	}

	if err := b.do(ctx, http.MethodDelete, "/pullzone/"+id, nil, nil); err != nil {
		if isNotFound(err) {
			return nil
		}

		return fmt.Errorf("delete pull zone: %w", err)
	}

	return nil
}

func (b *Backend) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader

	if body != nil {
		data, err := json.Marshal(body)

		if err != nil {
			return err
		}

		reader = bytes.NewReader(data)
	}

	request, err := http.NewRequestWithContext(ctx, method, apiBase+path, reader)

	if err != nil {
		return err
	}

	request.Header.Set("AccessKey", b.apiKey)
	request.Header.Set("Accept", "application/json")

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := b.http.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	data, _ := io.ReadAll(response.Body)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &apiError{status: response.StatusCode, body: string(data)}
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

type apiError struct {
	status int
	body   string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("bunny api error %d: %s", e.status, e.body)
}

func isNotFound(err error) bool {
	var apiErr *apiError
	return errors.As(err, &apiErr) && apiErr.status == http.StatusNotFound
}

func isAlreadyExists(err error) bool {
	var apiErr *apiError

	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.status == http.StatusConflict ||
		(apiErr.status == http.StatusBadRequest && strings.Contains(strings.ToLower(apiErr.body), "already"))
}
