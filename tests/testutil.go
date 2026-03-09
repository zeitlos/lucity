package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
)

const (
	defaultGatewayURL = "http://localhost:8080"
	defaultJWTSecret  = "change-me-in-production"
)

func gatewayURL() string {
	if u := os.Getenv("GATEWAY_URL"); u != "" {
		return u
	}
	return defaultGatewayURL
}

func jwtSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return defaultJWTSecret
}

// testToken generates a JWT token for integration tests.
func testToken(t *testing.T) string {
	t.Helper()
	token, err := makeToken()
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}
	return token
}

// makeToken generates a JWT token without requiring a *testing.T (for cleanup).
func makeToken() (string, error) {
	claims := &auth.Claims{
		Subject:     "test-user",
		Email:       "test@example.com",
		GitHubLogin: "testuser",
		AvatarURL:   "https://github.com/testuser.png",
		Roles:       []auth.Role{auth.RoleUser, auth.RoleAdmin},
		GitHubToken: githubToken(),
	}
	return auth.NewToken(claims, jwtSecret(), 1*time.Hour)
}

// githubToken returns a GitHub OAuth token for API tests.
// Checks GITHUB_TOKEN env var first, then falls back to `gh auth token`.
func githubToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	// Try gh CLI — check common install locations
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		// Try common Homebrew locations
		for _, p := range []string{"/opt/homebrew/bin/gh", "/usr/local/bin/gh"} {
			if _, err := os.Stat(p); err == nil {
				ghPath = p
				break
			}
		}
	}
	if ghPath == "" {
		return ""
	}
	out, err := exec.Command(ghPath, "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// doGraphQL sends a GraphQL query to the gateway and returns the parsed response.
func doGraphQL(t *testing.T, token, query string, variables map[string]any) *graphqlResponse {
	t.Helper()
	resp, err := doGraphQLRaw(token, query, variables)
	if err != nil {
		t.Fatalf("graphql request failed: %v", err)
	}
	return resp
}

// doGraphQLRaw sends a GraphQL query without requiring a *testing.T (for cleanup).
func doGraphQLRaw(token, query string, variables map[string]any) (*graphqlResponse, error) {
	body, err := json.Marshal(graphqlRequest{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", gatewayURL()+"/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Lucity-Workspace", "default")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp graphqlResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("decode response: %s", string(respBody))
	}

	return &gqlResp, nil
}

// requireNoErrors fails the test if the GraphQL response contains errors.
func requireNoErrors(t *testing.T, resp *graphqlResponse) {
	t.Helper()
	if len(resp.Errors) > 0 {
		msgs := make([]string, len(resp.Errors))
		for i, e := range resp.Errors {
			msgs[i] = e.Message
		}
		t.Fatalf("graphql errors: %v", msgs)
	}
}

// requireErrors fails the test if the GraphQL response has no errors.
func requireErrors(t *testing.T, resp *graphqlResponse) {
	t.Helper()
	if len(resp.Errors) == 0 {
		t.Fatal("expected graphql errors but got none")
	}
}

// unmarshalData decodes the GraphQL response data into target.
func unmarshalData(t *testing.T, resp *graphqlResponse, target any) {
	t.Helper()
	if err := json.Unmarshal(resp.Data, target); err != nil {
		t.Fatalf("failed to decode response data: %v\nraw: %s", err, string(resp.Data))
	}
}

// extractString extracts a string value from a nested JSON path.
// Example: extractString(t, resp.Data, "createProject", "id")
func extractString(t *testing.T, raw json.RawMessage, keys ...string) string {
	t.Helper()
	var current json.RawMessage = raw
	for _, key := range keys {
		var m map[string]json.RawMessage
		if err := json.Unmarshal(current, &m); err != nil {
			t.Fatalf("failed to extract key %q: %v", key, err)
		}
		val, ok := m[key]
		if !ok {
			t.Fatalf("key %q not found in response", key)
		}
		current = val
	}

	var s string
	if err := json.Unmarshal(current, &s); err != nil {
		t.Fatalf("value at path is not a string: %s", string(current))
	}
	return s
}

// extractBool extracts a boolean value from a nested JSON path.
func extractBool(t *testing.T, raw json.RawMessage, keys ...string) bool {
	t.Helper()
	var current json.RawMessage = raw
	for _, key := range keys {
		var m map[string]json.RawMessage
		if err := json.Unmarshal(current, &m); err != nil {
			t.Fatalf("failed to extract key %q: %v", key, err)
		}
		val, ok := m[key]
		if !ok {
			t.Fatalf("key %q not found in response", key)
		}
		current = val
	}

	var b bool
	if err := json.Unmarshal(current, &b); err != nil {
		t.Fatalf("value at path is not a bool: %s", string(current))
	}
	return b
}

// doHTTP performs an HTTP request and returns the response body and status code.
func doHTTP(t *testing.T, method, url, token string) ([]byte, int) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return body, resp.StatusCode
}

// requireProjectCreated fatals if the test project hasn't been created yet.
func requireProjectCreated(t *testing.T) {
	t.Helper()
	if testProjectName == "" {
		t.Fatal("test project not created — earlier test must have failed")
	}
}

// namespace returns the Kubernetes namespace for the test project and environment.
func namespace(env string) string {
	return "default-" + testProjectName + "-" + env
}
