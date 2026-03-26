package logto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ErrTokenExpired indicates the Logto access token is expired or invalid.
var ErrTokenExpired = errors.New("logto access token expired")

// GitHubToken retrieves the user's GitHub access token from the Logto Secret Vault
// via the Account API. Requires the user's Logto access token (not the M2M token).
// Logto auto-refreshes the GitHub token if expired and a refresh token is available.
func (c *Client) GitHubToken(ctx context.Context, logtoAccessToken string) (string, error) {
	reqURL := c.endpoint + "/api/my-account/identities/github/access-token"
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Account API request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+logtoAccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Account API request failed: %w", err)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	ct := resp.Header.Get("Content-Type")

	if resp.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("%w: %s", ErrTokenExpired, string(b))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Account API returned %d (content-type: %s): %s", resp.StatusCode, ct, string(b))
	}

	if !strings.Contains(ct, "application/json") {
		body := string(b)
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		return "", fmt.Errorf("Account API returned non-JSON (status %d, content-type: %s): %s", resp.StatusCode, ct, body)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(b, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode GitHub token response (status %d, content-type: %s): %w", resp.StatusCode, ct, err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty GitHub access token in response")
	}
	return tokenResp.AccessToken, nil
}
