package conductor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/dbquery"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

// DatabaseProvisioningError indicates the database is still being provisioned
// and is not yet ready for queries.
type DatabaseProvisioningError struct{}

func (e *DatabaseProvisioningError) Error() string { return "database is provisioning" }

type DatabaseID = platform.DatabaseID
type Database = platform.Database
type DatabaseStatus = platform.DatabaseStatus

type DatabaseTable struct {
	Name          string
	Schema        string
	EstimatedRows int
	Columns       []DatabaseColumn
}

type DatabaseColumn struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
}

type DatabaseTableData struct {
	Columns            []string
	Rows               [][]*string
	TotalEstimatedRows int
}

type QueryResult struct {
	Columns      []string
	Rows         [][]*string
	AffectedRows int
}

type DatabaseCredentials struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
	URI      string
}

func (c *Client) Databases(ctx context.Context, environment EnvironmentID) ([]Database, error) {
	return c.platform.Databases(ctx, environment)
}

func (c *Client) Database(ctx context.Context, id DatabaseID) (*Database, error) {
	return c.platform.Database(ctx, id)
}

func (c *Client) CreateDatabase(ctx context.Context, environment platform.EnvironmentID, name, version string, instances int, size string) (*Database, error) {
	if version == "" {
		version = "16"
	}

	if instances == 0 {
		instances = 1
	}

	if size == "" {
		size = "10Gi"
	}

	parsedSize, err := resource.ParseQuantity(size)

	if err != nil {
		return nil, fmt.Errorf("parse size: %w", err)
	}

	if _, err := c.deployer.Databases().Create(ctx, environment, name, deployer.DatabaseSpec{
		Version:   version,
		Instances: instances,
		Size:      parsedSize,
	}); err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	return &Database{
		Name:      name,
		Version:   version,
		Instances: instances,
		Size:      parsedSize,
	}, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, database platform.DatabaseID) (bool, error) {
	if err := c.deployer.Databases().Delete(ctx, database); err != nil {
		return false, fmt.Errorf("delete database: %w", err)
	}

	return true, nil
}

func (c *Client) DatabaseTables(ctx context.Context, database platform.DatabaseID) ([]DatabaseTable, error) {
	query, err := c.databaseQueryClient(ctx, database)

	if err != nil {
		return nil, err
	}

	tables, err := query.Tables(ctx)

	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}

	result := make([]DatabaseTable, 0, len(tables))

	for _, t := range tables {
		table := DatabaseTable{
			Name:          t.Name,
			Schema:        t.Schema,
			EstimatedRows: int(t.EstimatedRows),
		}

		for _, col := range t.Columns {
			table.Columns = append(table.Columns, DatabaseColumn{
				Name:       col.Name,
				Type:       col.Type,
				Nullable:   col.Nullable,
				PrimaryKey: col.PrimaryKey,
			})
		}

		result = append(result, table)
	}

	return result, nil
}

func (c *Client) DatabaseTableData(ctx context.Context, database platform.DatabaseID, table, schema string, limit, offset int) (*DatabaseTableData, error) {
	query, err := c.databaseQueryClient(ctx, database)

	if err != nil {
		return nil, err
	}

	rows, err := query.Rows(ctx, schema, table, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("read table rows: %w", err)
	}

	return &DatabaseTableData{
		Columns:            rows.Columns,
		Rows:               rows.Rows,
		TotalEstimatedRows: int(rows.TotalEstimatedRows),
	}, nil
}

func (c *Client) ExecuteQuery(ctx context.Context, database platform.DatabaseID, sql string) (*QueryResult, error) {
	query, err := c.databaseQueryClient(ctx, database)

	if err != nil {
		return nil, err
	}

	result, err := query.Query(ctx, sql)

	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}

	return &QueryResult{
		Columns:      result.Columns,
		Rows:         result.Rows,
		AffectedRows: int(result.AffectedRows),
	}, nil
}

func (c *Client) DatabaseCredentials(ctx context.Context, database platform.DatabaseID) (*DatabaseCredentials, error) {
	creds, err := c.platform.DatabaseCredentials(ctx, database)

	if errors.Is(err, platform.ErrDatabaseProvisioning) {
		return nil, &DatabaseProvisioningError{}
	}

	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	return &DatabaseCredentials{
		Host:     creds.Host,
		Port:     creds.Port,
		DBName:   creds.DBName,
		User:     creds.User,
		Password: creds.Password,
		URI:      databaseURI(creds),
	}, nil
}

// databaseQueryClient resolves DB credentials, builds a dev-friendly DSN
// (falls back to localhost when the cluster DNS doesn't resolve, which is
// the case from outside the cluster), and returns a dbquery.Client.
func (c *Client) databaseQueryClient(ctx context.Context, id platform.DatabaseID) (*dbquery.Client, error) {
	creds, err := c.platform.DatabaseCredentials(ctx, id)

	if errors.Is(err, platform.ErrDatabaseProvisioning) {
		return nil, &DatabaseProvisioningError{}
	}

	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	host := creds.Host

	if _, err := net.LookupHost(host); err != nil {
		slog.Debug("cnpg host unresolvable, falling back to localhost", "host", host)
		host = "localhost"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		creds.User, creds.Password, host, creds.Port, creds.DBName)

	return dbquery.New(dsn), nil
}

func databaseURI(creds *platform.DatabaseCredentials) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		creds.User, creds.Password, creds.Host, creds.Port, creds.DBName)
}
