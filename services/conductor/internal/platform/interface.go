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

	Services(ctx context.Context, environmentID EnvironmentID) ([]Service, error)
	Service(ctx context.Context, id ServiceID) (*Service, error)
	ServicesByRepo(ctx context.Context, repoURL, branch string) ([]ServiceID, error)

	Deployments(ctx context.Context, serviceID ServiceID) ([]Deployment, error)
	Deployment(ctx context.Context, id DeploymentID) (*Deployment, error)

	Databases(ctx context.Context, environmentID EnvironmentID) ([]Database, error)
	Database(ctx context.Context, id DatabaseID) (*Database, error)
	DatabaseCredentials(ctx context.Context, id DatabaseID) (*DatabaseCredentials, error)

	Volumes(ctx context.Context, environmentID EnvironmentID) ([]Volume, error)
	Volume(ctx context.Context, id VolumeID) (*Volume, error)

	Logs(ctx context.Context, id ServiceID, tail int) (<-chan LogEntry, error)

	ResourceAllocations(ctx context.Context) ([]ResourceAllocation, error)
}

// ResourceAllocation reports the actual resource usage of a single
// environment, summed across all pods (requests) and PVCs. Used by the
// billing pipeline (Cashier) to meter consumption.
type ResourceAllocation struct {
	EnvironmentID EnvironmentID
	Tier          ResourceTier
	CPUMillicores int
	MemoryMB      int
	DiskMB        int
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
}

// ErrDatabaseProvisioning is returned by DatabaseCredentials when the CNPG
// secret hasn't been created yet (cluster still bootstrapping). Callers
// typically translate this to a "still provisioning" UI state.
var ErrDatabaseProvisioning = errors.New("database is provisioning")

type WorkspaceScoped interface {
	WorkspaceID() string
}
