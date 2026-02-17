package tests

import (
	"net/http"
	"testing"
)

func TestGatewayHealth(t *testing.T) {
	resp, err := http.Get(gatewayURL() + "/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGraphQLPlayground(t *testing.T) {
	resp, err := http.Get(gatewayURL() + "/playground")
	if err != nil {
		t.Fatalf("playground request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
