package dbquery

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/zeitlos/lucity/pkg/usererror"
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

	size := 0

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
			size += len(s)
			cells[i] = &s
		}

		if size > maxResultBytes {
			return nil, usererror.New("the result is larger than 8 MiB; add a LIMIT or select fewer columns")
		}

		out = append(out, cells)
	}

	if err := rows.Err(); err != nil {
		return nil, rowTooLarge(err)
	}

	return out, nil
}

func rowTooLarge(err error) error {
	var exceeded *pgproto3.ExceededMaxBodyLenErr

	if errors.As(err, &exceeded) {
		return usererror.New("a row of the result is larger than 8 MiB; select fewer or smaller columns")
	}

	return err
}
