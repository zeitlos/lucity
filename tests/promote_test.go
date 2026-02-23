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

	// Create a staging environment for promotion testing.
	// The environment test may have already created and (partially) deleted staging,
	// so handle "already exists" gracefully.
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

		if len(resp.Errors) > 0 {
			// Staging may already exist from the environment test if delete failed
			t.Logf("createEnvironment returned error (staging may already exist): %s", resp.Errors[0].Message)
		} else {
			name := extractString(t, resp.Data, "createEnvironment", "name")
			if name != "staging" {
				t.Fatalf("expected 'staging', got %q", name)
			}
		}

		waitForNamespaceOK(t, namespace("staging"), 60*time.Second)
		t.Log("staging environment ready for promotion test")
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

	// Clean up staging (best-effort — don't fail the whole suite if this fails)
	t.Run("DeleteStagingAfterPromotion", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $environment: String!) {
				deleteEnvironment(projectId: $projectId, environment: $environment)
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "staging",
		})

		if len(resp.Errors) > 0 {
			t.Logf("deleteEnvironment error (non-fatal): %s", resp.Errors[0].Message)
			return
		}

		waitForNamespaceGone(t, namespace("staging"), 60*time.Second)
		t.Log("staging environment cleaned up")
	})
}
