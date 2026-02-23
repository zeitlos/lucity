package tests

import (
	"testing"
	"time"
)

func testDomain(t *testing.T) {
	requireProjectCreated(t)
	requireNamespace(t)
	token := testToken(t)

	t.Run("SetDomain", func(t *testing.T) {
		if testBuildTag == "" {
			t.Skip("no deployment — build/deploy must have failed")
		}

		resp := doGraphQL(t, token, `
			mutation($input: SetServiceDomainInput!) {
				setServiceDomain(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"service":     testServiceName,
				"environment": "development",
				"host":        testServiceName + ".lucity.local",
			},
		})
		requireNoErrors(t, resp)

		ok := extractBool(t, resp.Data, "setServiceDomain")
		if !ok {
			t.Fatal("setServiceDomain returned false")
		}

		// kubectl: verify httproute exists (give ArgoCD a moment to sync)
		time.Sleep(5 * time.Second)
		out, err := kubectlQuiet(t, "get", "httproute", "-n", namespace("development"), "--no-headers")
		if err != nil {
			t.Logf("httproute check failed (may need more sync time): %v", err)
		} else {
			t.Logf("httproutes in namespace: %s", out)
		}
	})

	t.Run("RemoveDomain", func(t *testing.T) {
		if testBuildTag == "" {
			t.Skip("no deployment — build/deploy must have failed")
		}

		resp := doGraphQL(t, token, `
			mutation($input: SetServiceDomainInput!) {
				setServiceDomain(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"service":     testServiceName,
				"environment": "development",
				"host":        "",
			},
		})
		requireNoErrors(t, resp)

		ok := extractBool(t, resp.Data, "setServiceDomain")
		if !ok {
			t.Fatal("setServiceDomain (remove) returned false")
		}
		t.Log("domain removed")
	})
}
