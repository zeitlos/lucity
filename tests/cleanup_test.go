package tests

import (
	"testing"
	"time"
)

func testCleanup(t *testing.T) {
	requireProjectCreated(t)
	token := testToken(t)

	t.Run("RemoveService", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $service: String!) {
				removeService(projectId: $projectId, service: $service)
			}
		`, map[string]any{
			"projectId": testProjectName,
			"service":   testServiceName,
		})

		if len(resp.Errors) > 0 {
			// Service removal may fail if packager is down — log and continue
			t.Logf("removeService error (non-fatal): %s", resp.Errors[0].Message)
		} else {
			removed := extractBool(t, resp.Data, "removeService")
			if !removed {
				t.Log("removeService returned false (service may already be gone)")
			} else {
				t.Logf("removed service %s", testServiceName)
			}
		}
	})

	t.Run("DeleteProject", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($id: ID!) {
				deleteProject(id: $id)
			}
		`, map[string]any{"id": testProjectName})

		if len(resp.Errors) > 0 {
			t.Logf("deleteProject error: %s", resp.Errors[0].Message)
			t.Log("project deletion failed — manual cleanup may be needed")
			t.Logf("  kubectl delete ns %s-development --ignore-not-found", testProjectName)
			t.Logf("  kubectl delete ns %s-staging --ignore-not-found", testProjectName)
			return
		}

		deleted := extractBool(t, resp.Data, "deleteProject")
		if !deleted {
			t.Fatal("deleteProject returned false")
		}
		t.Logf("deleted project %s", testProjectName)

		// kubectl: verify namespaces are gone (only if we could talk to the cluster earlier)
		if devNamespaceReady {
			waitForNamespaceGone(t, namespace("development"), 2*time.Minute)
			assertResourceGone(t, "application.argoproj.io", testProjectName+"-development", "lucity-system")
			t.Log("project fully cleaned up — namespaces and ArgoCD apps removed")
		} else {
			t.Log("skipping namespace verification (namespace was never ready)")
		}

		// Clear the project name so cleanup() in TestMain is a no-op
		testProjectName = ""
	})
}
