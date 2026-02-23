package tests

import (
	"math/rand"
	"os"
	"testing"
)

// Shared state across the integration test suite.
// Set during test phases and read by subsequent tests.
var (
	testProjectName string
	testServiceName = "vouch"
	testSourceURL   = "https://github.com/zeitlos/vouch"
	testServicePort = 3000
	testDBName      = "main"
	testBuildTag    string // set after successful build
	testBuildDigest string // set after successful build
)

func TestMain(m *testing.M) {
	testProjectName = "inttest-" + randomSuffix(6)
	code := m.Run()
	cleanup()
	os.Exit(code)
}

// TestIntegration is the single orchestrator that runs all integration tests
// sequentially. Tests share state via package-level variables — earlier tests
// create resources that later tests depend on.
func TestIntegration(t *testing.T) {
	// Phase 1: Gateway-only (always run)
	t.Run("Health", testHealth)
	t.Run("Auth", testAuth)

	if testing.Short() {
		t.Log("short mode: skipping full integration tests")
		return
	}

	// Phase 2: Full infrastructure required
	t.Run("Project", testProject)
	t.Run("Environment", testEnvironment)
	t.Run("Service", testService)
	t.Run("Variables", testVariables)
	t.Run("Database", testDatabase)
	t.Run("Build", testBuild)
	t.Run("Deploy", testDeploy)
	t.Run("Domain", testDomain)
	t.Run("Promote", testPromote)
	t.Run("Eject", testEject)
	t.Run("GitHub", testGitHub)
	t.Run("Cleanup", testCleanup)
}

// cleanup deletes the test project (best-effort) and verifies namespaces are gone.
func cleanup() {
	if testProjectName == "" {
		return
	}

	// Best-effort: delete via GraphQL
	token, err := makeToken()
	if err != nil {
		return
	}

	doGraphQLRaw(token, `
		mutation($id: ID!) {
			deleteProject(id: $id)
		}
	`, map[string]any{"id": testProjectName})

	// Best-effort: kubectl cleanup
	kubectlRaw("delete", "namespace", testProjectName+"-development", "--ignore-not-found", "--wait=false")
	kubectlRaw("delete", "namespace", testProjectName+"-staging", "--ignore-not-found", "--wait=false")
}

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomSuffix(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
