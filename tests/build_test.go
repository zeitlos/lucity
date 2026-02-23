package tests

import (
	"testing"
	"time"
)

func testBuild(t *testing.T) {
	requireProjectCreated(t)
	token := testToken(t)

	var buildID string

	t.Run("StartBuild", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($input: BuildServiceInput!) {
				buildService(input: $input) {
					id
					phase
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId": testProjectName,
				"service":   testServiceName,
			},
		})
		requireNoErrors(t, resp)

		buildID = extractString(t, resp.Data, "buildService", "id")
		phase := extractString(t, resp.Data, "buildService", "phase")

		if buildID == "" {
			t.Fatal("buildService returned empty id")
		}
		t.Logf("build started: id=%s phase=%s", buildID, phase)
	})

	t.Run("PollBuildStatus", func(t *testing.T) {
		if buildID == "" {
			t.Fatal("no build ID — StartBuild must have failed")
		}

		deadline := time.Now().Add(5 * time.Minute)
		for time.Now().Before(deadline) {
			time.Sleep(3 * time.Second)

			resp := doGraphQL(t, token, `
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
			requireNoErrors(t, resp)

			var data struct {
				BuildStatus struct {
					ID       string  `json:"id"`
					Phase    string  `json:"phase"`
					ImageRef *string `json:"imageRef"`
					Digest   *string `json:"digest"`
					Error    *string `json:"error"`
				} `json:"buildStatus"`
			}
			unmarshalData(t, resp, &data)

			phase := data.BuildStatus.Phase
			t.Logf("build %s: phase=%s", buildID, phase)

			switch phase {
			case "SUCCEEDED":
				if data.BuildStatus.ImageRef == nil || *data.BuildStatus.ImageRef == "" {
					t.Fatal("build succeeded but imageRef is empty")
				}
				if data.BuildStatus.Digest != nil {
					testBuildDigest = *data.BuildStatus.Digest
				}
				// Extract tag from imageRef (format: registry/project/service:tag)
				testBuildTag = extractTagFromImageRef(*data.BuildStatus.ImageRef)
				t.Logf("build succeeded: image=%s tag=%s digest=%s", *data.BuildStatus.ImageRef, testBuildTag, testBuildDigest)
				return
			case "FAILED":
				errMsg := ""
				if data.BuildStatus.Error != nil {
					errMsg = *data.BuildStatus.Error
				}
				t.Fatalf("build failed: %s", errMsg)
			}
		}

		t.Fatal("build timed out after 5 minutes")
	})
}

// extractTagFromImageRef extracts the tag from an image reference like "registry/project/service:tag".
func extractTagFromImageRef(imageRef string) string {
	for i := len(imageRef) - 1; i >= 0; i-- {
		if imageRef[i] == ':' {
			return imageRef[i+1:]
		}
	}
	return imageRef
}
