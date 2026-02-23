package tests

import (
	"testing"
)

func testGitHub(t *testing.T) {
	t.Run("Repositories", func(t *testing.T) {
		// This test requires a GitHub OAuth token embedded in the JWT.
		// The token is auto-detected from GITHUB_TOKEN env or `gh auth token`.
		if githubToken() == "" {
			t.Skip("skipping: no GitHub token (set GITHUB_TOKEN or run `gh auth login`)")
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
			t.Fatal("expected at least one repository")
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
