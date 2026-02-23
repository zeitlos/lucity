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
	// clearly attributable and don't block other checks.

	t.Run("VerifySoftServeRepoGone", func(t *testing.T) {
		httpResp, err := http.Get("http://localhost:23232/" + cleanupProjectName + "-gitops.git")
		if err != nil {
			t.Fatalf("could not reach Soft-serve: %v", err)
		}
		httpResp.Body.Close()
		if httpResp.StatusCode != http.StatusNotFound {
			t.Errorf("Soft-serve gitops repo still exists (HTTP %d)", httpResp.StatusCode)
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
			t.Errorf("ArgoCD repo credential secret still exists: %s", out)
		}
	})

	t.Run("VerifyArgoCDAppsGone", func(t *testing.T) {
		for _, env := range []string{"development", "staging"} {
			appName := cleanupProjectName + "-" + env
			gone := false
			for range 15 {
				if _, err := kubectlQuiet(t, "get", "application.argoproj.io", appName, "-n", "lucity-system"); err != nil {
					gone = true
					break
				}
				time.Sleep(2 * time.Second)
			}
			if !gone {
				t.Errorf("ArgoCD application %s still exists after 30s", appName)
			}
		}
	})

	t.Run("VerifyNamespacesGone", func(t *testing.T) {
		if !devNamespaceReady {
			t.Skip("namespace was never ready")
		}
		for _, env := range []string{"development", "staging"} {
			ns := cleanupProjectName + "-" + env
			if !waitForNamespaceGoneOK(t, ns, 2*time.Minute) {
				t.Errorf("namespace %s still exists after 2 minutes", ns)
			}
		}
	})

	t.Run("VerifyZotImagesGone", func(t *testing.T) {
		httpResp, err := http.Get("http://localhost:5000/v2/" + cleanupProjectName + "/" + testServiceName + "/tags/list")
		if err != nil {
			t.Fatalf("could not reach Zot registry: %v", err)
		}
		body, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		if httpResp.StatusCode == http.StatusOK {
			t.Errorf("Zot registry still has images: %s", string(body))
		}
		// 404 = good (no images), anything else = unexpected
		if httpResp.StatusCode != http.StatusNotFound && httpResp.StatusCode != http.StatusOK {
			t.Errorf("Zot registry unexpected status %d", httpResp.StatusCode)
		}
	})
}
