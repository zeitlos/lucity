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

	// Create staging from development. By this point, dev has:
	//   - Service: vouch (deployed with testBuildTag)
	//   - Shared vars: APP_ENV=test, LOG_LEVEL=debug
	//   - Service vars: PORT=3000, APP_ENV (fromShared)
	t.Run("CreateStaging", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($input: CreateEnvironmentInput!) {
				createEnvironment(input: $input) {
					id
					name
					namespace
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
			t.Logf("createEnvironment error (staging may already exist): %s", resp.Errors[0].Message)
		} else {
			name := extractString(t, resp.Data, "createEnvironment", "name")
			if name != "staging" {
				t.Fatalf("expected 'staging', got %q", name)
			}
		}

		waitForNamespace(t, namespace("staging"), 30*time.Second)
		assertResourceExists(t, "application.argoproj.io", testProjectName+"-staging", "lucity-system")
	})

	// Verify staging inherited the service definition from dev.
	t.Run("VerifyServiceInherited", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query($id: ID!) {
				project(id: $id) {
					environments {
						name
						services {
							name
							imageTag
						}
					}
					services {
						name
						port
						framework
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
						Name     string `json:"name"`
						ImageTag string `json:"imageTag"`
					} `json:"services"`
				} `json:"environments"`
				Services []struct {
					Name      string `json:"name"`
					Port      int    `json:"port"`
					Framework string `json:"framework"`
				} `json:"services"`
			} `json:"project"`
		}
		unmarshalData(t, resp, &data)

		// Verify the project-level service definition
		foundService := false
		for _, svc := range data.Project.Services {
			if svc.Name == testServiceName {
				foundService = true
				if svc.Port != testServicePort {
					t.Errorf("expected service port %d, got %d", testServicePort, svc.Port)
				}
				t.Logf("project service: %s port=%d framework=%s", svc.Name, svc.Port, svc.Framework)
			}
		}
		if !foundService {
			t.Fatalf("service %s not found in project", testServiceName)
		}

		// Verify staging has the service instance
		for _, env := range data.Project.Environments {
			if env.Name == "staging" {
				found := false
				for _, svc := range env.Services {
					if svc.Name == testServiceName {
						found = true
						t.Logf("staging has service %s (imageTag=%s)", svc.Name, svc.ImageTag)
					}
				}
				if !found {
					t.Errorf("staging environment does not have service %s", testServiceName)
				}
			}
		}
	})

	// Verify staging inherited shared variables from dev.
	t.Run("VerifySharedVarsInherited", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query($projectId: ID!, $environment: String!) {
				sharedVariables(projectId: $projectId, environment: $environment) {
					key
					value
				}
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "staging",
		})
		requireNoErrors(t, resp)

		var data struct {
			SharedVariables []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"sharedVariables"`
		}
		unmarshalData(t, resp, &data)

		vars := make(map[string]string)
		for _, v := range data.SharedVariables {
			vars[v.Key] = v.Value
		}

		if vars["APP_ENV"] != "test" {
			t.Errorf("expected staging APP_ENV=test, got %q", vars["APP_ENV"])
		}
		if vars["LOG_LEVEL"] != "debug" {
			t.Errorf("expected staging LOG_LEVEL=debug, got %q", vars["LOG_LEVEL"])
		}
		t.Logf("staging shared vars: %v", vars)
	})

	// Verify staging inherited service variables from dev.
	t.Run("VerifyServiceVarsInherited", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query($projectId: ID!, $environment: String!, $service: String!) {
				serviceVariables(projectId: $projectId, environment: $environment, service: $service) {
					key
					value
					fromShared
				}
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "staging",
			"service":     testServiceName,
		})
		requireNoErrors(t, resp)

		var data struct {
			ServiceVariables []struct {
				Key        string `json:"key"`
				Value      string `json:"value"`
				FromShared bool   `json:"fromShared"`
			} `json:"serviceVariables"`
		}
		unmarshalData(t, resp, &data)

		foundPort := false
		foundAppEnvShared := false
		for _, v := range data.ServiceVariables {
			t.Logf("staging service var: %s=%s (fromShared=%v)", v.Key, v.Value, v.FromShared)
			if v.Key == "PORT" && v.Value == "3000" && !v.FromShared {
				foundPort = true
			}
			if v.Key == "APP_ENV" && v.FromShared {
				foundAppEnvShared = true
			}
		}

		if !foundPort {
			t.Error("PORT=3000 not found in staging service variables")
		}
		if !foundAppEnvShared {
			t.Error("APP_ENV with fromShared=true not found in staging service variables")
		}
	})

	// Promote the deployed image tag from dev to staging.
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
		if imageTag != testBuildTag {
			t.Fatalf("expected promoted tag %s, got %s", testBuildTag, imageTag)
		}
		t.Logf("promoted %s from development to staging: tag=%s", testServiceName, imageTag)
	})

	// Verify the promoted tag is reflected in the staging environment query.
	t.Run("VerifyPromotedTag", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query($id: ID!) {
				project(id: $id) {
					environments {
						name
						services {
							name
							imageTag
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
						Name     string `json:"name"`
						ImageTag string `json:"imageTag"`
					} `json:"services"`
				} `json:"environments"`
			} `json:"project"`
		}
		unmarshalData(t, resp, &data)

		for _, env := range data.Project.Environments {
			if env.Name == "staging" {
				for _, svc := range env.Services {
					if svc.Name == testServiceName {
						if svc.ImageTag != testBuildTag {
							t.Fatalf("staging imageTag=%s, expected %s", svc.ImageTag, testBuildTag)
						}
						t.Logf("staging %s has correct imageTag=%s", svc.Name, svc.ImageTag)
						return
					}
				}
				t.Fatal("service not found in staging after promote")
			}
		}
		t.Fatal("staging environment not found in project")
	})

	// Wait for the promoted service to be running in staging and verify it responds.
	t.Run("VerifyStagingRunning", func(t *testing.T) {
		ns := namespace("staging")

		waitForPod(t, ns, "app.kubernetes.io/name="+testServiceName, 60*time.Second)

		svc := k8sServiceName(testProjectName, testServiceName)
		cmd := portForward(t, ns, svc, 18081, testServicePort)
		defer stopPortForward(t, cmd)

		waitForHTTP(t, "http://localhost:18081", 15*time.Second)
		t.Log("staging service is responding via port-forward")
	})

	// Clean up staging.
	t.Run("DeleteStaging", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $environment: String!) {
				deleteEnvironment(projectId: $projectId, environment: $environment)
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "staging",
		})
		requireNoErrors(t, resp)

		if waitForNamespaceGoneOK(t, namespace("staging"), 60*time.Second) {
			t.Log("staging environment cleaned up")
		} else {
			t.Log("staging namespace still exists (will be cleaned up with project deletion)")
		}
	})
}
