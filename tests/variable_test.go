package tests

import (
	"encoding/json"
	"testing"
)

func testVariables(t *testing.T) {
	requireProjectCreated(t)
	token := testToken(t)

	t.Run("SetSharedVariables", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $environment: String!, $variables: [VariableInput!]!) {
				setSharedVariables(projectId: $projectId, environment: $environment, variables: $variables)
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
			"variables": []map[string]any{
				{"key": "APP_ENV", "value": "test"},
				{"key": "LOG_LEVEL", "value": "debug"},
			},
		})
		requireNoErrors(t, resp)

		ok := extractBool(t, resp.Data, "setSharedVariables")
		if !ok {
			t.Fatal("setSharedVariables returned false")
		}
		t.Log("shared variables set: APP_ENV=test, LOG_LEVEL=debug")
	})

	t.Run("GetSharedVariables", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			query($projectId: ID!, $environment: String!) {
				sharedVariables(projectId: $projectId, environment: $environment) {
					key
					value
				}
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
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
			t.Errorf("expected APP_ENV=test, got %q", vars["APP_ENV"])
		}
		if vars["LOG_LEVEL"] != "debug" {
			t.Errorf("expected LOG_LEVEL=debug, got %q", vars["LOG_LEVEL"])
		}
		t.Logf("shared variables: %v", vars)
	})

	t.Run("SetServiceVariables", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $environment: String!, $service: String!, $variables: [ServiceVariableInput!]!) {
				setServiceVariables(projectId: $projectId, environment: $environment, service: $service, variables: $variables)
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
			"service":     testServiceName,
			"variables": []map[string]any{
				{"key": "PORT", "value": "3000"},
				{"key": "APP_ENV", "fromShared": true},
			},
		})
		requireNoErrors(t, resp)

		ok := extractBool(t, resp.Data, "setServiceVariables")
		if !ok {
			t.Fatal("setServiceVariables returned false")
		}
		t.Log("service variables set: PORT=3000, APP_ENV (from shared)")
	})

	t.Run("GetServiceVariables", func(t *testing.T) {
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
			"environment": "development",
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
			t.Logf("  %s=%s (fromShared=%v)", v.Key, v.Value, v.FromShared)
			if v.Key == "PORT" && v.Value == "3000" {
				foundPort = true
			}
			if v.Key == "APP_ENV" && v.FromShared {
				foundAppEnvShared = true
			}
		}

		if !foundPort {
			t.Error("PORT=3000 not found in service variables")
		}
		if !foundAppEnvShared {
			t.Error("APP_ENV with fromShared=true not found in service variables")
		}
	})

	t.Run("OverwriteSharedVariables", func(t *testing.T) {
		// Verify idempotency: set same variables again
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $environment: String!, $variables: [VariableInput!]!) {
				setSharedVariables(projectId: $projectId, environment: $environment, variables: $variables)
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
			"variables": []map[string]any{
				{"key": "APP_ENV", "value": "staging"},
				{"key": "LOG_LEVEL", "value": "info"},
				{"key": "NEW_VAR", "value": "hello"},
			},
		})
		requireNoErrors(t, resp)

		// Verify the new values
		resp2 := doGraphQL(t, token, `
			query($projectId: ID!, $environment: String!) {
				sharedVariables(projectId: $projectId, environment: $environment) {
					key
					value
				}
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
		})
		requireNoErrors(t, resp2)

		var data struct {
			SharedVariables []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"sharedVariables"`
		}
		unmarshalData(t, resp2, &data)

		vars := make(map[string]string)
		for _, v := range data.SharedVariables {
			vars[v.Key] = v.Value
		}
		if vars["APP_ENV"] != "staging" {
			t.Errorf("expected APP_ENV=staging after overwrite, got %q", vars["APP_ENV"])
		}
		if vars["NEW_VAR"] != "hello" {
			t.Errorf("expected NEW_VAR=hello, got %q", vars["NEW_VAR"])
		}

		// Reset back to original values for subsequent tests
		doGraphQL(t, token, `
			mutation($projectId: ID!, $environment: String!, $variables: [VariableInput!]!) {
				setSharedVariables(projectId: $projectId, environment: $environment, variables: $variables)
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
			"variables": []map[string]any{
				{"key": "APP_ENV", "value": "test"},
				{"key": "LOG_LEVEL", "value": "debug"},
			},
		})
	})
}

// unmarshalServiceVars is a helper for decoding service variable responses.
func unmarshalServiceVars(t *testing.T, data json.RawMessage) map[string]string {
	t.Helper()
	var d struct {
		ServiceVariables []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"serviceVariables"`
	}
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatalf("failed to decode service variables: %v", err)
	}
	vars := make(map[string]string)
	for _, v := range d.ServiceVariables {
		vars[v.Key] = v.Value
	}
	return vars
}
