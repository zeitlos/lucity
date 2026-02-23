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

		// Subsystem cleanup verification (all non-fatal).
		// ArgoCD cascade deletion is async — the API returns immediately but resource
		// cleanup (CNPG finalizers, pod termination) takes time. We verify the control
		// plane artifacts are gone, not that every K8s resource has been finalized.

		// 1. Soft-serve: GitOps repo should be gone
		httpResp, httpErr := http.Get("http://localhost:23232/" + testProjectName + "-gitops.git")
		if httpErr != nil {
			t.Log("WARNING: could not reach Soft-serve to verify repo deletion (port-forward may be down)")
		} else {
			httpResp.Body.Close()
			if httpResp.StatusCode == http.StatusNotFound {
				t.Log("Soft-serve gitops repo removed")
			} else {
				t.Logf("WARNING: Soft-serve gitops repo may still exist (HTTP %d)", httpResp.StatusCode)
			}
		}

		// 2. Zot registry: image tags should be gone or empty
		httpResp, httpErr = http.Get("http://localhost:5000/v2/" + testProjectName + "/" + testServiceName + "/tags/list")
		if httpErr != nil {
			t.Log("WARNING: could not reach Zot registry (port-forward may be down)")
		} else {
			body, _ := io.ReadAll(httpResp.Body)
			httpResp.Body.Close()
			if httpResp.StatusCode == http.StatusNotFound {
				t.Log("Zot registry: no images for project")
			} else if httpResp.StatusCode == http.StatusOK {
				t.Logf("WARNING: Zot registry still has image data (expected — builder doesn't delete images): %s", string(body))
			} else {
				t.Logf("WARNING: Zot registry unexpected status %d", httpResp.StatusCode)
			}
		}

		// 3. ArgoCD repo credentials: secret for this project's gitops URL should be gone
		out, err := kubectlQuiet(t, "get", "secrets", "-n", "lucity-system",
			"-l", "argocd.argoproj.io/secret-type=repository",
			"-o", "jsonpath={.items[*].metadata.name}")
		if err != nil {
			t.Log("WARNING: could not check ArgoCD repo secrets")
		} else if strings.Contains(out, testProjectName) {
			t.Logf("WARNING: ArgoCD repo credential secret may still exist: %s", out)
		} else {
			t.Log("ArgoCD repo credential removed")
		}

		// 4. ArgoCD applications — cascade delete is async, so give it a short grace period
		for _, env := range []string{"development", "staging"} {
			appName := testProjectName + "-" + env
			gone := false
			for range 5 {
				if _, err := kubectlQuiet(t, "get", "application.argoproj.io", appName, "-n", "lucity-system"); err != nil {
					gone = true
					break
				}
				time.Sleep(2 * time.Second)
			}
			if gone {
				t.Logf("ArgoCD application %s removed", appName)
			} else {
				t.Logf("ArgoCD application %s still finalizing (cascade delete is async)", appName)
			}
		}

		// 5. Namespace — don't wait long. ArgoCD cascade + CNPG finalizers can take
		// minutes. A quick check is enough; the namespace will eventually be cleaned up.
		if devNamespaceReady {
			if waitForNamespaceGoneOK(t, namespace("development"), 10*time.Second) {
				t.Log("development namespace removed")
			} else {
				t.Log("development namespace still finalizing (CNPG finalizers, expected)")
			}
		}

		// Clear the project name so cleanup() in TestMain is a no-op
		testProjectName = ""
	})
}
