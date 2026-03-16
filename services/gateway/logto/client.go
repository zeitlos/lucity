package logto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client wraps the Logto Management API and Account API.
// Uses M2M (machine-to-machine) authentication for Management API calls
// and user access tokens for Account API calls.
type Client struct {
	endpoint   string // e.g. "https://id.lucity.cloud"
	m2mAppID   string
	m2mSecret  string
	httpClient *http.Client

	mu       sync.Mutex
	m2mToken string
	m2mExpAt time.Time
}

// New creates a Logto API client.
func New(endpoint, m2mAppID, m2mSecret string) *Client {
	return &Client{
		endpoint:   strings.TrimRight(endpoint, "/"),
		m2mAppID:   m2mAppID,
		m2mSecret:  m2mSecret,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// m2mAccessToken returns a cached M2M access token, refreshing if expired.
func (c *Client) m2mAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.m2mToken != "" && time.Now().Before(c.m2mExpAt) {
		return c.m2mToken, nil
	}

	data := url.Values{
		"grant_type": {"client_credentials"},
		"resource":   {c.endpoint + "/api"},
		"scope":      {"all"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/oidc/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.SetBasicAuth(c.m2mAppID, c.m2mSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("M2M token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("M2M token request returned %d: %s", resp.StatusCode, string(b))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	c.m2mToken = tokenResp.AccessToken
	// Refresh 60 seconds before expiry
	c.m2mExpAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	slog.Debug("M2M access token refreshed", "expires_in", tokenResp.ExpiresIn)
	return c.m2mToken, nil
}

// doManagement executes an authenticated Management API request.
func (c *Client) doManagement(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	token, err := c.m2mAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get M2M token: %w", err)
	}

	reqURL := c.endpoint + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
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

// doJSON executes a Management API request and decodes the JSON response.
func (c *Client) doJSON(ctx context.Context, method, path string, body io.Reader, v interface{}) error {
	resp, err := c.doManagement(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logto %s %s returned %d: %s", method, path, resp.StatusCode, string(b))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response from %s: %w", path, err)
		}
	}
	return nil
}

// doNoContent executes a Management API request that expects no response body.
func (c *Client) doNoContent(ctx context.Context, method, path string, body io.Reader) error {
	resp, err := c.doManagement(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logto %s %s returned %d: %s", method, path, resp.StatusCode, string(b))
	}
	return nil
}
