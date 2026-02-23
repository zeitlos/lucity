package tests

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// dbReady is set to true once the CNPG cluster pod is running.
// Tests that need a connectable database check this before proceeding.
var dbReady bool

// dbPortForward holds the kubectl port-forward process for the CNPG service.
// Started after the database pod is ready, killed after DeleteDatabase.
var dbPortForward *exec.Cmd

func requireDBReady(t *testing.T) {
	t.Helper()
	if !dbReady {
		t.Skip("skipping: database pod not ready (CNPG may still be provisioning)")
	}
}

func testDatabase(t *testing.T) {
	requireProjectCreated(t)
	requireNamespace(t)
	token := testToken(t)

	t.Run("CreateDatabase", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($input: CreateDatabaseInput!) {
				createDatabase(input: $input) {
					name
					version
					instances
					size
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId": testProjectName,
				"name":      testDBName,
				"version":   "16",
				"instances": 1,
				"size":      "1Gi",
			},
		})
		requireNoErrors(t, resp)

		name := extractString(t, resp.Data, "createDatabase", "name")
		if name != testDBName {
			t.Fatalf("expected database name %q, got %q", testDBName, name)
		}
		t.Logf("created database: %s", name)

		// kubectl: CNPG Cluster CRD may not appear immediately
		// Helm generates the CNPG cluster name as: {namespace}-lucity-app-pg-{dbname}
		cnpgName := namespace("development") + "-lucity-app-pg-" + testDBName
		out, err := kubectlQuiet(t, "get", "cluster.postgresql.cnpg.io", cnpgName, "-n", namespace("development"))
		if err != nil {
			t.Logf("CNPG cluster %s not yet visible via kubectl", cnpgName)
		} else {
			t.Logf("CNPG cluster exists: %s", out)
		}
	})

	t.Run("WaitForReady", func(t *testing.T) {
		// Wait for the CNPG cluster's Ready condition, not just pod Running.
		// Pod Running doesn't mean PostgreSQL has finished initialization with credentials.
		cnpgClusterName := namespace("development") + "-lucity-app-pg-" + testDBName
		t.Logf("waiting for CNPG cluster %s to be Ready", cnpgClusterName)

		deadline := time.Now().Add(60 * time.Second)
		for time.Now().Before(deadline) {
			out, err := kubectlQuiet(t,
				"get", "cluster.postgresql.cnpg.io", cnpgClusterName,
				"-n", namespace("development"),
				"-o", "jsonpath={.status.conditions[?(@.type==\"Ready\")].status}",
			)
			if err == nil && strings.TrimSpace(out) == "True" {
				dbReady = true
				t.Log("CNPG cluster is Ready")
				return
			}
			time.Sleep(3 * time.Second)
		}
		t.Fatal("CNPG cluster did not become Ready within 60s")
	})

	t.Run("PortForward", func(t *testing.T) {
		requireDBReady(t)

		ns := namespace("development")
		svc := ns + "-lucity-app-pg-" + testDBName + "-rw"
		dbPortForward = portForward(t, ns, svc, 5432, 5432)
	})

	t.Run("DatabaseRef", func(t *testing.T) {
		// Add a databaseRef to existing service variables.
		// setServiceVariables is a full replacement, so we must include existing
		// variables (PORT, APP_ENV from shared) to avoid wiping them.
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
				{
					"key": "DATABASE_URL",
					"databaseRef": map[string]any{
						"database": testDBName,
						"key":      "uri",
					},
				},
			},
		})
		requireNoErrors(t, resp)

		set := extractBool(t, resp.Data, "setServiceVariables")
		if !set {
			t.Fatal("setServiceVariables returned false")
		}

		// Verify the variable was created with the databaseRef
		varsResp := doGraphQL(t, token, `
			query($projectId: ID!, $environment: String!, $service: String!) {
				serviceVariables(projectId: $projectId, environment: $environment, service: $service) {
					key
					databaseRef { database key }
				}
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
			"service":     testServiceName,
		})
		requireNoErrors(t, varsResp)

		var data struct {
			ServiceVariables []struct {
				Key         string `json:"key"`
				DatabaseRef *struct {
					Database string `json:"database"`
					Key      string `json:"key"`
				} `json:"databaseRef"`
			} `json:"serviceVariables"`
		}
		unmarshalData(t, varsResp, &data)

		found := false
		for _, v := range data.ServiceVariables {
			if v.Key == "DATABASE_URL" && v.DatabaseRef != nil {
				if v.DatabaseRef.Database != testDBName {
					t.Errorf("expected databaseRef.database=%q, got %q", testDBName, v.DatabaseRef.Database)
				}
				if v.DatabaseRef.Key != "uri" {
					t.Errorf("expected databaseRef.key=%q, got %q", "uri", v.DatabaseRef.Key)
				}
				found = true
				t.Logf("DATABASE_URL references database %s (key=%s)", v.DatabaseRef.Database, v.DatabaseRef.Key)
			}
		}
		if !found {
			t.Fatal("DATABASE_URL with databaseRef not found in service variables")
		}
	})

	t.Run("ExecuteQuery_CreateTable", func(t *testing.T) {
		requireDBReady(t)

		resp := doGraphQL(t, token, `
			mutation($input: DatabaseQueryInput!) {
				executeQuery(input: $input) {
					columns
					rows
					affectedRows
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"environment": "development",
				"database":    testDBName,
				"query":       "CREATE TABLE IF NOT EXISTS test_items (id SERIAL PRIMARY KEY, name TEXT NOT NULL, created_at TIMESTAMP DEFAULT NOW())",
			},
		})
		requireNoErrors(t, resp)
		t.Log("created test_items table")
	})

	t.Run("ExecuteQuery_Insert", func(t *testing.T) {
		requireDBReady(t)

		resp := doGraphQL(t, token, `
			mutation($input: DatabaseQueryInput!) {
				executeQuery(input: $input) {
					columns
					rows
					affectedRows
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"environment": "development",
				"database":    testDBName,
				"query":       "INSERT INTO test_items (name) VALUES ('hello'), ('world')",
			},
		})
		requireNoErrors(t, resp)

		var data struct {
			ExecuteQuery struct {
				AffectedRows int `json:"affectedRows"`
			} `json:"executeQuery"`
		}
		unmarshalData(t, resp, &data)

		if data.ExecuteQuery.AffectedRows != 2 {
			t.Fatalf("expected 2 affected rows, got %d", data.ExecuteQuery.AffectedRows)
		}
		t.Log("inserted 2 rows")
	})

	t.Run("DatabaseTables", func(t *testing.T) {
		requireDBReady(t)

		resp := doGraphQL(t, token, `
			query($projectId: ID!, $environment: String!, $database: String!) {
				databaseTables(projectId: $projectId, environment: $environment, database: $database) {
					name
					schema
					columns { name type nullable }
				}
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
			"database":    testDBName,
		})
		requireNoErrors(t, resp)

		var data struct {
			DatabaseTables []struct {
				Name    string `json:"name"`
				Schema  string `json:"schema"`
				Columns []struct {
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"databaseTables"`
		}
		unmarshalData(t, resp, &data)

		found := false
		for _, table := range data.DatabaseTables {
			if table.Name == "test_items" {
				found = true
				t.Logf("table test_items has %d columns", len(table.Columns))
			}
		}
		if !found {
			raw, _ := json.Marshal(data.DatabaseTables)
			t.Fatalf("test_items table not found in database tables: %s", string(raw))
		}
	})

	t.Run("DatabaseTableData", func(t *testing.T) {
		requireDBReady(t)

		resp := doGraphQL(t, token, `
			query($projectId: ID!, $environment: String!, $database: String!, $table: String!) {
				databaseTableData(projectId: $projectId, environment: $environment, database: $database, table: $table, limit: 10, offset: 0) {
					columns
					rows
					totalEstimatedRows
				}
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
			"database":    testDBName,
			"table":       "test_items",
		})
		requireNoErrors(t, resp)

		var data struct {
			DatabaseTableData struct {
				Columns            []string   `json:"columns"`
				Rows               [][]string `json:"rows"`
				TotalEstimatedRows int        `json:"totalEstimatedRows"`
			} `json:"databaseTableData"`
		}
		unmarshalData(t, resp, &data)

		if len(data.DatabaseTableData.Rows) < 2 {
			t.Fatalf("expected at least 2 rows, got %d", len(data.DatabaseTableData.Rows))
		}
		t.Logf("table data: %d columns, %d rows, estimated total=%d",
			len(data.DatabaseTableData.Columns),
			len(data.DatabaseTableData.Rows),
			data.DatabaseTableData.TotalEstimatedRows)
	})

	t.Run("ExecuteQuery_Select", func(t *testing.T) {
		requireDBReady(t)

		resp := doGraphQL(t, token, `
			mutation($input: DatabaseQueryInput!) {
				executeQuery(input: $input) {
					columns
					rows
					affectedRows
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"projectId":   testProjectName,
				"environment": "development",
				"database":    testDBName,
				"query":       "SELECT name FROM test_items ORDER BY id",
			},
		})
		requireNoErrors(t, resp)

		var data struct {
			ExecuteQuery struct {
				Columns []string   `json:"columns"`
				Rows    [][]string `json:"rows"`
			} `json:"executeQuery"`
		}
		unmarshalData(t, resp, &data)

		if len(data.ExecuteQuery.Rows) != 2 {
			t.Fatalf("expected 2 rows from SELECT, got %d", len(data.ExecuteQuery.Rows))
		}
		t.Logf("SELECT returned: %v", data.ExecuteQuery.Rows)
	})

	t.Run("DeleteDatabase", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $name: String!) {
				deleteDatabase(projectId: $projectId, name: $name)
			}
		`, map[string]any{
			"projectId": testProjectName,
			"name":      testDBName,
		})
		requireNoErrors(t, resp)

		deleted := extractBool(t, resp.Data, "deleteDatabase")
		if !deleted {
			t.Fatal("deleteDatabase returned false")
		}

		// kubectl: verify CNPG cluster is gone (with some delay for cleanup)
		cnpgName := namespace("development") + "-lucity-app-pg-" + testDBName
		time.Sleep(5 * time.Second)
		if _, err := kubectlQuiet(t, "get", "cluster.postgresql.cnpg.io", cnpgName, "-n", namespace("development")); err != nil {
			t.Log("database deleted and CNPG cluster removed")
		} else {
			t.Log("WARNING: CNPG cluster still exists (may take time to finalize)")
		}
	})

	t.Run("StopPortForward", func(t *testing.T) {
		stopPortForward(t, dbPortForward)
		dbPortForward = nil
	})
}

// containsRunning checks if the kubectl pod phase output contains "Running".
func containsRunning(out string) bool {
	for _, line := range splitLines(out) {
		if line == "Running" {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			if line != "" {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		line := s[start:]
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
