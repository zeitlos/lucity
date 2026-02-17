package tests

import (
	"encoding/json"
	"testing"
)

func TestMeQuery(t *testing.T) {
	token := testToken(t)

	resp := doGraphQL(t, token, `query { me { login name email avatarUrl } }`, nil)
	requireNoErrors(t, resp)

	var data struct {
		Me struct {
			Login     string `json:"login"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			AvatarURL string `json:"avatarUrl"`
		} `json:"me"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to decode me data: %v", err)
	}

	if data.Me.Login != "testuser" {
		t.Errorf("expected login 'testuser', got %q", data.Me.Login)
	}
	if data.Me.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", data.Me.Email)
	}
}

func TestUnauthenticatedQuery(t *testing.T) {
	resp := doGraphQL(t, "", `query { me { login } }`, nil)

	if len(resp.Errors) == 0 {
		t.Fatal("expected error for unauthenticated query, got none")
	}
}
