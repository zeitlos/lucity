package dbquery

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Client struct {
	dsn string
}

func New(dsn string) *Client {
	return &Client{dsn: dsn}
}

const (
	statementTimeout = "30s"
	maxQueryRows     = 1000
)

func (c *Client) connect(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, c.dsn)

	if err != nil {
		return nil, err
	}

	if _, err := conn.Exec(ctx, "SET statement_timeout = '"+statementTimeout+"'"); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("set statement_timeout: %w", err)
	}

	return conn, nil
}

type Table struct {
	Name          string
	Schema        string
	EstimatedRows int64
	Columns       []Column
}

type Column struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
}

type TableData struct {
	Columns            []string
	Rows               [][]*string
	TotalEstimatedRows int64
}

type Result struct {
	Columns      []string
	Rows         [][]*string
	AffectedRows int64
}
