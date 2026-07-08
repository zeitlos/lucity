// Package ciauth implements keyless authentication from GitHub Actions.
package ciauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var ErrNoIDToken = errors.New("GitHub Actions OIDC is unavailable — add `permissions: id-token: write` to the workflow job")

type Session struct {
	Token     string
	ExpiresAt time.Time
	Workspace string
	Services  []string
}

func Available() bool {
	return os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL") != "" && os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN") != ""
}

func Detected() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func Exchange(ctx context.Context, httpClient *http.Client, apiURL string) (*Session, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	audience := strings.TrimSuffix(apiURL, "/")

	idToken, err := fetchIDToken(ctx, httpClient, audience)
	if err != nil {
		return nil, err
	}

	return exchangeToken(ctx, httpClient, audience, idToken)
}

func fetchIDToken(ctx context.Context, httpClient *http.Client, audience string) (string, error) {
	requestURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	requestToken := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")

	if requestURL == "" || requestToken == "" {
		return "", ErrNoIDToken
	}

	parsed, err := url.Parse(requestURL)
	if err != nil {
		return "", fmt.Errorf("invalid ACTIONS_ID_TOKEN_REQUEST_URL: %w", err)
	}
	query := parsed.Query()
	query.Set("audience", audience)
	parsed.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request GitHub OIDC token: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub OIDC token request failed (HTTP %d)", resp.StatusCode)
	}

	var body struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.Value == "" {
		return "", errors.New("GitHub OIDC token response was empty")
	}

	return body.Value, nil
}

func exchangeToken(ctx context.Context, httpClient *http.Client, apiURL, idToken string) (*Session, error) {
	data, err := json.Marshal(map[string]string{"token": idToken})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/auth/ci/exchange", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange OIDC token at %s: %w", apiURL, err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CI deploy not authorized: %s", strings.TrimSpace(string(payload)))
	}

	var body struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expiresAt"`
		Workspace string    `json:"workspace"`
		Services  []string  `json:"services"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.Token == "" {
		return nil, errors.New("unexpected response from CI exchange endpoint")
	}

	return &Session{
		Token:     body.Token,
		ExpiresAt: body.ExpiresAt,
		Workspace: body.Workspace,
		Services:  body.Services,
	}, nil
}
