// Package data holds Go-native domain types shared between the
// inproc modules (packager, deployer, builder) and the handler / API
// layer. These were proto messages until we ripped out the proto
// dance for in-process communication.
//
// Vendor-neutral leaf types (typed IDs, RolloutHealth, etc.) live in
// internal/domain — this package layers richer composites on top.
package data

import "time"

// ProjectInfo is a project's GitOps-derived metadata.
type ProjectInfo struct {
	Name             string
	DisplayName      string
	GitopsRepoURL    string
	Environments     []string
	EnvironmentInfos []EnvironmentInfo
	Databases        []DatabaseInfo
	CreatedAt        time.Time
}

// EnvironmentInfo describes a single environment within a project.
type EnvironmentInfo struct {
	Name     string
	Services []ServiceInstanceInfo
}

// ServiceInstanceInfo describes a service's state in a specific
// environment. Definition fields come from base values; runtime
// fields come from the per-environment overrides.
type ServiceInstanceInfo struct {
	Name                 string
	ImageTag             string
	Domains              []string
	Image                string
	Port                 int
	Framework            string
	SourceURL            string
	ContextPath          string
	GitHubInstallationID int64
	CustomStartCommand   string
	StartCommand         string
}

// ServiceDef bundles the args needed to add a new service definition
// to a project's base values.
type ServiceDef struct {
	Name                 string
	Image                string
	Port                 int
	Framework            string
	SourceURL            string
	ContextPath          string
	GitHubInstallationID int64
	ImageTag             string
	ImagePullPolicy      string
	CustomStartCommand   string
	StartCommand         string
	Environment          string
}

// DatabaseInfo describes a PostgreSQL database from a project's
// base values.
type DatabaseInfo struct {
	Name      string
	Version   string
	Instances int
	Size      string
}

// DeploymentEntry is one row of a service's GitOps-derived deployment history.
type DeploymentEntry struct {
	ImageTag   string
	Revision   string
	DeployedAt time.Time
	Author     string
}

// DatabaseRef references a key in a CNPG-generated Kubernetes Secret.
type DatabaseRef struct {
	Database string
	Key      string
}

// ServiceRef references another service's internal URL.
type ServiceRef struct {
	Service string
}

// DeploymentStatus is the state of an environment's deployment in
// the cluster. Mirrors the GraphQL SyncStatus enum string-for-string.
type DeploymentStatus string

const (
	DeploymentStatusUnspecified DeploymentStatus = ""
	DeploymentStatusSynced      DeploymentStatus = "SYNCED"
	DeploymentStatusOutOfSync   DeploymentStatus = "OUT_OF_SYNC"
	DeploymentStatusProgressing DeploymentStatus = "PROGRESSING"
	DeploymentStatusDegraded    DeploymentStatus = "DEGRADED"
	DeploymentStatusUnknown     DeploymentStatus = "UNKNOWN"
)

// ResourceTier is a workload's resource-quota tier.
type ResourceTier string

const (
	ResourceTierUnspecified ResourceTier = ""
	ResourceTierEco         ResourceTier = "ECO"
	ResourceTierProduction  ResourceTier = "PRODUCTION"
)

// ResourceQuota is the resource ceiling for a single environment.
type ResourceQuota struct {
	Tier          ResourceTier
	CPUMillicores int
	MemoryMB      int
	DiskMB        int
}

// ResourceAllocation pairs a quota with the workspace/project/env it applies to.
type ResourceAllocation struct {
	Workspace     string
	Project       string
	Environment   string
	Tier          ResourceTier
	CPUMillicores int
	MemoryMB      int
	DiskMB        int
}

// VolumeInfo describes a persistent volume attached to a database.
type VolumeInfo struct {
	Name          string
	Size          string
	RequestedSize string
	UsedBytes     int64
	CapacityBytes int64
}

// DatabaseStatus is the runtime state of a database.
type DatabaseStatus struct {
	Ready     bool
	Instances int
	Volume    *VolumeInfo
}

// ServiceStatus is the runtime state of a service's K8s Deployment.
type ServiceStatus struct {
	Ready         bool
	Replicas      int
	ReadyReplicas int
	Scaling       *ServiceScaling
	Resources     *ServiceResources
}

// ServiceScaling is replicas + autoscaling config.
type ServiceScaling struct {
	Replicas           int
	AutoscalingEnabled bool
	MinReplicas        int
	MaxReplicas        int
	TargetCPU          int
}

// ServiceResources is CPU/memory request and limit for a single service.
type ServiceResources struct {
	CPUMillicores      int
	MemoryMB           int
	CPULimitMillicores int
	MemoryLimitMB      int
}

// AutoscalingConfig is the desired HPA shape for a service.
type AutoscalingConfig struct {
	Enabled     bool
	MinReplicas int
	MaxReplicas int
	TargetCPU   int
}

// DatabaseTable is one row of a database's table list.
type DatabaseTable struct {
	Name          string
	Schema        string
	EstimatedRows int64
	Columns       []DatabaseColumn
}

// DatabaseColumn is a single column in a DatabaseTable.
type DatabaseColumn struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
}

// DatabaseRow is one row of query results.
type DatabaseRow struct {
	Cells []DatabaseCell
}

// DatabaseCell is one cell value (rendered as string for transport).
type DatabaseCell struct {
	Value  string
	IsNull bool
}

// DatabaseTableData is a paginated table fetch result.
type DatabaseTableData struct {
	Columns            []string
	Rows               []DatabaseRow
	TotalEstimatedRows int64
}

// DatabaseQueryResult is the result of an arbitrary SQL query.
type DatabaseQueryResult struct {
	Columns      []string
	Rows         []DatabaseRow
	AffectedRows int64
}

// DatabaseCredentials are the resolved connection credentials.
type DatabaseCredentials struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
	URI      string
}

// CustomDomainStatus is the TLS provisioning state of a custom domain.
type CustomDomainStatus struct {
	TLSStatus string // NONE | PROVISIONING | ACTIVE | ERROR
	Message   string
}

// DetectedService is a service Railpack identified in a source repo.
type DetectedService struct {
	Name          string
	Provider      string
	Framework     string
	StartCommand  string
	SuggestedPort int
}

// BuildPhase is the lifecycle state of a build job.
type BuildPhase string

const (
	BuildPhaseUnspecified BuildPhase = ""
	BuildPhaseQueued      BuildPhase = "QUEUED"
	BuildPhaseCloning     BuildPhase = "CLONING"
	BuildPhaseBuilding    BuildPhase = "BUILDING"
	BuildPhasePushing     BuildPhase = "PUSHING"
	BuildPhaseSucceeded   BuildPhase = "SUCCEEDED"
	BuildPhaseFailed      BuildPhase = "FAILED"
)

// BuildStatus is the current state of a build job.
type BuildStatus struct {
	Phase    BuildPhase
	ImageRef string
	Digest   string
	Error    string
}

// UserGitHubToken is a stored OAuth token for a user.
type UserGitHubToken struct {
	Connected    bool
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}
