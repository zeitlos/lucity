package deployer

import (
	"context"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type RevisionID string

type Interface interface {
	Services() ServiceClient
	Databases() DatabaseClient
	Volumes() VolumeClient
	Environments() EnvironmentClient
}

type ServiceClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, spec ServiceSpec) (RevisionID, error)
	Remove(ctx context.Context, id platform.ServiceID) error

	SetImage(ctx context.Context, id platform.ServiceID, ref, digest string) (RevisionID, error)
	SetReplicas(ctx context.Context, id platform.ServiceID, replicas int) (RevisionID, error)
	SetAutoscaling(ctx context.Context, id platform.ServiceID, config Autoscaling) (RevisionID, error)
	SetResources(ctx context.Context, id platform.ServiceID, resources Resources) (RevisionID, error)
	SetCommand(ctx context.Context, id platform.ServiceID, command string) (RevisionID, error)
	SetBranch(ctx context.Context, id platform.ServiceID, branch string) (RevisionID, error)
	SetPort(ctx context.Context, id platform.ServiceID, port int) (RevisionID, error)
	SetVariables(ctx context.Context, id platform.ServiceID, vars map[string]string) (RevisionID, error)

	AddDomain(ctx context.Context, id platform.ServiceID, host string) (RevisionID, error)
	RemoveDomain(ctx context.Context, id platform.ServiceID, host string) (RevisionID, error)
	VerifyDomain(ctx context.Context, id platform.ServiceID, host string, verified bool) (RevisionID, error)

	Mount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID, mountPath string) (RevisionID, error)
	Unmount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID) (RevisionID, error)
}

type DatabaseClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, spec DatabaseSpec) (RevisionID, error)
	Delete(ctx context.Context, id platform.DatabaseID) error
}

type VolumeClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, spec VolumeSpec) (RevisionID, error)
	Delete(ctx context.Context, id platform.VolumeID) error
}

type EnvironmentClient interface {
	SetVariables(ctx context.Context, id platform.EnvironmentID, vars map[string]string) (RevisionID, error)
	Suspend(ctx context.Context, id platform.EnvironmentID, suspended bool) (RevisionID, error)
}

type ServiceSpec struct {
	Image                string
	Port                 int
	SourceURL            string
	ContextPath          string
	Branch               string
	GitHubInstallationID int64
	StartCommand         string
}

type DatabaseSpec struct {
	Version   string
	Instances int
	Size      resource.Quantity
}

type VolumeSpec struct {
	Size resource.Quantity
}

type Autoscaling struct {
	MinReplicas int
	MaxReplicas int
	TargetCPU   int
}

type Resources struct {
	Requests ResourceList
	Limits   ResourceList
}

type ResourceList struct {
	CPU    resource.Quantity
	Memory resource.Quantity
}
