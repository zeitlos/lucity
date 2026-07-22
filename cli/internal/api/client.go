package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseURL = "https://lucity.cloud"

const WorkspaceHeader = "X-Lucity-Workspace"

type Session struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	LogtoToken   string    `json:"logtoToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type CookieTokenSource interface {
	CookieTokens(ctx context.Context) (logtoToken, refreshToken string)
}

type RotatedTokenSink interface {
	PersistRotatedTokens(sessionToken, logtoToken, refreshToken string)
}

type WorkspaceMembership struct {
	Workspace string `json:"workspace"`
	Role      string `json:"role"`
}

type Identity struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	Email      string                `json:"email"`
	AvatarURL  string                `json:"avatarUrl"`
	Workspaces []WorkspaceMembership `json:"workspaces"`
}

type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

type StaticToken string

func (t StaticToken) Token(context.Context) (string, error) { return string(t), nil }

type Client struct {
	BaseURL   string
	Workspace string
	Tokens    TokenSource
	HTTP      *http.Client
}

func NewClient(baseURL, workspace string, tokens TokenSource) *Client {
	return &Client{
		BaseURL:   strings.TrimSuffix(baseURL, "/"),
		Workspace: workspace,
		Tokens:    tokens,
		HTTP:      &http.Client{Timeout: 60 * time.Second},
	}
}

type GraphQLError struct {
	Message    string         `json:"message"`
	Path       []any          `json:"path"`
	Extensions map[string]any `json:"extensions"`
}

type RequestError struct {
	Errors []GraphQLError
}

func (e *RequestError) Error() string {
	messages := make([]string, len(e.Errors))
	for i, gqlErr := range e.Errors {
		messages[i] = gqlErr.Message
	}
	return strings.Join(messages, "; ")
}

func (c *Client) GraphQL(ctx context.Context, query string, variables map[string]any, out any) error {
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/graphql", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	token, err := c.Tokens.Token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if c.Workspace != "" {
		req.Header.Set(WorkspaceHeader, c.Workspace)
	}

	if source, ok := c.Tokens.(CookieTokenSource); ok {
		logtoToken, refreshToken := source.CookieTokens(ctx)
		var cookies []string
		if logtoToken != "" {
			cookies = append(cookies, "lucity_token="+logtoToken)
		}
		if refreshToken != "" {
			cookies = append(cookies, "lucity_refresh="+refreshToken)
		}
		if len(cookies) > 0 {
			req.Header.Set("Cookie", strings.Join(cookies, "; "))
		}
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()

	if sink, ok := c.Tokens.(RotatedTokenSink); ok {
		var sessionToken, logtoToken, refreshToken string
		for _, cookie := range resp.Cookies() {
			switch cookie.Name {
			case "lucity_session":
				sessionToken = cookie.Value
			case "lucity_token":
				logtoToken = cookie.Value
			case "lucity_refresh":
				refreshToken = cookie.Value
			}
		}
		if sessionToken != "" || logtoToken != "" || refreshToken != "" {
			sink.PersistRotatedTokens(sessionToken, logtoToken, refreshToken)
		}
	}

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return err
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []GraphQLError  `json:"errors"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return fmt.Errorf("unexpected response from %s (HTTP %d): %s", c.BaseURL, resp.StatusCode, truncate(string(payload), 300))
	}
	if len(envelope.Errors) > 0 {
		return &RequestError{Errors: envelope.Errors}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, c.BaseURL, truncate(string(payload), 300))
	}
	if out != nil {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) Me(ctx context.Context) (*Identity, error) {
	token, err := c.Tokens.Token(ctx)
	if err != nil {
		return nil, err
	}
	return Me(ctx, c.HTTP, c.BaseURL, token)
}

func Me(ctx context.Context, httpClient *http.Client, baseURL, token string) (*Identity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(baseURL, "/")+"/auth/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	var identity Identity
	if err := doJSON(httpClient, req, &identity); err != nil {
		return nil, err
	}
	return &identity, nil
}

type AuthConfig struct {
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
}

func Config(ctx context.Context, httpClient *http.Client, baseURL string) (*AuthConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(baseURL, "/")+"/auth/config", nil)
	if err != nil {
		return nil, err
	}
	var cfg AuthConfig
	if err := doJSON(httpClient, req, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func ExchangeCode(ctx context.Context, httpClient *http.Client, baseURL, code string) (*Session, error) {
	return postAuthJSON(ctx, httpClient, strings.TrimSuffix(baseURL, "/")+"/auth/cli/exchange", "", map[string]string{"code": code})
}

func Refresh(ctx context.Context, httpClient *http.Client, baseURL, sessionToken, refreshToken string) (*Session, error) {
	return postAuthJSON(ctx, httpClient, strings.TrimSuffix(baseURL, "/")+"/auth/refresh", sessionToken, map[string]string{"refreshToken": refreshToken})
}

func postAuthJSON(ctx context.Context, httpClient *http.Client, url, bearer string, body any) (*Session, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	var session Session
	if err := doJSON(httpClient, req, &session); err != nil {
		return nil, err
	}
	if session.Token == "" {
		return nil, fmt.Errorf("no token in response from %s", url)
	}
	return &session, nil
}

func doJSON(httpClient *http.Client, req *http.Request, out any) error {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, req.URL.Path, truncate(strings.TrimSpace(string(payload)), 200))
	}
	return json.Unmarshal(payload, out)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
