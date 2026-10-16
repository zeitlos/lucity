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
		n.nspname,
		c.relname,
		COALESCE(s.n_live_tup, 0) AS estimated_rows,
		a.attname,
		format_type(a.atttypid, a.atttypmod),
		NOT a.attnotnull AS nullable,
		COALESCE(a.attnum = ANY(pk.conkey), false) AS is_primary_key
	FROM pg_class c
	JOIN pg_namespace n
		ON n.oid = c.relnamespace
	JOIN pg_attribute a
		ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped
	LEFT JOIN pg_stat_user_tables s
		ON s.relid = c.oid
	LEFT JOIN pg_constraint pk
		ON pk.conrelid = c.oid AND pk.contype = 'p'
	WHERE c.relkind IN ('r', 'p')
		AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		AND n.nspname NOT LIKE 'pg\_%'
		AND has_table_privilege(c.oid, 'SELECT')
	ORDER BY n.nspname, c.relname, a.attnum
`
