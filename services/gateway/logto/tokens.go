package logto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GitHubToken retrieves the user's GitHub access token from the Logto Secret Vault
// via the Account API. Requires the user's Logto access token (not the M2M token).
// Logto auto-refreshes the GitHub token if expired and a refresh token is available.
func (c *Client) GitHubToken(ctx context.Context, logtoAccessToken string) (string, error) {
	reqURL := c.endpoint + "/my-account/identities/github/access-token"
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

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("no GitHub identity found for user")
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("Logto access token expired")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Account API returned %d: %s", resp.StatusCode, string(b))
	}

	var tokenResp struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode GitHub token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty GitHub access token in response")
	}
	return tokenResp.AccessToken, nil
}

// RefreshAccessToken uses a Logto refresh token to obtain a new access token.
// Returns the new access token and optionally a new refresh token.
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {c.m2mAppID},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/oidc/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("refresh token request returned %d: %s", resp.StatusCode, string(b))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", "", fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return tokenResp.AccessToken, tokenResp.RefreshToken, nil
}
