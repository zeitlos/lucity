// Package domain holds vendor-neutral value types shared across the
// conductor's internal packages. Nothing in here imports any other
// internal package — these are leaf types meant to flow freely.
//
// "Lucity vocabulary" lives here only to the extent that everything in
// the platform agrees on it (e.g. a project has an ID, an environment
// has a name). Richer entities like a Project (with members, settings,
// etc.) live above this layer in internal/api/handler.
package domain

import "fmt"

// Identifiers. Typed strings so the compiler can catch mix-ups
// (passing an EnvName where a ProjectID is expected).
type (
	WorkspaceID string
	ProjectID   string
	EnvName     string
	ServiceName string
	UserID      string
)

// ImageRef is a typed reference to an OCI image.
//
// Repository is the full path (e.g. "registry.example.com/workspace/project/api").
// Tag is the human-readable tag. Digest, when set, pins the image to a
// specific manifest. Either Tag or Digest must be set; both is fine
// (digest wins for pulls).
type ImageRef struct {
	Repository string
	Tag        string
	Digest     string
}

func (r ImageRef) String() string {
	if r.Digest != "" {
		return fmt.Sprintf("%s@%s", r.Repository, r.Digest)
	}
	return fmt.Sprintf("%s:%s", r.Repository, r.Tag)
}

// WorkloadType describes the runtime shape of a deployable unit.
// The Backend translates this to the appropriate K8s primitive
// (Deployment for web/worker, CronJob for cron).
type WorkloadType string

const (
	WorkloadWeb    WorkloadType = "web"
	WorkloadWorker WorkloadType = "worker"
	WorkloadCron   WorkloadType = "cron"
)

// RolloutHealth is the platform-facing health summary for a workload.
// It is deliberately coarse — the Backend's richer status (pod
// conditions, replica counts, sync state) gets distilled into one of
// these values for API consumers.
type RolloutHealth string

const (
	HealthHealthy     RolloutHealth = "Healthy"
	HealthProgressing RolloutHealth = "Progressing"
	HealthDegraded    RolloutHealth = "Degraded"
	HealthUnknown     RolloutHealth = "Unknown"
)

// DeploymentStatus is the terminal state of a single Deployment
// revision. Use RolloutHealth for live workload health.
type DeploymentStatus string

const (
	DeploymentSucceeded  DeploymentStatus = "Succeeded"
	DeploymentFailed     DeploymentStatus = "Failed"
	DeploymentSuperseded DeploymentStatus = "Superseded"
)
