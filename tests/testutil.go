package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
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
	claims := &auth.Claims{
		Subject:     "test-user",
		Email:       "test@example.com",
		GitHubLogin: "testuser",
		AvatarURL:   "https://github.com/testuser.png",
		Roles:       []auth.Role{auth.RoleUser, auth.RoleAdmin},
	}
	token, err := auth.NewToken(claims, jwtSecret(), 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}
	return token
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

	body, err := json.Marshal(graphqlRequest{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", gatewayURL()+"/graphql", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("graphql request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp graphqlResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		t.Fatalf("failed to decode response: %s", string(respBody))
	}

	return &gqlResp
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
