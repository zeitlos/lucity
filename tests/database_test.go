package tests

import (
	"encoding/json"
	"testing"
	"time"
)

func testDatabase(t *testing.T) {
	requireProjectCreated(t)
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

		// kubectl: verify CNPG Cluster CRD exists
		assertResourceExists(t, "cluster.postgresql.cnpg.io", testDBName, namespace("development"))
	})

	t.Run("WaitForReady", func(t *testing.T) {
		// CNPG clusters can take a while to provision
		waitForPod(t, namespace("development"), "cnpg.io/cluster="+testDBName, 3*time.Minute)
		t.Log("database pod is running")
	})

	t.Run("ConnectDatabase", func(t *testing.T) {
		resp := doGraphQL(t, token, `
			mutation($projectId: ID!, $environment: String!, $database: String!) {
				connectDatabase(projectId: $projectId, environment: $environment, database: $database)
			}
		`, map[string]any{
			"projectId":   testProjectName,
			"environment": "development",
			"database":    testDBName,
		})
		requireNoErrors(t, resp)

		connected := extractBool(t, resp.Data, "connectDatabase")
		if !connected {
			t.Fatal("connectDatabase returned false")
		}

		// Verify DATABASE_URL was created as a shared variable
		varsResp := doGraphQL(t, token, `
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
		requireNoErrors(t, varsResp)

		var data struct {
			SharedVariables []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"sharedVariables"`
		}
		unmarshalData(t, varsResp, &data)

		hasDatabaseURL := false
		for _, v := range data.SharedVariables {
			if v.Key == "DATABASE_URL" {
				hasDatabaseURL = true
				t.Logf("DATABASE_URL is set (length=%d)", len(v.Value))
				break
			}
		}
		if !hasDatabaseURL {
			t.Fatal("DATABASE_URL not found in shared variables after connectDatabase")
		}
	})

	t.Run("ExecuteQuery_CreateTable", func(t *testing.T) {
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
		resp := doGraphQL(t, token, `
			query($projectId: ID!, $environment: String!, $database: String!, $table: String!) {
				databaseTableData(projectId: $projectId, environment: $environment, database: $database, table: $table, limit: 10, offset: 0) {
					columns
					rows
					totalCount
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
				Columns    []string   `json:"columns"`
				Rows       [][]string `json:"rows"`
				TotalCount int        `json:"totalCount"`
			} `json:"databaseTableData"`
		}
		unmarshalData(t, resp, &data)

		if data.DatabaseTableData.TotalCount < 2 {
			t.Fatalf("expected at least 2 rows, got %d", data.DatabaseTableData.TotalCount)
		}
		t.Logf("table data: %d columns, %d rows, total=%d",
			len(data.DatabaseTableData.Columns),
			len(data.DatabaseTableData.Rows),
			data.DatabaseTableData.TotalCount)
	})

	t.Run("ExecuteQuery_Select", func(t *testing.T) {
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
		time.Sleep(5 * time.Second)
		assertResourceGone(t, "cluster.postgresql.cnpg.io", testDBName, namespace("development"))
		t.Log("database deleted and CNPG cluster removed")
	})
}
