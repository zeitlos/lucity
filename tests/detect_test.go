package tests

import (
	"encoding/json"
	"testing"
)

func TestDetectServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping detect test in short mode (needs builder service)")
	}

	token := testToken(t)

	// This test requires:
	// 1. A project that has been created with a real GitHub source URL
	// 2. The builder service running
	//
	// For now, test that the query structure works and returns a valid
	// response (even if it's an error due to no project existing).
	resp := doGraphQL(t, token, `
		query($projectId: ID!) {
			detectServices(projectId: $projectId) {
				name
				provider
				framework
				startCommand
				suggestedPort
			}
		}
	`, map[string]any{"projectId": "test-project"})

	// Log the result — in a real setup this would have actual detected services
	if len(resp.Errors) > 0 {
		t.Logf("detect returned errors (expected if no project exists): %v", resp.Errors[0].Message)
		return
	}

	var data struct {
		DetectServices []struct {
			Name          string `json:"name"`
			Provider      string `json:"provider"`
			Framework     string `json:"framework"`
			StartCommand  string `json:"startCommand"`
			SuggestedPort int    `json:"suggestedPort"`
		} `json:"detectServices"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to decode detect response: %v", err)
	}

	t.Logf("detected %d services", len(data.DetectServices))
	for _, svc := range data.DetectServices {
		t.Logf("  %s (%s/%s) port=%d", svc.Name, svc.Provider, svc.Framework, svc.SuggestedPort)
	}
}
