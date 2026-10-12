package dbquery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	maxStatements      = 25
	maxStatementLength = 4000
)

type StatementStatistics struct {
	Enabled    bool
	Since      *time.Time
	Statements []Statement
}

type Statement struct {
	Query             string
	Calls             int64
	Rows              int64
	TotalMilliseconds float64
	MeanMilliseconds  float64
	Share             float64
}

const statementsExtensionQuery = `
SELECT n.nspname
FROM pg_catalog.pg_extension e
JOIN pg_catalog.pg_namespace n ON n.oid = e.extnamespace
WHERE e.extname = 'pg_stat_statements'`

const statementsQuery = `
SELECT pg_catalog.left(query, $1)
  , calls
  , rows
  , total_exec_time
  , mean_exec_time
  , COALESCE(total_exec_time / NULLIF(pg_catalog.sum(total_exec_time) OVER (), 0), 0)
FROM %s
WHERE dbid = (SELECT oid FROM pg_catalog.pg_database WHERE datname = pg_catalog.current_database())
  AND userid = (SELECT oid FROM pg_catalog.pg_roles WHERE rolname = current_user)
  AND toplevel
  AND query NOT LIKE '%%pg_stat_statements%%'
ORDER BY total_exec_time DESC
LIMIT $2`

func (c *Client) Statements(ctx context.Context) (StatementStatistics, error) {
	conn, err := c.connect(ctx)

	if err != nil {
		return StatementStatistics{}, err
	}

	defer conn.Close(ctx)

	var schema string

	err = conn.QueryRow(ctx, statementsExtensionQuery).Scan(&schema)

	if errors.Is(err, pgx.ErrNoRows) {
		return StatementStatistics{}, nil
	}

	if err != nil {
		return StatementStatistics{}, fmt.Errorf("find extension: %w", err)
	}

	result := StatementStatistics{Enabled: true}

	info := pgx.Identifier{schema, "pg_stat_statements_info"}.Sanitize()

	if err := conn.QueryRow(ctx, "SELECT stats_reset FROM "+info).Scan(&result.Since); err != nil {
		return StatementStatistics{}, fmt.Errorf("read statistics reset: %w", err)
	}

	view := pgx.Identifier{schema, "pg_stat_statements"}.Sanitize()

	rows, err := conn.Query(ctx, fmt.Sprintf(statementsQuery, view), maxStatementLength, maxStatements)

	if err != nil {
		return StatementStatistics{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var s Statement

		if err := rows.Scan(&s.Query, &s.Calls, &s.Rows, &s.TotalMilliseconds, &s.MeanMilliseconds, &s.Share); err != nil {
			return StatementStatistics{}, err
		}

		result.Statements = append(result.Statements, s)
	}

	if err := rows.Err(); err != nil {
		return StatementStatistics{}, err
	}

	return result, nil
}
