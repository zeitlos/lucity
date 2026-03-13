package tests

import (
	"encoding/json"
	"testing"
	"time"
)

func testService(t *testing.T) {
	requireProjectCreated(t)
	token := testToken(t)

	t.Run("DetectServices", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query($sourceUrl: String!) {
				detectServices(sourceUrl: $sourceUrl) {
					name
					language
					framework
					startCommand
					suggestedPort
				}
			}
		`, map[string]any{"sourceUrl": testSourceURL})

		// DetectServices clones the repo — may fail on private repos without auth.
		if len(resp.Errors) > 0 {
			t.Logf("detectServices error (may need auth for private repo): %s", resp.Errors[0].Message)
			return
		}

		var data struct {
			DetectServices []struct {
				Name          string `json:"name"`
				Language      string `json:"language"`
				Framework     string `json:"framework"`
				StartCommand  string `json:"startCommand"`
				SuggestedPort int    `json:"suggestedPort"`
			} `json:"detectServices"`
		}
		unmarshalData(t, resp, &data)

		if len(data.DetectServices) == 0 {
			t.Fatal("expected at least one detected service")
		}

		t.Logf("detected %d services:", len(data.DetectServices))
		for _, svc := range data.DetectServices {
			t.Logf("  %s (%s/%s) port=%d cmd=%q", svc.Name, svc.Language, svc.Framework, svc.SuggestedPort, svc.StartCommand)
		}
	})

	t.Run("AddService", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($input: AddServiceInput!) {
				addService(input: $input) {
					id
					name
					environment
					port
					framework
					sourceUrl
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"environment": "development",
				"name":        testServiceName,
				"port":        testServicePort,
				"framework":   "nextjs",
				"sourceUrl":   testSourceURL,
			},
		})
		requireNoErrors(t, resp)

		name := extractString(t, resp.Data, "addService", "name")
		if name != testServiceName {
			t.Fatalf("expected service name %q, got %q", testServiceName, name)
		}
		env := extractString(t, resp.Data, "addService", "environment")
		if env != "development" {
			t.Fatalf("expected environment %q, got %q", "development", env)
		}
		t.Logf("added service: %s (env=%s)", name, env)
	})

	t.Run("GetServiceViaEnvironment", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query($id: ID!) {
				project(id: $id) {
					environments {
						name
						services {
							name
							port
							framework
							sourceUrl
						}
					}
				}
			}
		`, map[string]any{"id": testProjectName})
		requireNoErrors(t, resp)

		var data struct {
			Project struct {
				Environments []struct {
					Name     string `json:"name"`
					Services []struct {
						Name      string `json:"name"`
						Port      int    `json:"port"`
						Framework string `json:"framework"`
						SourceURL string `json:"sourceUrl"`
					} `json:"services"`
				} `json:"environments"`
			} `json:"project"`
		}
		unmarshalData(t, resp, &data)

		for _, env := range data.Project.Environments {
			if env.Name == "development" {
				for _, svc := range env.Services {
					if svc.Name == testServiceName {
						if svc.Port != testServicePort {
							t.Errorf("expected port %d, got %d", testServicePort, svc.Port)
						}
						t.Logf("service in %s: %s port=%d framework=%s", env.Name, svc.Name, svc.Port, svc.Framework)
						return
					}
				}
				t.Fatalf("service %q not found in development environment", testServiceName)
			}
		}
		t.Fatal("development environment not found")
	})

	// Now that a service has been added, ArgoCD has resources to deploy.
	// The deployer creates the namespace explicitly before the ArgoCD Application.
	t.Run("WaitForNamespace", func(t *testing.T) {
		if waitForNamespaceOK(t, namespace("development"), 30*time.Second) {
			devNamespaceReady = true
			assertResourceExists(t, "application.argoproj.io", testProjectName+"-development", "lucity-system")
		} else {
			t.Fatal("namespace did not appear within 30s")
		}
	})

	t.Run("GetProject_WithService", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query($id: ID!) {
				project(id: $id) {
					environments {
						name
						services { name port framework sourceUrl }
					}
				}
			}
		`, map[string]any{"id": testProjectName})
		requireNoErrors(t, resp)

		var data struct {
			Project struct {
				Environments []struct {
					Name     string `json:"name"`
					Services []struct {
						Name string `json:"name"`
					} `json:"services"`
				} `json:"environments"`
			} `json:"project"`
		}
		unmarshalData(t, resp, &data)

		found := false
		for _, env := range data.Project.Environments {
			for _, s := range env.Services {
				if s.Name == testServiceName {
					found = true
				}
			}
		}
		if !found {
			raw, _ := json.Marshal(data.Project.Environments)
			t.Fatalf("service %q not found in any environment: %s", testServiceName, string(raw))
		}
	})
}
