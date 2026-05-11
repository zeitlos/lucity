package platform

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Interface interface {
	Projects(ctx context.Context, workspaceID string) ([]Project, error)
	Project(ctx context.Context, id ProjectID) (*Project, error)

	Environments(ctx context.Context, projectID ProjectID) ([]Environment, error)
	Environment(ctx context.Context, id EnvironmentID) (*Environment, error)

	Services(ctx context.Context, environmentID EnvironmentID) ([]Service, error)
	Service(ctx context.Context, id string) (*Service, error)

	Deployments(ctx context.Context, serviceID string) ([]Deployment, error)
	Deployment(ctx context.Context, id string) (*Deployment, error)

	Databases(ctx context.Context, environmentID EnvironmentID) ([]Database, error)
	Database(ctx context.Context, id string) (*Database, error)

	Volumes(ctx context.Context, environmentID EnvironmentID) ([]Volume, error)
	Volume(ctx context.Context, id string) (*Volume, error)

	// Buckets(ctx context.Context, environmentID EnvironmentID) ([]Bucket, error)
	// Bucket(ctx context.Context, id string) (*Bucket, error)
}

type ProjectID struct {
	Workspace string
	Name      string
}

type Project struct {
	ID           ProjectID
	Name         string
	Environments []Environment
}

type EnvironmentID struct {
	Workspace string
	Project   string
	Name      string
}

type ResourceTier int

const (
	EcoTier ResourceTier = iota
	ProductionTier
)

type Environment struct {
	ID           EnvironmentID
	Name         string
	CreatedAt    time.Time
	ResourceTier ResourceTier
}

type Service struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type Deployment struct {
	ID    string
	Image string
}

type Database struct {
	ID string
}

type Volume struct {
	ID string
}

// type Bucket struct {
// 	ID string
// }

func ParseProjectID(s string) (ProjectID, error) {
	ws, name, ok := strings.Cut(s, "/")

	if !ok || ws == "" || name == "" {
		return ProjectID{}, fmt.Errorf("invalid project id %q", s)
	}

	return ProjectID{Workspace: ws, Name: name}, nil
}

func (p ProjectID) String() string {
	return p.Workspace + "/" + p.Name
}

func ParseEnvironmentID(s string) (EnvironmentID, error) {
	ws, rest, ok := strings.Cut(s, "/")

	if !ok || ws == "" {
		return EnvironmentID{}, fmt.Errorf("invalid environment id %q", s)
	}

	proj, name, ok := strings.Cut(rest, "/")

	if !ok || proj == "" || name == "" {
		return EnvironmentID{}, fmt.Errorf("invalid environment id %q", s)
	}

	return EnvironmentID{Workspace: ws, Project: proj, Name: name}, nil
}

func (e EnvironmentID) String() string {
	return e.Workspace + "/" + e.Project + "/" + e.Name
}

func (e EnvironmentID) ProjectID() ProjectID {
	return ProjectID{Workspace: e.Workspace, Name: e.Project}
}

func (e EnvironmentID) Namespace() string {
	return e.Workspace + "-" + e.Project + "-" + e.Name
}
