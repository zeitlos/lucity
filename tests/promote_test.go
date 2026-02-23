package tests

import (
	"testing"
	"time"
)

func testPromote(t *testing.T) {
	requireProjectCreated(t)
	requireNamespace(t)
	token := testToken(t)

	if testBuildTag == "" {
		t.Skip("no build tag — build/deploy must have failed")
	}

	// Create a staging environment for promotion testing
	t.Run("CreateStagingForPromotion", func(t *testing.T) {
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
			t.Fatalf("expected 'staging', got %q", name)
		}

		waitForNamespace(t, namespace("staging"), 60*time.Second)
		t.Log("staging environment created for promotion test")
	})

	t.Run("Promote", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($input: PromoteInput!) {
				promote(input: $input) {
					name
					environment
					imageTag
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":       testProjectName,
				"service":         testServiceName,
				"fromEnvironment": "development",
				"toEnvironment":   "staging",
			},
		})
		requireNoErrors(t, resp)

		imageTag := extractString(t, resp.Data, "promote", "imageTag")
		if imageTag == "" {
			t.Fatal("promote returned empty imageTag")
		}
		t.Logf("promoted %s from development to staging: tag=%s", testServiceName, imageTag)
	})

	// Clean up staging
	t.Run("DeleteStagingAfterPromotion", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $environment: String!) {
				deleteEnvironment(projectId: $projectId, environment: $environment)
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "staging",
		})
		requireNoErrors(t, resp)

		waitForNamespaceGone(t, namespace("staging"), 60*time.Second)
		t.Log("staging environment cleaned up")
	})
}
