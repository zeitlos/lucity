package handler

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/pkg/tenant"
)

// DatabaseProvisioningError indicates the database is still being provisioned
// and is not yet ready for queries.
type DatabaseProvisioningError struct{}

func (e *DatabaseProvisioningError) Error() string { return "database is provisioning" }

// dbQueryTimeout is longer than grpcTimeout because database queries can be slow.
const dbQueryTimeout = 35 * time.Second

type Database struct {
	Name      string
	Version   string
	Instances int
	Size      string
}

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

type DatabaseInstance struct {
	Name        string
	Environment string
	Ready       bool
	Instances   int
	Version     string
	Size        string
	Volume      *Volume
}

type Volume struct {
	Name          string
	Size          string
	RequestedSize string
	UsedBytes     int64
	CapacityBytes int64
}

func (c *Client) CreateDatabase(ctx context.Context, projectID, name, version string, instances int, size string) (*Database, error) {
	if _, err := tenant.Require(ctx); err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)
	ctx = tenant.OutgoingContext(ctx)

	if version == "" {
		version = "16"
	}
	if instances == 0 {
		instances = 1
	}
	if size == "" {
		size = "10Gi"
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.AddDatabase(callCtx, &packager.AddDatabaseRequest{
		Project:   projectID,
		Name:      name,
		Version:   version,
		Instances: int32(instances),
		Size:      size,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return &Database{
		Name:      name,
		Version:   version,
		Instances: instances,
		Size:      size,
	}, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, projectID, name string) (bool, error) {
	if _, err := tenant.Require(ctx); err != nil {
		return false, err
	}
	ctx = auth.OutgoingContext(ctx)
	ctx = tenant.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.RemoveDatabase(callCtx, &packager.RemoveDatabaseRequest{
		Project: projectID,
		Name:    name,
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete database: %w", err)
	}
	return true, nil
}

func (c *Client) Databases(ctx context.Context, projectID string) ([]Database, error) {
	proj, err := c.Project(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return proj.Databases, nil
}

func (c *Client) DatabaseTables(ctx context.Context, projectID, environment, database string) ([]DatabaseTable, error) {
	if _, err := tenant.Require(ctx); err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)
	ctx = tenant.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	resp, err := c.Deployer.DatabaseTables(callCtx, &deployer.DatabaseTablesRequest{
		Project:     projectID,
		Environment: environment,
		Database:    database,
	})
	if err != nil {
		if s, ok := grpcstatus.FromError(err); ok && s.Code() == codes.FailedPrecondition {
			return nil, &DatabaseProvisioningError{}
		}
		return nil, fmt.Errorf("failed to get database tables: %w", err)
	}

	tables := make([]DatabaseTable, 0, len(resp.Tables))
	for _, t := range resp.Tables {
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
		tables = append(tables, table)
	}
	return tables, nil
}

func (c *Client) DatabaseTableData(ctx context.Context, projectID, environment, database, table, schema string, limit, offset int) (*DatabaseTableData, error) {
	if _, err := tenant.Require(ctx); err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)
	ctx = tenant.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	resp, err := c.Deployer.DatabaseTableData(callCtx, &deployer.DatabaseTableDataRequest{
		Project:     projectID,
		Environment: environment,
		Database:    database,
		Table:       table,
		Schema:      schema,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		if s, ok := grpcstatus.FromError(err); ok && s.Code() == codes.FailedPrecondition {
			return nil, &DatabaseProvisioningError{}
		}
		return nil, fmt.Errorf("failed to get table data: %w", err)
	}

	return &DatabaseTableData{
		Columns:            resp.Columns,
		Rows:               convertDatabaseRows(resp.Rows),
		TotalEstimatedRows: int(resp.TotalEstimatedRows),
	}, nil
}

func (c *Client) ExecuteQuery(ctx context.Context, projectID, environment, database, query string) (*QueryResult, error) {
	if _, err := tenant.Require(ctx); err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)
	ctx = tenant.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	resp, err := c.Deployer.DatabaseQuery(callCtx, &deployer.DatabaseQueryRequest{
		Project:     projectID,
		Environment: environment,
		Database:    database,
		Query:       query,
	})
	if err != nil {
		if s, ok := grpcstatus.FromError(err); ok && s.Code() == codes.FailedPrecondition {
			return nil, &DatabaseProvisioningError{}
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &QueryResult{
		Columns:      resp.Columns,
		Rows:         convertDatabaseRows(resp.Rows),
		AffectedRows: int(resp.AffectedRows),
	}, nil
}

type DatabaseCredentials struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
	URI      string
}

func (c *Client) DatabaseCredentials(ctx context.Context, projectID, environment, database string) (*DatabaseCredentials, error) {
	if _, err := tenant.Require(ctx); err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)
	ctx = tenant.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Deployer.DatabaseCredentials(callCtx, &deployer.DatabaseCredentialsRequest{
		Project:     projectID,
		Environment: environment,
		Database:    database,
	})
	if err != nil {
		if s, ok := grpcstatus.FromError(err); ok && s.Code() == codes.FailedPrecondition {
			return nil, &DatabaseProvisioningError{}
		}
		return nil, fmt.Errorf("failed to get database credentials: %w", err)
	}

	return &DatabaseCredentials{
		Host:     resp.Host,
		Port:     resp.Port,
		DBName:   resp.Dbname,
		User:     resp.User,
		Password: resp.Password,
		URI:      resp.Uri,
	}, nil
}

// convertDatabaseRows converts proto DatabaseRow messages to [][]*string for GraphQL.
func convertDatabaseRows(rows []*deployer.DatabaseRow) [][]*string {
	result := make([][]*string, 0, len(rows))
	for _, row := range rows {
		vals := make([]*string, len(row.Cells))
		for i, cell := range row.Cells {
			if cell.IsNull {
				vals[i] = nil
			} else {
				v := cell.Value
				vals[i] = &v
			}
		}
		result = append(result, vals)
	}
	return result
}
