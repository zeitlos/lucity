package tests

import (
	"testing"
	"time"
)

func testEnvironment(t *testing.T) {
	requireProjectCreated(t)
	token := testToken(t)

	t.Run("CreateEnvironment", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($input: CreateEnvironmentInput!) {
				createEnvironment(input: $input) {
					id
					name
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":       testProjectName,
				"name":            "staging",
				"fromEnvironment": "development",
			},
		})
		requireNoErrors(t, resp)

		name := extractString(t, resp.Data, "createEnvironment", "name")
		if name != "staging" {
			t.Fatalf("expected environment name 'staging', got %q", name)
		}
		t.Logf("created environment: staging")

		// kubectl: verify namespace (ArgoCD sync can take a few minutes)
		if devNamespaceReady {
			waitForNamespaceOK(t, namespace("staging"), 3*time.Minute)
			assertResourceExists(t, "application.argoproj.io", testProjectName+"-staging", "lucity-system")
		}
	})

	t.Run("SyncChart", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!) {
				syncChart(projectId: $projectId)
			}
		`, map[string]any{"projectId": testProjectName})
		requireNoErrors(t, resp)

		synced := extractBool(t, resp.Data, "syncChart")
		if !synced {
			t.Fatal("syncChart returned false")
		}
		t.Log("chart synced")
	})

	t.Run("DeleteEnvironment", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $environment: String!) {
				deleteEnvironment(projectId: $projectId, environment: $environment)
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "staging",
		})

		requireNoErrors(t, resp)

		deleted := extractBool(t, resp.Data, "deleteEnvironment")
		if !deleted {
			t.Fatal("deleteEnvironment returned false")
		}

		// kubectl: verify namespace is gone (or terminating)
		if devNamespaceReady {
			waitForNamespaceGone(t, namespace("staging"), 60*time.Second)
			assertResourceGone(t, "application.argoproj.io", testProjectName+"-staging", "lucity-system")
		}

		t.Log("staging environment deleted")
	})
}
