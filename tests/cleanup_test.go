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
		requireNoErrors(t, resp)

		removed := extractBool(t, resp.Data, "removeService")
		if !removed {
			t.Fatal("removeService returned false")
		}
		t.Logf("removed service %s", testServiceName)
	})

	t.Run("DeleteProject", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($id: ID!) {
				deleteProject(id: $id)
			}
		`, map[string]any{"id": testProjectName})
		requireNoErrors(t, resp)

		deleted := extractBool(t, resp.Data, "deleteProject")
		if !deleted {
			t.Fatal("deleteProject returned false")
		}
		t.Logf("deleted project %s", testProjectName)

		// kubectl: verify namespaces are gone
		waitForNamespaceGone(t, namespace("development"), 2*time.Minute)

		// kubectl: verify ArgoCD Applications are gone
		assertResourceGone(t, "application.argoproj.io", testProjectName+"-development", "lucity-system")

		t.Log("project fully cleaned up — namespaces and ArgoCD apps removed")

		// Clear the project name so cleanup() in TestMain is a no-op
		testProjectName = ""
	})
}
