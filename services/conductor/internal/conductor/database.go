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
	"github.com/zeitlos/lucity/services/conductor/internal/resources"
)

var (
	minDatabaseCPU    = resource.MustParse("250m")
	minDatabaseMemory = resource.MustParse("256Mi")
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
	Type     EndpointType
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

func (c *Client) CreateDatabase(ctx context.Context, environment platform.EnvironmentID, name, version string, size string) (*Database, error) {
	if version == "" {
		version = "17"
	}

	if size == "" {
		size = "20Gi"
	}

	parsedSize, err := resource.ParseQuantity(size)

	if err != nil {
		return nil, fmt.Errorf("parse size: %w", err)
	}

	env, err := c.platform.Environment(ctx, environment)

	if err != nil {
		return nil, fmt.Errorf("read environment: %w", err)
	}

	if _, err := c.deployer.Databases().Create(ctx, environment, name, deployer.DatabaseSpec{
		Version:      version,
		Size:         parsedSize,
		ResourceTier: env.ResourceTier,
	}); err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	return &Database{
		Name:    name,
		Version: version,
		Size:    parsedSize,
	}, nil
}

func (c *Client) SetDatabaseResources(ctx context.Context, database platform.DatabaseID, cpu, memory string) (*Database, error) {
	cpuQuantity, err := resource.ParseQuantity(cpu)

	if err != nil {
		return nil, fmt.Errorf("invalid cpu value %q: %w", cpu, err)
	}

	memoryQuantity, err := resource.ParseQuantity(memory)

	if err != nil {
		return nil, fmt.Errorf("invalid memory value %q: %w", memory, err)
	}

	if cpuQuantity.Cmp(minDatabaseCPU) < 0 {
		return nil, fmt.Errorf("cpu must be at least %s", minDatabaseCPU.String())
	}

	if memoryQuantity.Cmp(minDatabaseMemory) < 0 {
		return nil, fmt.Errorf("memory must be at least %s", minDatabaseMemory.String())
	}

	if cpuQuantity.Cmp(resources.DefaultCPUQuota) > 0 {
		return nil, fmt.Errorf("cpu exceeds the maximum of %s", resources.DefaultCPUQuota.String())
	}

	if memoryQuantity.Cmp(resources.DefaultMemoryQuota) > 0 {
		return nil, fmt.Errorf("memory exceeds the maximum of %s", resources.DefaultMemoryQuota.String())
	}

	environment, err := c.platform.Environment(ctx, database.EnvironmentID())

	if err != nil {
		return nil, err
	}

	spec := deployer.Resources{CPU: cpuQuantity, Memory: memoryQuantity}

	if _, err := c.deployer.Databases().SetResources(ctx, database, environment.ResourceTier, spec); err != nil {
		return nil, fmt.Errorf("set resources: %w", err)
	}

	return c.Database(ctx, database)
}

func (c *Client) SetDatabaseStorage(ctx context.Context, database platform.DatabaseID, size string) (*Database, error) {
	parsedSize, err := resource.ParseQuantity(size)

	if err != nil {
		return nil, fmt.Errorf("invalid storage size %q: %w", size, err)
	}

	if parsedSize.Cmp(resources.MaxDatabaseStorage) > 0 {
		return nil, fmt.Errorf("storage exceeds the maximum of %s", resources.MaxDatabaseStorage.String())
	}

	current, err := c.platform.Database(ctx, database)

	if err != nil {
		return nil, err
	}

	if parsedSize.Cmp(current.Size) < 0 {
		return nil, fmt.Errorf("storage can only be increased, not shrunk from %s to %s", current.Size.String(), parsedSize.String())
	}

	if _, err := c.deployer.Databases().SetStorage(ctx, database, parsedSize); err != nil {
		return nil, fmt.Errorf("set storage: %w", err)
	}

	return c.Database(ctx, database)
}

func (c *Client) DeleteDatabase(ctx context.Context, database platform.DatabaseID) (bool, error) {
	if err := c.deployer.Databases().Delete(ctx, database); err != nil {
		return false, fmt.Errorf("delete database: %w", err)
	}

	return true, nil
}

func (c *Client) ExposeDatabase(ctx context.Context, id platform.DatabaseID) (*Database, error) {
	database, err := c.platform.Database(ctx, id)

	if err != nil {
		return nil, err
	}

	if database.PublicHost != "" {
		return database, nil
	}

	host := fmt.Sprintf("%s-%s-%s.%s", id.Name, id.Environment, randCrockford32(5), c.config.DatabaseDomain)

	if err := c.deployer.Databases().Expose(ctx, id, host); err != nil {
		return nil, fmt.Errorf("expose database: %w", err)
	}

	database.PublicHost = host

	return database, nil
}

func (c *Client) UnexposeDatabase(ctx context.Context, id platform.DatabaseID) (*Database, error) {
	if err := c.deployer.Databases().Unexpose(ctx, id); err != nil {
		return nil, fmt.Errorf("unexpose database: %w", err)
	}

	database, err := c.platform.Database(ctx, id)

	if err != nil {
		return nil, err
	}

	database.PublicHost = ""

	return database, nil
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

func (c *Client) DatabaseCredentials(ctx context.Context, database platform.DatabaseID) ([]DatabaseCredentials, error) {
	creds, err := c.platform.DatabaseCredentials(ctx, database)

	if errors.Is(err, platform.ErrDatabaseProvisioning) {
		return nil, &DatabaseProvisioningError{}
	}

	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	result := []DatabaseCredentials{{
		Type:     InternalEndpoint,
		Host:     creds.Host,
		Port:     creds.Port,
		DBName:   creds.DBName,
		User:     creds.User,
		Password: creds.Password,
		URI:      databaseURI(creds),
	}}

	db, err := c.platform.Database(ctx, database)

	if err != nil {
		return nil, err
	}

	if db.PublicHost != "" {
		public := *creds
		public.Host = db.PublicHost
		public.Port = "5432"

		result = append(result, DatabaseCredentials{
			Type:     PlatformEndpoint,
			Host:     public.Host,
			Port:     public.Port,
			DBName:   public.DBName,
			User:     public.User,
			Password: public.Password,
			URI:      databaseURI(&public) + "?sslmode=require",
		})
	}

	return result, nil
}

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

	creds.Host = host

	return dbquery.New(databaseURI(creds) + "?sslmode=require"), nil
}

func databaseURI(creds *platform.DatabaseCredentials) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		creds.User, creds.Password, creds.Host, creds.Port, creds.DBName)
}
