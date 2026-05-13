package platform

import (
	"context"
)

type Interface interface {
	Projects(ctx context.Context, workspaceID string) ([]Project, error)
	Project(ctx context.Context, id ProjectID) (*Project, error)

	Environments(ctx context.Context, projectID ProjectID) ([]Environment, error)
	Environment(ctx context.Context, id EnvironmentID) (*Environment, error)

	Services(ctx context.Context, environmentID EnvironmentID) ([]Service, error)
	Service(ctx context.Context, id ServiceID) (*Service, error)

	Deployments(ctx context.Context, serviceID ServiceID) ([]Deployment, error)
	Deployment(ctx context.Context, id DeploymentID) (*Deployment, error)

	Databases(ctx context.Context, environmentID EnvironmentID) ([]Database, error)
	Database(ctx context.Context, id DatabaseID) (*Database, error)

	Volumes(ctx context.Context, environmentID EnvironmentID) ([]Volume, error)
	Volume(ctx context.Context, id VolumeID) (*Volume, error)
}

type WorkspaceScoped interface {
	WorkspaceID() string
}
