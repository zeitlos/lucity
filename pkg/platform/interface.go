package platform

import (
	"context"
	"errors"
)

type Interface interface {
	Projects(ctx context.Context, workspaceID string) ([]Project, error)
	Project(ctx context.Context, id ProjectID) (*Project, error)

	Environments(ctx context.Context, projectID ProjectID) ([]Environment, error)
	Environment(ctx context.Context, id EnvironmentID) (*Environment, error)
	EnvironmentsByWorkspace(ctx context.Context, workspaceID string) ([]Environment, error)

	Services(ctx context.Context, environmentID EnvironmentID) ([]Service, error)
	Service(ctx context.Context, id ServiceID) (*Service, error)
	ServicesByRepo(ctx context.Context, repoURL string) ([]RepoService, error)
	PendingDomains(ctx context.Context) ([]PendingDomain, error)
	RoutedWorkspaces(ctx context.Context) (map[string]bool, error)
	BuildVariables(ctx context.Context, id ServiceID) (map[string]string, error)

	Deployments(ctx context.Context, serviceID ServiceID) ([]Deployment, error)
	Deployment(ctx context.Context, id DeploymentID) (*Deployment, error)

	Databases(ctx context.Context, environmentID EnvironmentID) ([]Database, error)
	Database(ctx context.Context, id DatabaseID) (*Database, error)
	DatabaseCredentials(ctx context.Context, id DatabaseID) (*DatabaseCredentials, error)
	DatabaseBackups(ctx context.Context, id DatabaseID) (*DatabaseBackups, error)
	CreateDatabaseBackup(ctx context.Context, id DatabaseID) (*DatabaseBackup, error)

	KeyValueStores(ctx context.Context, environmentID EnvironmentID) ([]KeyValueStore, error)
	KeyValueStore(ctx context.Context, id KeyValueStoreID) (*KeyValueStore, error)
	KeyValueStoreCredentials(ctx context.Context, id KeyValueStoreID) (*KeyValueStoreCredentials, error)

	AvailableVariables(ctx context.Context, environmentID EnvironmentID) ([]Variable, error)

	Volumes(ctx context.Context, environmentID EnvironmentID) ([]Volume, error)
	Volume(ctx context.Context, id VolumeID) (*Volume, error)

	Logs(ctx context.Context, id ServiceID, tail int) (<-chan LogEntry, error)
}

type Changes interface {
	// Watch signals every change to the objects behind the environment until
	// ctx is done. A burst of changes arrives as one signal.
	Watch(ctx context.Context, environment EnvironmentID) <-chan struct{}
}

type LogEntry struct {
	Pod  string
	Line string
}

type DatabaseCredentials struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
	URI      string
}

// ErrDatabaseProvisioning is returned by DatabaseCredentials when the CNPG
// secret hasn't been created yet (cluster still bootstrapping). Callers
// typically translate this to a "still provisioning" UI state.
var ErrDatabaseProvisioning = errors.New("database is provisioning")

type WorkspaceScoped interface {
	WorkspaceID() string
}
