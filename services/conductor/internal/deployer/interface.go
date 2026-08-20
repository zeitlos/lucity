package deployer

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/image"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type RevisionID string

type Interface interface {
	Services() ServiceClient
	Databases() DatabaseClient
	KeyValueStores() KeyValueStoreClient
	Volumes() VolumeClient
	Environments() EnvironmentClient
}

type ServiceClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, spec ServiceSpec) (RevisionID, error)
	Delete(ctx context.Context, id platform.ServiceID) error

	SetImage(ctx context.Context, id platform.ServiceID, ref image.Ref, provenance ImageProvenance, release ReleaseMeta) (RevisionID, error)
	SetReplicas(ctx context.Context, id platform.ServiceID, replicas int) (RevisionID, error)
	SetAutoscaling(ctx context.Context, id platform.ServiceID, config Autoscaling) (RevisionID, error)
	SetResources(ctx context.Context, id platform.ServiceID, tier platform.ResourceTier, resources Resources) (RevisionID, error)
	SetCommand(ctx context.Context, id platform.ServiceID, command string) (RevisionID, error)
	SetBranch(ctx context.Context, id platform.ServiceID, branch string) (RevisionID, error)
	SetAutoDeploy(ctx context.Context, id platform.ServiceID, enabled bool) (RevisionID, error)
	SetCIDeploy(ctx context.Context, id platform.ServiceID, enabled bool) (RevisionID, error)
	SetPort(ctx context.Context, id platform.ServiceID, port int) (RevisionID, error)
	SetHealthCheck(ctx context.Context, id platform.ServiceID, healthCheck *HealthCheck) (RevisionID, error)
	SetSecurityContext(ctx context.Context, id platform.ServiceID, sc SecurityContext) (RevisionID, error)

	Variables(ctx context.Context, id platform.ServiceID) (ServiceVariablesSpec, error)
	SetVariables(ctx context.Context, id platform.ServiceID, spec ServiceVariablesSpec) (RevisionID, error)

	AddDomain(ctx context.Context, id platform.ServiceID, host string, ownListener bool) (RevisionID, error)
	RemoveDomain(ctx context.Context, id platform.ServiceID, host string) (RevisionID, error)
	AttachDomain(ctx context.Context, id platform.ServiceID, host string, attached bool) (RevisionID, error)

	Mount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID, mountPath string) (RevisionID, error)
	Unmount(ctx context.Context, volume platform.VolumeID) (RevisionID, error)
}

type DatabaseClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, spec DatabaseSpec) (RevisionID, error)
	Restore(ctx context.Context, source platform.DatabaseID, name string, spec DatabaseSpec, targetTime *time.Time) (RevisionID, error)
	Delete(ctx context.Context, id platform.DatabaseID) error
	SetResources(ctx context.Context, id platform.DatabaseID, tier platform.ResourceTier, resources Resources) (RevisionID, error)
	SetStorage(ctx context.Context, id platform.DatabaseID, size resource.Quantity) (RevisionID, error)
	Expose(ctx context.Context, id platform.DatabaseID, host string) error
	Unexpose(ctx context.Context, id platform.DatabaseID) error
}

type KeyValueStoreClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, spec KeyValueStoreSpec) (RevisionID, error)
	Delete(ctx context.Context, id platform.KeyValueStoreID) error
}

type VolumeClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, size resource.Quantity) (RevisionID, error)
	Delete(ctx context.Context, id platform.VolumeID) error
	Expand(ctx context.Context, id platform.VolumeID, size resource.Quantity) (RevisionID, error)
}

type EnvironmentClient interface {
	Variables(ctx context.Context, id platform.EnvironmentID) (map[string]string, error)
	SetVariables(ctx context.Context, id platform.EnvironmentID, vars map[string]string) (RevisionID, error)
	Suspend(ctx context.Context, id platform.EnvironmentID, suspended bool) (RevisionID, error)
	Reconcile(ctx context.Context, id platform.EnvironmentID) (RevisionID, error)
	Export(ctx context.Context, id platform.EnvironmentID) ([]byte, error)
}

type ServiceVariablesSpec struct {
	Literals map[string]string
	Refs     map[string]VariableRef
}

type VariableRef struct {
	Secret string
	Key    string
}

type ServiceSpec struct {
	Image                string
	SourceURL            string
	ContextPath          string
	GitHubInstallationID int64
	AutoDeploy           bool
	StartCommand         string
	Port                 int
	Resources            Resources
	ResourceTier         platform.ResourceTier
	Env                  map[string]string
	SecurityContext      SecurityContext
}

type ImageProvenance struct {
	Commit        string
	CommitMessage string
	BuildID       string
}

type DatabaseSpec struct {
	Version      string
	Size         resource.Quantity
	Resources    Resources
	ResourceTier platform.ResourceTier
}

type KeyValueStoreSpec struct {
	Version  string
	Size     resource.Quantity
	Password string
}

type Autoscaling struct {
	MinReplicas int
	MaxReplicas int
	TargetCPU   int
}

type HealthCheck struct {
	Path                    string
	Port                    int
	InitialDelaySeconds     int
	PeriodSeconds           int
	TimeoutSeconds          int
	FailureThreshold        int
	StartupFailureThreshold int
}

type Resources struct {
	CPU    resource.Quantity
	Memory resource.Quantity
}

// SecurityContext holds the run-as user/group and the volume-owning group
// (fsGroup) for a service. A nil field means unset (image default).
type SecurityContext struct {
	RunAsUser  *int64
	RunAsGroup *int64
	FsGroup    *int64
}
