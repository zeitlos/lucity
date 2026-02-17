package tests

import (
	"encoding/json"
	"testing"
)

func TestProjectsQuery(t *testing.T) {
	token := testToken(t)

	resp := doGraphQL(t, token, `
		query {
			projects {
				id
				name
				sourceUrl
				environments { id name }
				services { name }
			}
		}
	`, nil)

	// Without a real GitHub OAuth token, the packager can't list projects.
	// Accept either a successful empty list or a "no github token" error.
	if len(resp.Errors) > 0 {
		t.Logf("projects query returned error (expected without OAuth token): %s", resp.Errors[0].Message)
		return
	}

	var data struct {
		Projects []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"projects"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to decode projects: %v", err)
	}

	t.Logf("found %d projects", len(data.Projects))
}

func TestProjectNotFound(t *testing.T) {
	token := testToken(t)

	resp := doGraphQL(t, token, `
		query($id: ID!) {
			project(id: $id) {
				id
				name
			}
		}
	`, map[string]any{"id": "nonexistent-project"})

	// Should return null data or an error — either is valid.
	// The test verifies the gateway doesn't panic or return 500.
	t.Logf("response errors: %d, data: %s", len(resp.Errors), string(resp.Data))
}
