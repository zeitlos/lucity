package platform

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
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
	ResourceTier ResourceTier
	CreatedAt    time.Time
}

type ReplicaCount struct {
	Desired int
	Ready   int
}

type ServiceStatus string

const (
	ServiceHealthy   ServiceStatus = "healthy"
	ServiceDegraded  ServiceStatus = "degraded" // some replicas not ready
	ServiceDeploying ServiceStatus = "deploying"
	ServiceFailed    ServiceStatus = "failed"  // no working replicas
	ServiceStopped   ServiceStatus = "stopped" // intentionally scaled to 0
)

type ServiceID struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
}

type Service struct {
	ID   ServiceID
	Name string

	Status   ServiceStatus
	Replicas ReplicaCount

	InternalHost string
	URLs         []url.URL

	CreatedAt time.Time
}

type DeploymentStatus string

const (
	DeploymentDeploying  DeploymentStatus = "deploying"
	DeploymentActive     DeploymentStatus = "active"
	DeploymentSuperseded DeploymentStatus = "superseded"
	DeploymentFailed     DeploymentStatus = "failed"
)

type DeploymentID struct {
	Workspace   string
	Project     string
	Environment string
	Service     string
	Hash        string
}

type Deployment struct {
	ID DeploymentID

	Image       string
	ImageDigest string

	Commit string
	Ref    string
	Repo   string

	Status   DeploymentStatus
	Replicas ReplicaCount

	CreatedAt time.Time
}

type DatabaseStatus string

const (
	DatabaseHealthy  DatabaseStatus = "healthy"
	DatabaseDegraded DatabaseStatus = "degraded"
	DatabaseFailed   DatabaseStatus = "failed"
	DatabasePending  DatabaseStatus = "pending"
	DatabaseStopped  DatabaseStatus = "stopped"
)

type DatabaseID struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
}

// TODO: Rename to Postgres, PostgresDatabase or PostgresCluster
type Database struct {
	ID        DatabaseID
	Name      string
	Version   string
	Instances int
	Status    DatabaseStatus
	CreatedAt time.Time
}

type VolumeID struct {
	Workspace   string
	Project     string
	Environment string
	Name        string
}

type VolumeStatus string

const (
	VolumeReady   VolumeStatus = "ready"
	VolumePending VolumeStatus = "pending"
	VolumeFailed  VolumeStatus = "failed"
)

type Volume struct {
	ID        VolumeID
	Name      string
	Size      resource.Quantity
	Status    VolumeStatus
	CreatedAt time.Time
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

func ParseServiceID(s string) (ServiceID, error) {
	parts := strings.SplitN(s, "/", 4)

	if len(parts) != 4 || slices.Contains(parts, "") {
		return ServiceID{}, fmt.Errorf("invalid service id %q", s)
	}

	return ServiceID{
		Workspace: parts[0], Project: parts[1],
		Environment: parts[2], Name: parts[3],
	}, nil
}

func (s ServiceID) String() string {
	return s.Workspace + "/" + s.Project + "/" + s.Environment + "/" + s.Name
}

func (s ServiceID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: s.Workspace, Project: s.Project, Name: s.Environment}
}

func (s ServiceID) Namespace() string {
	return s.EnvironmentID().Namespace()
}

func ParseDeploymentID(s string) (DeploymentID, error) {
	parts := strings.SplitN(s, "/", 5)

	if len(parts) != 5 || slices.Contains(parts, "") {
		return DeploymentID{}, fmt.Errorf("invalid deployment id %q", s)
	}

	return DeploymentID{
		Workspace:   parts[0],
		Project:     parts[1],
		Environment: parts[2],
		Service:     parts[3],
		Hash:        parts[4],
	}, nil
}

func (d DeploymentID) String() string {
	return d.Workspace + "/" + d.Project + "/" + d.Environment + "/" + d.Service + "/" + d.Hash
}

func (d DeploymentID) ServiceID() ServiceID {
	return ServiceID{
		Workspace:   d.Workspace,
		Project:     d.Project,
		Environment: d.Environment,
		Name:        d.Service,
	}
}

func (d DeploymentID) Namespace() string {
	return d.ServiceID().Namespace()
}

func ParseDatabaseID(s string) (DatabaseID, error) {
	parts := strings.SplitN(s, "/", 4)

	if len(parts) != 4 || slices.Contains(parts, "") {
		return DatabaseID{}, fmt.Errorf("invalid database id %q", s)
	}

	return DatabaseID{
		Workspace: parts[0], Project: parts[1],
		Environment: parts[2], Name: parts[3],
	}, nil
}

func (s DatabaseID) String() string {
	return s.Workspace + "/" + s.Project + "/" + s.Environment + "/" + s.Name
}

func (s DatabaseID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: s.Workspace, Project: s.Project, Name: s.Environment}
}

func (s DatabaseID) Namespace() string {
	return s.EnvironmentID().Namespace()
}

func ParseVolumeID(s string) (VolumeID, error) {
	parts := strings.SplitN(s, "/", 4)

	if len(parts) != 4 || slices.Contains(parts, "") {
		return VolumeID{}, fmt.Errorf("invalid volume id %q", s)
	}

	return VolumeID{
		Workspace: parts[0], Project: parts[1],
		Environment: parts[2], Name: parts[3],
	}, nil
}

func (s VolumeID) String() string {
	return s.Workspace + "/" + s.Project + "/" + s.Environment + "/" + s.Name
}

func (s VolumeID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: s.Workspace, Project: s.Project, Name: s.Environment}
}

func (s VolumeID) Namespace() string {
	return s.EnvironmentID().Namespace()
}
