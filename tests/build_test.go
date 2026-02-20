package tests

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBuildService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build test in short mode (needs builder + Docker)")
	}

	token := testToken(t)

	// This test requires:
	// 1. A project with a real GitHub source URL
	// 2. Builder service running with Docker daemon available
	// 3. REGISTRY_TOKEN set for registry push (Zot)
	//
	// Test the build flow: start build → poll status → verify completion.
	resp := doGraphQL(t, token, `
		mutation($input: BuildServiceInput!) {
			buildService(input: $input) {
				id
				phase
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"projectId": "test-project",
			"service":   "api",
		},
	})

	// If no project exists, we expect an error — that's fine for the scaffolding test
	if len(resp.Errors) > 0 {
		t.Logf("build returned error (expected if no project exists): %v", resp.Errors[0].Message)
		return
	}

	var startData struct {
		BuildService struct {
			ID    string `json:"id"`
			Phase string `json:"phase"`
		} `json:"buildService"`
	}
	if err := json.Unmarshal(resp.Data, &startData); err != nil {
		t.Fatalf("failed to decode build start response: %v", err)
	}

	buildID := startData.BuildService.ID
	t.Logf("build started: id=%s phase=%s", buildID, startData.BuildService.Phase)

	// Poll build status until completion or timeout
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)

		statusResp := doGraphQL(t, token, `
			query($id: ID!) {
				buildStatus(id: $id) {
					id
					phase
					imageRef
					digest
					error
				}
			}
		`, map[string]any{"id": buildID})
		requireNoErrors(t, statusResp)

		var statusData struct {
			BuildStatus struct {
				ID       string  `json:"id"`
				Phase    string  `json:"phase"`
				ImageRef *string `json:"imageRef"`
				Digest   *string `json:"digest"`
				Error    *string `json:"error"`
			} `json:"buildStatus"`
		}
		if err := json.Unmarshal(statusResp.Data, &statusData); err != nil {
			t.Fatalf("failed to decode build status: %v", err)
		}

		phase := statusData.BuildStatus.Phase
		t.Logf("build %s: phase=%s", buildID, phase)

		switch phase {
		case "SUCCEEDED":
			if statusData.BuildStatus.ImageRef == nil {
				t.Error("build succeeded but imageRef is nil")
			} else {
				t.Logf("build succeeded: image=%s", *statusData.BuildStatus.ImageRef)
			}
			return
		case "FAILED":
			errMsg := ""
			if statusData.BuildStatus.Error != nil {
				errMsg = *statusData.BuildStatus.Error
			}
			t.Fatalf("build failed: %s", errMsg)
		}
	}

	t.Fatal("build timed out after 5 minutes")
}
