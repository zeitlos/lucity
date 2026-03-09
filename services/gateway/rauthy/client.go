package rauthy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client wraps the Rauthy admin REST API.
type Client struct {
	baseURL    string // e.g. "https://id.lucity.cloud/auth/v1"
	apiKey     string
	httpClient *http.Client
}

// New creates a Rauthy API client.
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// do executes an authenticated request to Rauthy and returns the response.
// The caller is responsible for closing resp.Body.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", path, err)
	}

	return resp, nil
}

// doJSON executes an authenticated request and decodes the JSON response into v.
func (c *Client) doJSON(ctx context.Context, method, path string, body io.Reader, v interface{}) error {
	resp, err := c.do(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rauthy %s %s returned %d: %s", method, path, resp.StatusCode, string(b))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response from %s: %w", path, err)
		}
	}
	return nil
}

// doNoContent executes an authenticated request that expects no response body.
func (c *Client) doNoContent(ctx context.Context, method, path string, body io.Reader) error {
	resp, err := c.do(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rauthy %s %s returned %d: %s", method, path, resp.StatusCode, string(b))
	}
	return nil
}
