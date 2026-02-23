package tests

import (
	"net/http"
	"testing"
)

func testEject(t *testing.T) {
	requireProjectCreated(t)
	token := testToken(t)

	t.Run("EjectProject", func(t *testing.T) {
		body, status := doHTTP(t, "GET", gatewayURL()+"/api/eject/"+testProjectName, token)

		if status != http.StatusOK {
			t.Fatalf("eject returned status %d: %s", status, string(body))
		}

		if len(body) == 0 {
			t.Fatal("eject returned empty body")
		}

		// Verify it's a zip archive (zip magic bytes: PK\x03\x04)
		if len(body) < 4 || body[0] != 'P' || body[1] != 'K' {
			t.Fatalf("eject response does not look like a zip archive (first bytes: %x)", body[:4])
		}

		t.Logf("eject returned %d bytes (valid zip archive)", len(body))
	})
}
