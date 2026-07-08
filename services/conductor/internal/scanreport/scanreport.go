package scanreport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Finding struct {
	Rule     string `json:"rule"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Commit   string `json:"commit"`
	Secret   string `json:"secret"`
	Author   string `json:"author,omitempty"`
	Verified bool   `json:"verified"`
	URL      string `json:"-"`
}

type Report struct {
	Commit    string    `json:"commit"`
	ScannedAt time.Time `json:"scannedAt"`
	Findings  []Finding `json:"findings"`
}

type Config struct {
	Endpoint     string
	DialEndpoint string
	Keychain     authn.Keychain
}

type Client struct {
	config Config
}

func New(config Config) *Client {
	if config.DialEndpoint == "" {
		config.DialEndpoint = config.Endpoint
	}

	return &Client{config: config}
}

// Latest returns the most recent secret-scan report, or nil when the
// service has never been scanned.
func (c *Client) Latest(ctx context.Context, service platform.ServiceID) (*Report, error) {
	repoPath := service.ImageRepository() + "/scans"

	repo, err := name.NewRepository(c.config.DialEndpoint+"/"+repoPath, name.Insecure)

	if err != nil {
		return nil, err
	}

	auth := authn.Anonymous

	if c.config.Keychain != nil {
		endpoint, err := name.NewRegistry(c.config.Endpoint)

		if err != nil {
			return nil, err
		}

		auth, err = c.config.Keychain.Resolve(endpoint)

		if err != nil {
			return nil, err
		}
	}

	img, err := remote.Image(repo.Tag("secrets-latest"), remote.WithContext(ctx), remote.WithAuth(auth))

	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("fetch scan report: %w", err)
	}

	layers, err := img.Layers()

	if err != nil || len(layers) == 0 {
		return nil, fmt.Errorf("scan report has no layers: %w", err)
	}

	reader, err := layers[0].Uncompressed()

	if err != nil {
		return nil, err
	}

	defer reader.Close()

	data, err := io.ReadAll(io.LimitReader(reader, 4<<20))

	if err != nil {
		return nil, err
	}

	report := new(Report)

	if err := json.Unmarshal(data, report); err != nil {
		return nil, fmt.Errorf("parse scan report: %w", err)
	}

	return report, nil
}

func isNotFound(err error) bool {
	var transportErr *transport.Error

	if errors.As(err, &transportErr) {
		return transportErr.StatusCode == 404
	}

	return strings.Contains(err.Error(), "NAME_UNKNOWN") || strings.Contains(err.Error(), "MANIFEST_UNKNOWN")
}
