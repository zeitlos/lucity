package tests

import (
	"testing"
	"time"
)

func testDeploy(t *testing.T) {
	requireProjectCreated(t)
	requireNamespace(t)
	token := testToken(t)

	t.Run("Deploy", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($input: DeployInput!) {
				deploy(input: $input) {
					id
					phase
					buildId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"service":     testServiceName,
				"environment": "development",
			},
		})
		requireNoErrors(t, resp)

		deployID := extractString(t, resp.Data, "deploy", "id")
		phase := extractString(t, resp.Data, "deploy", "phase")
		t.Logf("deploy started: id=%s phase=%s", deployID, phase)

		// Poll deploy status until completion
		deadline := time.Now().Add(7 * time.Minute)
		for time.Now().Before(deadline) {
			time.Sleep(3 * time.Second)

			statusResp := doGraphQL(t, token, `
				query($id: ID!) {
					deployStatus(id: $id) {
						id
						phase
						imageRef
						digest
						error
						rolloutHealth
						rolloutMessage
					}
				}
			`, map[string]any{"id": deployID})
			requireNoErrors(t, statusResp)

			var data struct {
				DeployStatus struct {
					ID             string  `json:"id"`
					Phase          string  `json:"phase"`
					ImageRef       *string `json:"imageRef"`
					Error          *string `json:"error"`
					RolloutHealth  *string `json:"rolloutHealth"`
					RolloutMessage *string `json:"rolloutMessage"`
				} `json:"deployStatus"`
			}
			unmarshalData(t, statusResp, &data)

			t.Logf("deploy %s: phase=%s", deployID, data.DeployStatus.Phase)

			switch data.DeployStatus.Phase {
			case "SUCCEEDED":
				if data.DeployStatus.ImageRef != nil {
					testBuildTag = extractTagFromImageRef(*data.DeployStatus.ImageRef)
					t.Logf("deploy succeeded: image=%s", *data.DeployStatus.ImageRef)
				}

				// kubectl: verify deployment exists and pod is running
				assertResourceExists(t, "deployment", testServiceName, namespace("development"))
				waitForPod(t, namespace("development"), "app.kubernetes.io/name="+testServiceName, 60*time.Second)
				t.Log("deployment verified in Kubernetes")
				return
			case "FAILED":
				errMsg := ""
				if data.DeployStatus.Error != nil {
					errMsg = *data.DeployStatus.Error
				}
				t.Fatalf("deploy failed: %s", errMsg)
			}
		}

		t.Fatal("deploy timed out after 7 minutes")
	})

	t.Run("DeployBuild", func(t *testing.T) {
		if testBuildTag == "" {
			t.Skip("no build tag — build or deploy must have failed")
		}

		resp := doGraphQL(t, token, `
			mutation($input: DeployBuildInput!) {
				deployBuild(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"service":     testServiceName,
				"environment": "development",
				"tag":         testBuildTag,
				"digest":      testBuildDigest,
			},
		})
		requireNoErrors(t, resp)

		deployed := extractBool(t, resp.Data, "deployBuild")
		if !deployed {
			t.Fatal("deployBuild returned false")
		}
		t.Logf("deployBuild succeeded with tag=%s", testBuildTag)
	})

	t.Run("Rollback", func(t *testing.T) {
		if testBuildTag == "" {
			t.Skip("no build tag — build or deploy must have failed")
		}

		resp := doGraphQL(t, token, `
			mutation($input: RollbackInput!) {
				rollback(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"service":     testServiceName,
				"environment": "development",
				"imageTag":    testBuildTag,
			},
		})
		requireNoErrors(t, resp)

		rolled := extractBool(t, resp.Data, "rollback")
		if !rolled {
			t.Fatal("rollback returned false")
		}
		t.Logf("rollback succeeded to tag=%s", testBuildTag)

		// kubectl: verify the deployment still exists after rollback
		assertResourceExists(t, "deployment", testServiceName, namespace("development"))
	})
}
