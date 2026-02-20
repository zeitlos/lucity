package argocd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is a thin HTTP client for the ArgoCD REST API.
type Client struct {
	baseURL    string // e.g., "http://localhost:8443"
	token      string
	httpClient *http.Client
}

// NewClient creates an ArgoCD API client.
func NewClient(baseURL, token string, insecure bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Transport: transport},
	}
}

// CreateApplication creates a new ArgoCD Application.
func (c *Client) CreateApplication(ctx context.Context, app Application) (*Application, error) {
	body, err := json.Marshal(app)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal application: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v1/applications", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var result Application
	if err := c.do(req, &result); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}
	return &result, nil
}

// Application retrieves an ArgoCD Application by name.
func (c *Client) Application(ctx context.Context, name string) (*Application, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/api/v1/applications/"+name, nil)
	if err != nil {
		return nil, err
	}

	var result Application
	if err := c.do(req, &result); err != nil {
		return nil, fmt.Errorf("failed to get application %s: %w", name, err)
	}
	return &result, nil
}

// DeleteApplication removes an ArgoCD Application.
// If cascade is true, all managed resources are deleted too.
func (c *Client) DeleteApplication(ctx context.Context, name string, cascade bool) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		c.baseURL+"/api/v1/applications/"+name, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	if cascade {
		q.Set("cascade", "true")
	}
	req.URL.RawQuery = q.Encode()

	if err := c.do(req, nil); err != nil {
		return fmt.Errorf("failed to delete application %s: %w", name, err)
	}
	return nil
}

// SyncApplication triggers a sync for an ArgoCD Application.
func (c *Client) SyncApplication(ctx context.Context, name string) (*Application, error) {
	body, err := json.Marshal(SyncRequest{Prune: true})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v1/applications/"+name+"/sync", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var result Application
	if err := c.do(req, &result); err != nil {
		return nil, fmt.Errorf("failed to sync application %s: %w", name, err)
	}
	return &result, nil
}

// DeleteRepository removes a Git repository credential from ArgoCD.
// Idempotent: returns success if the repository doesn't exist.
func (c *Client) DeleteRepository(ctx context.Context, repoURL string) error {
	encoded := url.QueryEscape(repoURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		c.baseURL+"/api/v1/repositories/"+encoded, nil)
	if err != nil {
		return err
	}

	if err := c.do(req, nil); err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return fmt.Errorf("failed to delete repository: %w", err)
	}
	return nil
}

// CreateRepository registers a Git repository in ArgoCD.
// Idempotent: returns success if the repository already exists.
func (c *Client) CreateRepository(ctx context.Context, repo Repository) error {
	body, err := json.Marshal(repo)
	if err != nil {
		return fmt.Errorf("failed to marshal repository: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v1/repositories", bytes.NewReader(body))
	if err != nil {
		return err
	}

	if err := c.do(req, nil); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create repository: %w", err)
	}
	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
