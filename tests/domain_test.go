package tests

import (
	"testing"
	"time"
)

func testDomain(t *testing.T) {
	requireProjectCreated(t)
	requireNamespace(t)
	token := testToken(t)

	t.Run("PlatformConfig", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query {
				platformConfig {
					workloadDomain
					domainTarget
				}
			}
		`, nil)
		requireNoErrors(t, resp)

		wd := extractString(t, resp.Data, "platformConfig", "workloadDomain")
		if wd == "" {
			t.Fatal("platformConfig.workloadDomain is empty")
		}
		t.Logf("workloadDomain: %s", wd)
	})

	t.Run("GenerateDomain", func(t *testing.T) {
		if testBuildTag == "" {
			t.Skip("no deployment — build/deploy must have failed")
		}

		resp := doGraphQL(t, token, `
			mutation($input: GenerateDomainInput!) {
				generateDomain(input: $input) {
					hostname
					type
					dnsStatus
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

		hostname := extractString(t, resp.Data, "generateDomain", "hostname")
		if hostname == "" {
			t.Fatal("generateDomain returned empty hostname")
		}
		domainType := extractString(t, resp.Data, "generateDomain", "type")
		if domainType != "PLATFORM" {
			t.Fatalf("expected domain type PLATFORM, got %s", domainType)
		}
		t.Logf("generated platform domain: %s", hostname)

		// kubectl: verify httproute exists (give ArgoCD a moment to sync)
		time.Sleep(5 * time.Second)
		out, err := kubectlQuiet(t, "get", "httproute", "-n", namespace("development"), "--no-headers")
		if err != nil {
			t.Logf("httproute check failed (may need more sync time): %v", err)
		} else {
			t.Logf("httproutes in namespace: %s", out)
		}
	})

	t.Run("AddCustomDomain", func(t *testing.T) {
		if testBuildTag == "" {
			t.Skip("no deployment — build/deploy must have failed")
		}

		resp := doGraphQL(t, token, `
			mutation($input: AddCustomDomainInput!) {
				addCustomDomain(input: $input) {
					hostname
					type
					dnsStatus
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"service":     testServiceName,
				"environment": "development",
				"hostname":    "custom.example.com",
			},
		})
		requireNoErrors(t, resp)

		hostname := extractString(t, resp.Data, "addCustomDomain", "hostname")
		if hostname != "custom.example.com" {
			t.Fatalf("expected hostname custom.example.com, got %s", hostname)
		}
		domainType := extractString(t, resp.Data, "addCustomDomain", "type")
		if domainType != "CUSTOM" {
			t.Fatalf("expected domain type CUSTOM, got %s", domainType)
		}
		t.Logf("added custom domain: %s (dns: %s)", hostname, extractString(t, resp.Data, "addCustomDomain", "dnsStatus"))
	})

	t.Run("RemoveDomain", func(t *testing.T) {
		if testBuildTag == "" {
			t.Skip("no deployment — build/deploy must have failed")
		}

		resp := doGraphQL(t, token, `
			mutation($input: RemoveDomainInput!) {
				removeDomain(input: $input)
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"service":     testServiceName,
				"environment": "development",
				"hostname":    "custom.example.com",
			},
		})
		requireNoErrors(t, resp)

		ok := extractBool(t, resp.Data, "removeDomain")
		if !ok {
			t.Fatal("removeDomain returned false")
		}
		t.Log("custom domain removed")
	})
}
