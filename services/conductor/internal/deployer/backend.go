// Package deployer defines the abstract Backend that applies and
// observes desired state for individual deployable units. It owns the
// relationship between user intent (what should be running) and
// cluster reality (what is running).
//
// The Backend interface is deliberately atomic: it knows only about
// Target (opaque addressing) and Spec (desired state). It has no
// notion of workspaces, projects, or environments — those are Lucity
// vocabulary that lives above this layer (in internal/api/handler).
// Composition happens there; mechanism happens here.
//
// Implementations live in subpackages:
//
//   - deployer/argo: Soft-serve GitOps + ArgoCD (today's only impl)
//   - deployer/helm: direct Helm upgrade (future, not yet implemented)
//
// Swapping the implementation is a single line in main.go; no
// consumer-facing types change.
package deployer

import (
	"context"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/domain"
)

// Target identifies a single deployable unit. Backends do not
// interpret Namespace or Name — they store and address them as
// opaque strings. Higher layers compose Targets from
// (workspace, project, environment, service) according to whatever
// naming convention the platform uses.
type Target struct {
	Namespace string
	Name      string
}

// Spec is the full desired state of a deployable unit. It is a
// vendor-neutral shape; backends translate it to whatever they apply
// (Helm values, ArgoCD Application spec, raw manifests).
type Spec struct {
	Image     domain.ImageRef
	Type      domain.WorkloadType
	Ports     []Port
	Env       EnvVars
	Scale     ScaleSpec
	Resources Resources
	Cron      string // populated when Type == WorkloadCron
	Health    HealthChecks
}

// EnvVars is a stable, deterministically-iterable collection of
// environment variables. Implementations should sort by key when
// rendering, so semantically-equivalent Specs produce identical output.
type EnvVars map[string]string

// Port describes a single exposed port on a workload.
type Port struct {
	Number   int
	Protocol string // "TCP" | "UDP"; default TCP when empty
	Name     string // optional; used when multiple ports are exposed
}

// ScaleSpec captures the workload's scaling intent.
//
// When Min == Max, the workload is fixed at that replica count.
// When Min < Max and TargetCPU is non-nil, the Backend should
// configure an HPA. Min == Max == 0 means "suspended".
type ScaleSpec struct {
	Min       int
	Max       int
	TargetCPU *int // optional HPA CPU target (percent)
}

// Resources are CPU + memory requests/limits expressed as Kubernetes
// resource strings (e.g. "100m", "256Mi"). Empty strings mean unset.
type Resources struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
}

// HealthChecks describes liveness/readiness probes. Either may be
// zero-valued, in which case the Backend skips that probe.
type HealthChecks struct {
	Liveness  ProbeSpec
	Readiness ProbeSpec
}

// ProbeSpec is a minimal HTTP/TCP probe definition. Path is set for
// HTTP probes; Port is required for both HTTP and TCP probes; an empty
// ProbeSpec (Port == 0) means no probe.
type ProbeSpec struct {
	Port                int
	Path                string // HTTP path; empty for TCP probes
	InitialDelaySeconds int
	PeriodSeconds       int
	TimeoutSeconds      int
}

// DeploymentID uniquely identifies one application of a Spec to a
// Target. Backends generate these (e.g. ArgoCD revision SHA, Helm
// release revision number). Treat as opaque.
type DeploymentID string

// Deployment is a single revision: one application of a Spec to a
// Target. Returned by History and identified by Rollback.
type Deployment struct {
	ID         DeploymentID
	Target     Target
	Spec       Spec
	Status     domain.DeploymentStatus
	DeployedAt time.Time
	DeployedBy string // free-form, set by the caller (e.g. "user:alice", "webhook:github")
}

// Status is the observed cluster state for a Target. It distills
// backend-specific signals (ArgoCD sync, Helm release status, pod
// conditions) into a small uniform shape.
type Status struct {
	Health          domain.RolloutHealth
	DesiredReplicas int
	ReadyReplicas   int
	Message         string // human-readable summary; empty when Healthy
	Pods            []PodStatus
}

// PodStatus is a tiny per-pod summary used to surface failing pods
// up to the API. Stable shape regardless of backend.
type PodStatus struct {
	Name    string
	Phase   string // "Pending" | "Running" | "Succeeded" | "Failed" | "Unknown"
	Ready   bool
	Reason  string // optional — e.g. "CrashLoopBackOff"
	Message string // optional — last container/Pod message
}

// Backend is the deployment engine. Implementations apply Specs to
// Targets and report observed state.
//
// Concurrency: implementations must be safe for concurrent calls on
// different Targets. Calls on the same Target may serialize internally
// (e.g. Helm release locks); callers should not rely on parallelism
// within a single Target.
type Backend interface {
	// Apply records and applies the desired Spec to the Target.
	// Returns when the change is durable (committed, queued for sync,
	// or applied — implementation choice). Observe progress via
	// Status. The returned DeploymentID identifies this revision.
	Apply(ctx context.Context, t Target, spec Spec) (DeploymentID, error)

	// Get returns the currently-applied Spec. Returns ErrNotFound when
	// the Target has never been applied.
	Get(ctx context.Context, t Target) (Spec, error)

	// Remove deletes the deployable unit entirely. Idempotent: removing
	// a Target that does not exist returns nil.
	Remove(ctx context.Context, t Target) error

	// Status reports observed cluster state for a Target. Returns
	// ErrNotFound when the Target has never been applied.
	Status(ctx context.Context, t Target) (Status, error)

	// History returns past Deployments for this Target, newest first.
	// Implementations may bound the returned slice (e.g. Helm's
	// max-history); callers should not assume unbounded history.
	History(ctx context.Context, t Target) ([]Deployment, error)

	// Rollback re-applies the Spec from a prior Deployment, identified
	// by id. Returns the new DeploymentID created by the rollback —
	// rollbacks are themselves new revisions, not erasures of history.
	Rollback(ctx context.Context, t Target, id DeploymentID) (DeploymentID, error)
}
