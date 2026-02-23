package tests

import (
	"net/http"
	"testing"
)

func testHealth(t *testing.T) {
	t.Run("GatewayHealth", func(t *testing.T) {
		resp, err := http.Get(gatewayURL() + "/health")
		if err != nil {
			t.Fatalf("health check failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Playground", func(t *testing.T) {
		resp, err := http.Get(gatewayURL() + "/playground")
		if err != nil {
			t.Fatalf("playground request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}
