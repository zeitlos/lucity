package dbquery

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
)

func (c *Client) Tables(ctx context.Context) ([]Table, error) {
	conn, err := c.connect(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, tablesQuery)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	type key struct{ schema, name string }

	byTable := map[key]*Table{}
	order := []key{}

	for rows.Next() {
		var (
			schema, name, colName, colType string
			estimated                      int64
			nullable, primaryKey           bool
		)

		if err := rows.Scan(&schema, &name, &estimated, &colName, &colType, &nullable, &primaryKey); err != nil {
			return nil, err
		}

		k := key{schema, name}
		t, ok := byTable[k]

		if !ok {
			t = &Table{Name: name, Schema: schema, EstimatedRows: estimated}
			byTable[k] = t
			order = append(order, k)
		}

		t.Columns = append(t.Columns, Column{
			Name:       colName,
			Type:       colType,
			Nullable:   nullable,
			PrimaryKey: primaryKey,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]Table, 0, len(order))

	for _, k := range order {
		result = append(result, *byTable[k])
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Schema != result[j].Schema {
			return result[i].Schema < result[j].Schema
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (c *Client) Rows(ctx context.Context, schema, table string, limit, offset int) (TableData, error) {
	if limit <= 0 {
		limit = 50
	}

	if schema == "" {
		schema = "public"
	}

	conn, err := c.connect(ctx)

	if err != nil {
		return TableData{}, err
	}

	defer conn.Close(ctx)

	var estimated int64

	err = conn.QueryRow(ctx,
		"SELECT COALESCE(n_live_tup, 0) FROM pg_stat_user_tables WHERE schemaname = $1 AND relname = $2",
		schema, table,
	).Scan(&estimated)

	if err != nil {
		estimated = 0
	}

	quoted := pgx.Identifier{schema, table}.Sanitize()
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", quoted)

	rows, err := conn.Query(ctx, query, limit, offset)

	if err != nil {
		return TableData{}, err
	}

	defer rows.Close()

	columns := columnNames(rows)
	data, err := collectRows(rows)

	if err != nil {
		return TableData{}, err
	}

	return TableData{
		Columns:            columns,
		Rows:               data,
		TotalEstimatedRows: estimated,
	}, nil
}

const tablesQuery = `
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
`
