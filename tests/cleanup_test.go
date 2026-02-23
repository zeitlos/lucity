package tests

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func testCleanup(t *testing.T) {
	requireProjectCreated(t)
	token := testToken(t)

	// Save the project name before deletion clears it — verification subtests need it.
	cleanupProjectName := testProjectName

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
			t.Fatalf("removeService error: %s", resp.Errors[0].Message)
		}
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

		if len(resp.Errors) > 0 {
			t.Fatalf("deleteProject error: %s", resp.Errors[0].Message)
		}
		deleted := extractBool(t, resp.Data, "deleteProject")
		if !deleted {
			t.Fatal("deleteProject returned false")
		}
		t.Logf("deleted project %s", testProjectName)

		// Clear the project name so cleanup() in TestMain is a no-op
		testProjectName = ""
	})

	// Subsystem verification — each check is its own subtest so failures are
	// clearly attributable. All checks use tight timeouts and fail on timeout.

	t.Run("VerifySoftServeRepoGone", func(t *testing.T) {
		httpResp, err := http.Get("http://localhost:23232/" + cleanupProjectName + "-gitops.git")
		if err != nil {
			t.Fatalf("could not reach Soft-serve: %v", err)
		}
		httpResp.Body.Close()
		if httpResp.StatusCode != http.StatusNotFound {
			t.Fatalf("Soft-serve gitops repo still exists (HTTP %d)", httpResp.StatusCode)
		}
	})

	t.Run("VerifyArgoCDRepoCredGone", func(t *testing.T) {
		out, err := kubectlQuiet(t, "get", "secrets", "-n", "lucity-system",
			"-l", "argocd.argoproj.io/secret-type=repository",
			"-o", "jsonpath={.items[*].metadata.name}")
		if err != nil {
			t.Fatalf("could not check ArgoCD repo secrets: %v", err)
		}
		if strings.Contains(out, cleanupProjectName) {
			t.Fatalf("ArgoCD repo credential secret still exists: %s", out)
		}
	})

	t.Run("VerifyArgoCDAppsGone", func(t *testing.T) {
		deadline := time.Now().Add(90 * time.Second)
		for _, env := range []string{"development", "staging"} {
			appName := cleanupProjectName + "-" + env
			for {
				if _, err := kubectlQuiet(t, "get", "application.argoproj.io", appName, "-n", "lucity-system"); err != nil {
					t.Logf("ArgoCD application %s removed", appName)
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("ArgoCD application %s still exists after 90s", appName)
				}
				time.Sleep(3 * time.Second)
			}
		}
	})

	t.Run("VerifyNamespacesGone", func(t *testing.T) {
		if !devNamespaceReady {
			t.Skip("namespace was never ready")
		}
		for _, env := range []string{"development", "staging"} {
			ns := cleanupProjectName + "-" + env
			if !waitForNamespaceGoneOK(t, ns, 90*time.Second) {
				t.Fatalf("namespace %s still exists after 90s", ns)
			}
			t.Logf("namespace %s removed", ns)
		}
	})

	t.Run("VerifyZotImagesGone", func(t *testing.T) {
		httpResp, err := http.Get("http://localhost:5000/v2/" + cleanupProjectName + "/" + testServiceName + "/tags/list")
		if err != nil {
			t.Fatalf("could not reach Zot registry: %v", err)
		}
		body, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()

		switch httpResp.StatusCode {
		case http.StatusNotFound:
			// Repo gone entirely
		case http.StatusOK:
			// Repo metadata may linger in Zot after manifests are deleted.
			// Accept if tags list is empty.
			if !strings.Contains(string(body), `"tags":[]`) && !strings.Contains(string(body), `"tags":null`) {
				t.Fatalf("Zot registry still has tagged images: %s", string(body))
			}
		default:
			t.Fatalf("Zot registry unexpected status %d: %s", httpResp.StatusCode, string(body))
		}
	})
}
