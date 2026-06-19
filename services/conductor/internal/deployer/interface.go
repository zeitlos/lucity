package deployer

import (
	"context"

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

	SetImage(ctx context.Context, id platform.ServiceID, ref image.Ref, provenance ImageProvenance) (RevisionID, error)
	SetReplicas(ctx context.Context, id platform.ServiceID, replicas int) (RevisionID, error)
	SetAutoscaling(ctx context.Context, id platform.ServiceID, config Autoscaling) (RevisionID, error)
	SetResources(ctx context.Context, id platform.ServiceID, tier platform.ResourceTier, resources Resources) (RevisionID, error)
	SetCommand(ctx context.Context, id platform.ServiceID, command string) (RevisionID, error)
	SetBranch(ctx context.Context, id platform.ServiceID, branch string) (RevisionID, error)
	SetPort(ctx context.Context, id platform.ServiceID, port int) (RevisionID, error)

	Variables(ctx context.Context, id platform.ServiceID) (ServiceVariablesSpec, error)
	SetVariables(ctx context.Context, id platform.ServiceID, spec ServiceVariablesSpec) (RevisionID, error)

	AddDomain(ctx context.Context, id platform.ServiceID, host string) (RevisionID, error)
	RemoveDomain(ctx context.Context, id platform.ServiceID, host string) (RevisionID, error)
	VerifyDomain(ctx context.Context, id platform.ServiceID, host string, verified bool) (RevisionID, error)

	Mount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID, mountPath string) (RevisionID, error)
	Unmount(ctx context.Context, id platform.ServiceID, volume platform.VolumeID) (RevisionID, error)
}

type DatabaseClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, spec DatabaseSpec) (RevisionID, error)
	Delete(ctx context.Context, id platform.DatabaseID) error
	Expose(ctx context.Context, id platform.DatabaseID, host string) error
	Unexpose(ctx context.Context, id platform.DatabaseID) error
}

type KeyValueStoreClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, spec KeyValueStoreSpec) (RevisionID, error)
	Delete(ctx context.Context, id platform.KeyValueStoreID) error
}

type VolumeClient interface {
	Create(ctx context.Context, env platform.EnvironmentID, name string, spec VolumeSpec) (RevisionID, error)
	Delete(ctx context.Context, id platform.VolumeID) error
}

type EnvironmentClient interface {
	Variables(ctx context.Context, id platform.EnvironmentID) (map[string]string, error)
	SetVariables(ctx context.Context, id platform.EnvironmentID, vars map[string]string) (RevisionID, error)
	Suspend(ctx context.Context, id platform.EnvironmentID, suspended bool) (RevisionID, error)
	Reconcile(ctx context.Context, id platform.EnvironmentID) (RevisionID, error)
}

// ServiceVariablesSpec is the full per-service variable state.
//
// Literals: KEY=VALUE pairs the user typed directly. Rendered into the
// container's `env:` list, overriding anything with the same name coming
// from the env-level shared bag.
//
// DatabaseRefs: KEY → {database, secretKey} bindings. Rendered as
// container env entries sourced from the CNPG-managed Secret. Resolved at
// pod startup.
//
// SharedRefs: keys from the env-level shared bag that the user explicitly
// chose to surface as variables on this service. PURE UI METADATA — the
// chart does not render these (envFrom on the shared ConfigMap makes
// every shared var available to every service unconditionally). The
// resolver uses this list to mark dashboard rows as "from shared" on
// round-trip.
type ServiceVariablesSpec struct {
	Literals     map[string]string
	DatabaseRefs map[string]DatabaseRef
	SharedRefs   []string
}

type DatabaseRef struct {
	Database string
	Key      string
}

type ServiceSpec struct {
	Image                string
	SourceURL            string
	ContextPath          string
	GitHubInstallationID int64
	StartCommand         string
	Port                 int
	Resources            Resources
	ResourceTier         platform.ResourceTier
}

type ImageProvenance struct {
	Commit        string
	CommitMessage string
	BuildID       string
}

type DatabaseSpec struct {
	Version   string
	Instances int
	Size      resource.Quantity
}

type KeyValueStoreSpec struct {
	Version  string
	Size     resource.Quantity
	Password string
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
	CPU    resource.Quantity
	Memory resource.Quantity
}
