package conductor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/dbquery"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/resources"
)

var (
	minDatabaseCPU        = resource.MustParse("250m")
	minDatabaseMemory     = resource.MustParse("256Mi")
	maxDatabaseCPU        = resource.MustParse("4")
	maxDatabaseMemory     = resource.MustParse("16Gi")
	defaultDatabaseCPU    = resources.DefaultCPULimit
	defaultDatabaseMemory = resources.DefaultMemoryLimit
)

const (
	backupRetentionDays = 30
	backupRetryInterval = time.Hour
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

func (c *Client) DatabaseBackups(ctx context.Context, id DatabaseID) (*platform.DatabaseBackups, error) {
	backups, err := c.platform.DatabaseBackups(ctx, id)

	if err != nil {
		return nil, err
	}

	backups.RetentionDays = backupRetentionDays
	backups.LatestRestorePoint = c.latestRestorePoint(ctx, id, backups.ServerName)

	return backups, nil
}

func (c *Client) latestRestorePoint(ctx context.Context, id DatabaseID, serverName string) *time.Time {
	if c.config.BackupArchive == nil || serverName == "" {
		return nil
	}

	latest, err := c.config.BackupArchive.LatestRestorePoint(ctx, id.Workspace, serverName)

	if err != nil {
		slog.Warn("failed to read latest restore point", "error", err, "database", id)
		return nil
	}

	return latest
}

func (c *Client) CreateDatabaseBackup(ctx context.Context, id DatabaseID) (*platform.DatabaseBackup, error) {
	return c.platform.CreateDatabaseBackup(ctx, id)
}

// ReconcileDatabaseBackups gives every archiving database a recovery floor. A
// scheduled backup fires the moment archiving is switched on, before the pods
// carrying the backup sidecar have rolled, so the first one always fails and the
// next is a week out. Until a base backup exists there is nothing to recover to,
// however much write-ahead log has been archived.
func (c *Client) ReconcileDatabaseBackups(ctx context.Context) error {
	workspaces, err := c.directory.Workspaces(ctx)

	if err != nil {
		return err
	}

	for _, workspace := range workspaces {
		environments, err := c.platform.EnvironmentsByWorkspace(ctx, workspace.ID)

		if err != nil {
			slog.Warn("reconcile backups: failed to list environments", "error", err, "workspace", workspace.ID)
			continue
		}

		for _, environment := range environments {
			databases, err := c.platform.Databases(ctx, environment.ID)

			if err != nil {
				slog.Warn("reconcile backups: failed to list databases", "error", err, "environment", environment.ID)
				continue
			}

			for _, database := range databases {
				if database.Status != platform.DatabaseHealthy {
					continue
				}

				if err := c.ensureRecoveryFloor(ctx, database.ID); err != nil {
					slog.Warn("reconcile backups: failed to ensure recovery floor", "error", err, "database", database.ID)
				}
			}
		}
	}

	return nil
}

func (c *Client) ensureRecoveryFloor(ctx context.Context, id DatabaseID) error {
	backups, err := c.platform.DatabaseBackups(ctx, id)

	if err != nil {
		return err
	}

	if !backups.Enabled || backups.EarliestRestorePoint != nil {
		return nil
	}

	for _, backup := range backups.Backups {
		if backup.Status == platform.BackupRunning || backup.Status == platform.BackupPending {
			return nil
		}
	}

	// Sorted newest first, so this bounds retries when backups fail for a reason
	// retrying will not fix.
	if len(backups.Backups) > 0 && time.Since(backups.Backups[0].CreatedAt) < backupRetryInterval {
		return nil
	}

	slog.Info("reconcile backups: no recovery point, starting a backup", "database", id)

	_, err = c.platform.CreateDatabaseBackup(ctx, id)

	return err
}

func (c *Client) CreateDatabase(ctx context.Context, environment platform.EnvironmentID, name string, size, cpu, memory string) (*Database, error) {
	const version = "17"
	const defaultDiskSize = "16Gi"

	if size == "" {
		size = defaultDiskSize
	}

	if cpu == "" {
		cpu = defaultDatabaseCPU.String()
	}

	if memory == "" {
		memory = defaultDatabaseMemory.String()
	}

	parsedSize, err := resource.ParseQuantity(size)

	if err != nil {
		return nil, fmt.Errorf("parse size: %w", err)
	}

	resources, err := validateDatabaseResources(cpu, memory)

	if err != nil {
		return nil, err
	}

	env, err := c.platform.Environment(ctx, environment)

	if err != nil {
		return nil, fmt.Errorf("read environment: %w", err)
	}

	if _, err := c.deployer.Databases().Create(ctx, environment, name, deployer.DatabaseSpec{
		Version:      version,
		Size:         parsedSize,
		ResourceTier: env.ResourceTier,
		Resources:    resources,
	}); err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	return &Database{
		Name:      name,
		Version:   version,
		Size:      parsedSize,
		Resources: platform.Resources{CPU: resources.CPU, Memory: resources.Memory},
	}, nil
}

type RestoreResult struct {
	Database Database
	Clamped  bool
}

func (c *Client) RestoreDatabase(ctx context.Context, source platform.DatabaseID, name string, targetTime *time.Time) (*RestoreResult, error) {
	origin, err := c.platform.Database(ctx, source)

	if err != nil {
		return nil, err
	}

	backups, err := c.DatabaseBackups(ctx, source)

	if err != nil {
		return nil, err
	}

	if !backups.Enabled {
		return nil, fmt.Errorf("database %q has no backups to restore from", source.Name)
	}

	if backups.EarliestRestorePoint == nil {
		return nil, errors.New("no backup has completed yet, so there is nothing to restore to")
	}

	// A nil target means the caller asked for the most recent state, which needs
	// no recovery target at all.
	var target *time.Time
	clamped := false

	if targetTime != nil {
		if targetTime.Before(*backups.EarliestRestorePoint) {
			return nil, fmt.Errorf("cannot restore to %s: the earliest available point is %s",
				targetTime.UTC().Format(time.RFC3339), backups.EarliestRestorePoint.UTC().Format(time.RFC3339))
		}

		if targetTime.After(time.Now()) {
			return nil, errors.New("cannot restore to a time in the future")
		}

		// Recovery stops at the first commit after the target, so a target past
		// the last archived commit can never be reached. Dropping the target
		// recovers to the end of the archive, which is the same state when the
		// gap exists because nothing was written. When archiving is broken the
		// gap means lost writes, and quietly handing back older data hides that.
		if latest := backups.LatestRestorePoint; latest != nil && targetTime.After(*latest) {
			if !backups.ArchivingHealthy {
				return nil, fmt.Errorf("cannot restore to %s: archiving is failing and nothing has reached the archive since %s",
					targetTime.UTC().Format(time.RFC3339), latest.UTC().Format(time.RFC3339))
			}

			clamped = true
		} else {
			target = targetTime
		}
	}

	env, err := c.platform.Environment(ctx, source.EnvironmentID())

	if err != nil {
		return nil, fmt.Errorf("read environment: %w", err)
	}

	resources, err := validateDatabaseResources(origin.Resources.CPU.String(), origin.Resources.Memory.String())

	if err != nil {
		return nil, err
	}

	if _, err := c.deployer.Databases().Restore(ctx, source, name, deployer.DatabaseSpec{
		Version:      origin.Version,
		Size:         origin.Size,
		ResourceTier: env.ResourceTier,
		Resources:    resources,
	}, target); err != nil {
		return nil, fmt.Errorf("restore database: %w", err)
	}

	return &RestoreResult{
		Clamped: clamped,
		Database: Database{
			ID:        platform.DatabaseID{Workspace: source.Workspace, Project: source.Project, Environment: source.Environment, Name: name},
			Name:      name,
			Version:   origin.Version,
			Size:      origin.Size,
			Status:    platform.DatabasePending,
			Resources: platform.Resources{CPU: resources.CPU, Memory: resources.Memory},
		},
	}, nil
}

func (c *Client) SetDatabaseResources(ctx context.Context, database platform.DatabaseID, cpu, memory string) (*Database, error) {
	spec, err := validateDatabaseResources(cpu, memory)

	if err != nil {
		return nil, err
	}

	environment, err := c.platform.Environment(ctx, database.EnvironmentID())

	if err != nil {
		return nil, err
	}

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
		URI:      creds.URI,
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

func validateDatabaseResources(cpu, memory string) (deployer.Resources, error) {
	return validateResources(cpu, memory, minDatabaseCPU, maxDatabaseCPU, minDatabaseMemory, maxDatabaseMemory)
}

func databaseURI(creds *platform.DatabaseCredentials) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		creds.User, creds.Password, creds.Host, creds.Port, creds.DBName)
}
