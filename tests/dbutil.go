package tests

import (
	"os/exec"
	"testing"
)

// psqlDirect runs a SQL query via the psql CLI and returns the output.
// Use this for direct verification when you have a connection string.
// For most tests, prefer the GraphQL executeQuery mutation instead.
func psqlDirect(t *testing.T, connString, query string) string {
	t.Helper()
	cmd := exec.Command("psql", connString, "-c", query, "--no-psqlrc", "-t", "-A")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("psql failed: %v\noutput: %s", err, string(out))
	}
	return string(out)
}
