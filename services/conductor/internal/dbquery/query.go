package dbquery

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (c *Client) Query(ctx context.Context, sql string) (Result, error) {
	conn, err := c.connect(ctx)

	if err != nil {
		return Result{}, err
	}

	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, sql)

	if err != nil {
		return Result{}, err
	}

	defer rows.Close()

	columns := columnNames(rows)
	data, err := collectRows(rows)

	if err != nil {
		return Result{}, err
	}

	return Result{
		Columns:      columns,
		Rows:         data,
		AffectedRows: rows.CommandTag().RowsAffected(),
	}, nil
}

func columnNames(rows pgx.Rows) []string {
	descs := rows.FieldDescriptions()
	names := make([]string, len(descs))

	for i, d := range descs {
		names[i] = d.Name
	}

	return names
}

func collectRows(rows pgx.Rows) ([][]*string, error) {
	var out [][]*string

	for rows.Next() {
		if len(out) >= maxQueryRows {
			break
		}

		vals, err := rows.Values()

		if err != nil {
			return nil, err
		}

		cells := make([]*string, len(vals))

		for i, v := range vals {
			if v == nil {
				cells[i] = nil
				continue
			}

			s := fmt.Sprintf("%v", v)
			cells[i] = &s
		}

		out = append(out, cells)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
