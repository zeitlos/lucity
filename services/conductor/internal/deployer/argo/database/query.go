package database

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"

	"github.com/zeitlos/lucity/pkg/deployer"
)

// Tables returns metadata for all user tables in the database.
// Uses a single query joining information_schema with pg_stat for efficiency.
func Tables(ctx context.Context, conn *pgx.Conn) ([]*deployer.DatabaseTable, error) {
	rows, err := conn.Query(ctx, `
		SELECT
			t.table_schema,
			t.table_name,
			COALESCE(s.n_live_tup, 0) AS estimated_rows,
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' AS nullable,
			CASE WHEN kcu.column_name IS NOT NULL THEN true ELSE false END AS is_primary_key
		FROM information_schema.tables t
		JOIN information_schema.columns c
			ON c.table_schema = t.table_schema AND c.table_name = t.table_name
		LEFT JOIN pg_stat_user_tables s
			ON s.schemaname = t.table_schema AND s.relname = t.table_name
		LEFT JOIN information_schema.table_constraints tc
			ON tc.table_schema = t.table_schema
			AND tc.table_name = t.table_name
			AND tc.constraint_type = 'PRIMARY KEY'
		LEFT JOIN information_schema.key_column_usage kcu
			ON kcu.constraint_name = tc.constraint_name
			AND kcu.table_schema = tc.table_schema
			AND kcu.column_name = c.column_name
		WHERE t.table_schema NOT IN ('pg_catalog', 'information_schema')
			AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_schema, t.table_name, c.ordinal_position
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	// Group results by table.
	type tableKey struct{ schema, name string }
	tableMap := make(map[tableKey]*deployer.DatabaseTable)
	var tableOrder []tableKey

	for rows.Next() {
		var (
			schema, tableName, colName, colType string
			estimatedRows                       int64
			nullable, primaryKey                bool
		)
		if err := rows.Scan(&schema, &tableName, &estimatedRows, &colName, &colType, &nullable, &primaryKey); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		key := tableKey{schema, tableName}
		tbl, exists := tableMap[key]
		if !exists {
			tbl = &deployer.DatabaseTable{
				Name:          tableName,
				Schema:        schema,
				EstimatedRows: estimatedRows,
			}
			tableMap[key] = tbl
			tableOrder = append(tableOrder, key)
		}

		tbl.Columns = append(tbl.Columns, &deployer.DatabaseColumn{
			Name:       colName,
			Type:       colType,
			Nullable:   nullable,
			PrimaryKey: primaryKey,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate table rows: %w", err)
	}

	// Preserve order.
	result := make([]*deployer.DatabaseTable, 0, len(tableOrder))
	for _, key := range tableOrder {
		result = append(result, tableMap[key])
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Schema != result[j].Schema {
			return result[i].Schema < result[j].Schema
		}
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// TableData returns paginated rows from a specific table.
// Uses pgx.Identifier for safe identifier quoting to prevent SQL injection.
func TableData(ctx context.Context, conn *pgx.Conn, schema, table string, limit, offset int) ([]string, []*deployer.DatabaseRow, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if schema == "" {
		schema = "public"
	}

	// Safe identifier quoting.
	quotedTable := pgx.Identifier{schema, table}.Sanitize()

	// Estimated row count.
	var estimatedRows int64
	err := conn.QueryRow(ctx,
		"SELECT COALESCE(n_live_tup, 0) FROM pg_stat_user_tables WHERE schemaname = $1 AND relname = $2",
		schema, table).Scan(&estimatedRows)
	if err != nil {
		// Non-fatal: just use 0.
		estimatedRows = 0
	}

	// Query data.
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", quotedTable)
	rows, err := conn.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to query table data: %w", err)
	}
	defer rows.Close()

	columns := columnNames(rows)
	dataRows, err := collectRows(rows)
	if err != nil {
		return nil, nil, 0, err
	}

	return columns, dataRows, estimatedRows, nil
}

// Query executes arbitrary SQL and returns column names, rows, and affected row count.
func Query(ctx context.Context, conn *pgx.Conn, sql string) ([]string, []*deployer.DatabaseRow, int64, error) {
	rows, err := conn.Query(ctx, sql)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	columns := columnNames(rows)
	dataRows, err := collectRows(rows)
	if err != nil {
		return nil, nil, 0, err
	}

	affected := rows.CommandTag().RowsAffected()
	return columns, dataRows, affected, nil
}

// columnNames extracts column names from row field descriptions.
func columnNames(rows pgx.Rows) []string {
	descs := rows.FieldDescriptions()
	names := make([]string, len(descs))
	for i, d := range descs {
		names[i] = d.Name
	}
	return names
}

// collectRows reads all rows and converts values to DatabaseRow proto messages.
func collectRows(rows pgx.Rows) ([]*deployer.DatabaseRow, error) {
	var result []*deployer.DatabaseRow
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to read row values: %w", err)
		}

		cells := make([]*deployer.DatabaseCell, len(values))
		for i, v := range values {
			if v == nil {
				cells[i] = &deployer.DatabaseCell{IsNull: true}
			} else {
				cells[i] = &deployer.DatabaseCell{Value: fmt.Sprintf("%v", v)}
			}
		}
		result = append(result, &deployer.DatabaseRow{Cells: cells})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}
	return result, nil
}
