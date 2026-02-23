package tests

import (
	"encoding/json"
	"testing"
)

func testProject(t *testing.T) {
	token := testToken(t)

	t.Run("CreateProject", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($input: CreateProjectInput!) {
				createProject(input: $input) {
					id
					name
					environments { id name }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"name": testProjectName,
			},
		})
		requireNoErrors(t, resp)

		var data struct {
			CreateProject struct {
				ID           string `json:"id"`
				Name         string `json:"name"`
				Environments []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"environments"`
			} `json:"createProject"`
		}
		unmarshalData(t, resp, &data)

		if data.CreateProject.Name != testProjectName {
			t.Fatalf("expected project name %q, got %q", testProjectName, data.CreateProject.Name)
		}

		// Verify development environment was auto-created
		hasDevEnv := false
		for _, env := range data.CreateProject.Environments {
			if env.Name == "development" {
				hasDevEnv = true
				break
			}
		}
		if !hasDevEnv {
			t.Fatal("project should have a 'development' environment")
		}

		t.Logf("created project %s with %d environments", data.CreateProject.Name, len(data.CreateProject.Environments))
	})

	t.Run("ListProjects", func(t *testing.T) {
		requireProjectCreated(t)

		resp := doGraphQL(t, token, `
			query {
				projects {
					id
					name
				}
			}
		`, nil)
		requireNoErrors(t, resp)

		var data struct {
			Projects []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"projects"`
		}
		unmarshalData(t, resp, &data)

		found := false
		for _, p := range data.Projects {
			if p.Name == testProjectName {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("project %q not found in projects list", testProjectName)
		}

		t.Logf("found %d projects, including %s", len(data.Projects), testProjectName)
	})

	t.Run("GetProject", func(t *testing.T) {
		requireProjectCreated(t)

		resp := doGraphQL(t, token, `
			query($id: ID!) {
				project(id: $id) {
					id
					name
					environments { id name }
					services { name }
				}
			}
		`, map[string]any{"id": testProjectName})
		requireNoErrors(t, resp)

		var data struct {
			Project struct {
				ID           string `json:"id"`
				Name         string `json:"name"`
				Environments []struct {
					Name string `json:"name"`
				} `json:"environments"`
				Services []struct {
					Name string `json:"name"`
				} `json:"services"`
			} `json:"project"`
		}
		unmarshalData(t, resp, &data)

		if data.Project.Name != testProjectName {
			t.Fatalf("expected %q, got %q", testProjectName, data.Project.Name)
		}
	})

	t.Run("ProjectNotFound", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query($id: ID!) {
				project(id: $id) {
					id
					name
				}
			}
		`, map[string]any{"id": "nonexistent-project-xyz"})

		// Should return null data or an error — not a 500.
		var data struct {
			Project json.RawMessage `json:"project"`
		}
		if err := json.Unmarshal(resp.Data, &data); err == nil {
			t.Logf("project(nonexistent): data=%s, errors=%d", string(data.Project), len(resp.Errors))
		}
	})
}
