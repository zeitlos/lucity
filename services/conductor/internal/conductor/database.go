package conductor

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

// DatabaseProvisioningError indicates the database is still being provisioned
// and is not yet ready for queries.
type DatabaseProvisioningError struct{}

func (e *DatabaseProvisioningError) Error() string { return "database is provisioning" }

// dbQueryTimeout is longer than grpcTimeout because database queries can be slow.
const dbQueryTimeout = 35 * time.Second

type DatabaseID = platform.DatabaseID
type Database = platform.Database

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
	Name      string
	Ready     bool
	Instances int
	Version   string
	Size      string
	Volume    *Volume
}

type Volume struct {
	ID            platform.VolumeID
	Name          string
	Size          string
	RequestedSize string
	UsedBytes     int64
	CapacityBytes int64
}

func (c *Client) CreateDatabase(ctx context.Context, environment platform.EnvironmentID, name, version string, instances int, size string) (*Database, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

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
		return nil, fmt.Errorf("failed to parse size: %v", err)
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	if err := c.Packager.AddDatabase(callCtx, ws, environment.Project, data.DatabaseInfo{
		Name:      name,
		Version:   version,
		Instances: instances,
		Size:      size,
	}); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return &Database{
		Name:      name,
		Version:   version,
		Instances: instances,
		Size:      parsedSize,
	}, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, database platform.DatabaseID) (bool, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	if err := c.Packager.RemoveDatabase(callCtx, ws, database.Project, database.Name); err != nil {
		return false, fmt.Errorf("failed to delete database: %w", err)
	}
	return true, nil
}

func (c *Client) DatabaseTables(ctx context.Context, database platform.DatabaseID) ([]DatabaseTable, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	dbTables, err := c.Deployer.DatabaseTables(callCtx, ws, database.Project, database.Environment, database.Name)
	if err != nil {
		if s, ok := grpcstatus.FromError(err); ok && s.Code() == codes.FailedPrecondition {
			return nil, &DatabaseProvisioningError{}
		}
		return nil, fmt.Errorf("failed to get database tables: %w", err)
	}

	tables := make([]DatabaseTable, 0, len(dbTables))
	for _, t := range dbTables {
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

func (c *Client) DatabaseTableData(ctx context.Context, database platform.DatabaseID, table, schema string, limit, offset int) (*DatabaseTableData, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	resp, err := c.Deployer.DatabaseTableData(callCtx, ws, database.Project, database.Environment, database.Name, schema, table, limit, offset)
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

func (c *Client) ExecuteQuery(ctx context.Context, database platform.DatabaseID, query string) (*QueryResult, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, dbQueryTimeout)
	defer cancel()
	resp, err := c.Deployer.DatabaseQuery(callCtx, ws, database.Project, database.Environment, database.Name, query)
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

func (c *Client) DatabaseCredentials(ctx context.Context, database platform.DatabaseID) (*DatabaseCredentials, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	creds, err := c.Deployer.DatabaseCredentials(callCtx, ws, database.Project, database.Environment, database.Name)
	if err != nil {
		if s, ok := grpcstatus.FromError(err); ok && s.Code() == codes.FailedPrecondition {
			return nil, &DatabaseProvisioningError{}
		}
		return nil, fmt.Errorf("failed to get database credentials: %w", err)
	}

	return &DatabaseCredentials{
		Host:     creds.Host,
		Port:     creds.Port,
		DBName:   creds.DBName,
		User:     creds.User,
		Password: creds.Password,
		URI:      creds.URI,
	}, nil
}

// convertDatabaseRows converts data.DatabaseRow values to [][]*string for GraphQL.
func convertDatabaseRows(rows []data.DatabaseRow) [][]*string {
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
