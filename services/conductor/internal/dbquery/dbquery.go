package dbquery

import (
	"context"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgproto3"
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
	maxMessageBytes  = 8 << 20
	maxResultBytes   = 8 << 20
	maxTableColumns  = 20000
)

func (c *Client) connect(ctx context.Context) (*pgx.Conn, error) {
	config, err := pgx.ParseConfig(c.dsn)

	if err != nil {
		return nil, err
	}

	config.BuildFrontend = func(r io.Reader, w io.Writer) *pgproto3.Frontend {
		frontend := pgproto3.NewFrontend(r, w)
		frontend.SetMaxBodyLen(maxMessageBytes)

		return frontend
	}

	conn, err := pgx.ConnectConfig(ctx, config)

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
