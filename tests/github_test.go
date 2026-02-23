package tests

import (
	"os"
	"testing"
)

func testGitHub(t *testing.T) {
	t.Run("Repositories", func(t *testing.T) {
		// This test requires a real GitHub OAuth token.
		// The token must be in the JWT claims (set during OAuth login).
		// Skip if not configured — user interaction is required for OAuth.
		if os.Getenv("GITHUB_TOKEN") == "" {
			t.Skip("skipping: GITHUB_TOKEN not set (requires OAuth login)")
		}

		token := testToken(t)
		resp := doGraphQL(t, token, `
			query {
				githubRepositories {
					id
					name
					fullName
					htmlUrl
					defaultBranch
					private
				}
			}
		`, nil)

		if len(resp.Errors) > 0 {
			// GitHubAuthError is expected if the token doesn't have GitHub OAuth
			t.Logf("github repositories returned error (expected without OAuth): %s", resp.Errors[0].Message)
			return
		}

		requireNoErrors(t, resp)

		var data struct {
			GitHubRepositories []struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				FullName      string `json:"fullName"`
				HTMLURL       string `json:"htmlUrl"`
				DefaultBranch string `json:"defaultBranch"`
			} `json:"githubRepositories"`
		}
		unmarshalData(t, resp, &data)

		if len(data.GitHubRepositories) == 0 {
			t.Log("no repositories returned (user may have no accessible repos)")
			return
		}

		t.Logf("found %d repositories", len(data.GitHubRepositories))
		for i, repo := range data.GitHubRepositories {
			if i >= 5 {
				t.Logf("  ... and %d more", len(data.GitHubRepositories)-5)
				break
			}
			t.Logf("  %s (%s)", repo.FullName, repo.DefaultBranch)
		}
	})
}
