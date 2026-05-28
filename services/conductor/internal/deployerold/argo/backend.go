// Package argo implements deployerold.Backend on top of Soft-serve
// (GitOps repo storage) and ArgoCD (sync + reconciliation).
//
// One ArgoCD Application exists per environment of a project; that
// Application's source is the project's GitOps repo on Soft-serve.
// Each Backend.Apply call mutates one service's slice of the
// environment's values.yaml, commits, and triggers an ArgoCD sync.
//
// Project + environment lifecycle (creating the repo, creating the
// ArgoCD Application, creating the namespace) is the handler's job —
// the Backend assumes the namespace exists and carries
// lucity.dev/{workspace,project,environment} labels.
package argo

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/services/conductor/internal/deployerold"
	"github.com/zeitlos/lucity/services/conductor/internal/deployerold/argo/argocd"
	"github.com/zeitlos/lucity/services/conductor/internal/deployerold/argo/gitops"
	"github.com/zeitlos/lucity/services/conductor/internal/domain"
	"github.com/zeitlos/lucity/services/conductor/internal/kube"
)

// Backend is the GitOps + ArgoCD implementation of deployerold.Backend.
//
// It is safe for concurrent calls on different Targets. Calls on the
// same Target may serialize on the underlying Soft-serve repo
// (cloning + committing is sequential per project).
type Backend struct {
	forge gitops.Forge
	argo  *argocd.Client
	k8s   kubernetes.Interface
}

// New constructs a Backend.
func New(forge gitops.Forge, argo *argocd.Client, k8s kubernetes.Interface) *Backend {
	return &Backend{forge: forge, argo: argo, k8s: k8s}
}

// Statically assert Backend implements deployerold.Backend.
var _ deployerold.Backend = (*Backend)(nil)

// ----------------------------------------------------------------------------
// Apply
// ----------------------------------------------------------------------------

// Apply records and applies the desired Spec to the Target. It clones
// the project's GitOps repo, mutates the target service's slice of the
// environment values, commits, and triggers an ArgoCD sync.
//
// If the service does not yet exist in the environment, AddService is
// called instead of UpdateImageTag. Scaling settings (Spec.Scale) are
// always applied. Resource limits, environment variables, and probes
// from Spec are not yet plumbed through — phase 3 handler refactor
// will extend the gitops Repository interface to accept them as part
// of a single SetService call.
func (b *Backend) Apply(ctx context.Context, t deployerold.Target, spec deployerold.Spec) (deployerold.DeploymentID, error) {
	nctx, err := kube.ResolveNamespace(ctx, b.k8s, t.Namespace)
	if err != nil {
		return "", fmt.Errorf("resolve namespace: %w", err)
	}

	repo, cleanup, err := b.openRepo(ctx, nctx)
	if err != nil {
		return "", err
	}
	defer cleanup()

	env := string(nctx.Environment)
	serviceName := t.Name

	exists, err := serviceExists(ctx, repo, env, serviceName)
	if err != nil {
		return "", fmt.Errorf("look up service: %w", err)
	}

	if !exists {
		def := specToServiceDef(serviceName, spec)
		if err := repo.AddService(ctx, env, def); err != nil {
			return "", fmt.Errorf("add service: %w", err)
		}
	} else if spec.Image.Tag != "" {
		if err := repo.UpdateImageTag(ctx, env, serviceName, spec.Image.Tag, spec.Image.Digest, "deploy"); err != nil {
			return "", fmt.Errorf("update image tag: %w", err)
		}
	}

	auto := autoscalingFromSpec(spec.Scale)
	if err := repo.SetServiceScaling(ctx, env, serviceName, spec.Scale.Min, auto); err != nil {
		return "", fmt.Errorf("set scaling: %w", err)
	}

	if spec.Type == domain.WorkloadCron && spec.Cron == "" {
		return "", fmt.Errorf("workload type cron requires Cron schedule")
	}

	appName := gitops.NamespaceFor(string(nctx.Workspace), string(nctx.Project), env)
	if _, err := b.argo.SyncApplication(ctx, appName); err != nil {
		return "", fmt.Errorf("argocd sync %s: %w", appName, err)
	}

	// Phase 2b: DeploymentID is a synthetic timestamp identifier. Phase 3
	// will plumb the actual git commit SHA through the gitops Repository
	// interface so Rollback can address it.
	return deployerold.DeploymentID(fmt.Sprintf("%s@%d", appName, time.Now().UnixNano())), nil
}

// ----------------------------------------------------------------------------
// Get
// ----------------------------------------------------------------------------

// Get returns the currently-applied Spec for the Target. Returns
// deployerold.ErrNotFound when the service does not exist in the
// environment values.
func (b *Backend) Get(ctx context.Context, t deployerold.Target) (deployerold.Spec, error) {
	nctx, err := kube.ResolveNamespace(ctx, b.k8s, t.Namespace)
	if err != nil {
		return deployerold.Spec{}, fmt.Errorf("resolve namespace: %w", err)
	}

	repo, cleanup, err := b.openRepo(ctx, nctx)
	if err != nil {
		return deployerold.Spec{}, err
	}
	defer cleanup()

	services, err := repo.EnvironmentServices(ctx, string(nctx.Environment))
	if err != nil {
		return deployerold.Spec{}, fmt.Errorf("read environment services: %w", err)
	}

	for _, s := range services {
		if s.Name == t.Name {
			return serviceMetaToSpec(s), nil
		}
	}
	return deployerold.Spec{}, deployerold.ErrNotFound
}

// ----------------------------------------------------------------------------
// Remove
// ----------------------------------------------------------------------------

// Remove deletes the service from the environment's values and
// triggers an ArgoCD sync to converge the cluster. Idempotent: if the
// service is not present, returns nil.
func (b *Backend) Remove(ctx context.Context, t deployerold.Target) error {
	nctx, err := kube.ResolveNamespace(ctx, b.k8s, t.Namespace)
	if err != nil {
		// Namespace gone is fine for Remove.
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("resolve namespace: %w", err)
	}

	repo, cleanup, err := b.openRepo(ctx, nctx)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := repo.RemoveService(ctx, string(nctx.Environment), t.Name); err != nil {
		// RemoveService is idempotent on the gitops side already.
		return fmt.Errorf("remove service: %w", err)
	}

	appName := gitops.NamespaceFor(string(nctx.Workspace), string(nctx.Project), string(nctx.Environment))
	if _, err := b.argo.SyncApplication(ctx, appName); err != nil {
		return fmt.Errorf("argocd sync after remove: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Status
// ----------------------------------------------------------------------------

// Status aggregates the ArgoCD Application's sync + health status with
// per-Pod readiness in the workload namespace. The ArgoCD signal
// covers manifest sync drift; the Pod signal covers actual workload
// health (CrashLoopBackOff, ImagePullBackOff, etc.). The combined
// shape is collapsed into a single domain.RolloutHealth.
func (b *Backend) Status(ctx context.Context, t deployerold.Target) (deployerold.Status, error) {
	nctx, err := kube.ResolveNamespace(ctx, b.k8s, t.Namespace)
	if err != nil {
		return deployerold.Status{}, fmt.Errorf("resolve namespace: %w", err)
	}

	appName := gitops.NamespaceFor(string(nctx.Workspace), string(nctx.Project), string(nctx.Environment))
	app, err := b.argo.Application(ctx, appName)
	if err != nil {
		return deployerold.Status{}, fmt.Errorf("get argocd app %s: %w", appName, err)
	}

	pods, err := b.k8s.CoreV1().Pods(t.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", t.Name),
	})
	if err != nil {
		return deployerold.Status{}, fmt.Errorf("list pods: %w", err)
	}

	return aggregateStatus(app, pods.Items), nil
}

// ----------------------------------------------------------------------------
// History
// ----------------------------------------------------------------------------

// History returns past deploy events for the target, derived from the
// project's GitOps repo commit log. Newest first.
//
// Specs returned in History entries are NOT populated in this phase
// — DeploymentEntry from gitops only carries the image tag. Phase 3
// will plumb full Spec snapshots through if the API needs it.
func (b *Backend) History(ctx context.Context, t deployerold.Target) ([]deployerold.Deployment, error) {
	nctx, err := kube.ResolveNamespace(ctx, b.k8s, t.Namespace)
	if err != nil {
		return nil, fmt.Errorf("resolve namespace: %w", err)
	}

	repo, cleanup, err := b.openRepo(ctx, nctx)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	entries, err := repo.DeploymentHistory(ctx, string(nctx.Environment), t.Name)
	if err != nil {
		return nil, fmt.Errorf("deployment history: %w", err)
	}

	return entriesToDeployments(t, entries), nil
}

// ----------------------------------------------------------------------------
// Rollback
// ----------------------------------------------------------------------------

// Rollback re-applies a prior Deployment.
//
// Phase 2b: not yet implemented. The gitops repo carries history but
// has no first-class "revert this commit and sync" operation today;
// the existing rollback behavior is implemented at the deployer-grpc
// layer by reading the prior tag and re-issuing UpdateImageTag with a
// "rollback" commit prefix. Phase 3 will lift that into Backend.
func (b *Backend) Rollback(ctx context.Context, t deployerold.Target, id deployerold.DeploymentID) (deployerold.DeploymentID, error) {
	return "", fmt.Errorf("argo backend: Rollback not implemented in phase 2b")
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// openRepo locates the project's GitOps repo, clones it, and tags the
// returned Repository with its workspace so ProjectName resolves
// correctly. The cleanup func MUST be called by the caller (defer).
func (b *Backend) openRepo(ctx context.Context, nctx kube.NamespaceContext) (gitops.Repository, func(), error) {
	entry, err := b.forge.Repo(ctx, string(nctx.Project), string(nctx.Workspace))
	if err != nil {
		return nil, func() {}, fmt.Errorf("locate gitops repo for %s/%s: %w", nctx.Workspace, nctx.Project, err)
	}
	repo, err := b.forge.CloneRepo(ctx, entry.HTTPURL)
	if err != nil {
		return nil, func() {}, fmt.Errorf("clone gitops repo %s: %w", entry.HTTPURL, err)
	}
	repo.SetWorkspace(string(nctx.Workspace))
	return repo, repo.Cleanup, nil
}

// serviceExists checks whether a service is already present in the
// environment values.
func serviceExists(ctx context.Context, repo gitops.Repository, env, name string) (bool, error) {
	services, err := repo.EnvironmentServices(ctx, env)
	if err != nil {
		return false, err
	}
	for _, s := range services {
		if s.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// specToServiceDef builds a gitops.ServiceDef for a fresh service
// being added to an environment. Only the fields representable in
// today's chart values are populated.
func specToServiceDef(name string, spec deployerold.Spec) gitops.ServiceDef {
	port := 0
	if len(spec.Ports) > 0 {
		port = spec.Ports[0].Number
	}
	return gitops.ServiceDef{
		Name:               name,
		Image:              spec.Image.Repository,
		ImageTag:           spec.Image.Tag,
		Port:               port,
		CustomStartCommand: "", // not yet plumbed through Spec
	}
}

// autoscalingFromSpec builds a gitops.AutoscalingConfig from a Spec's
// scaling intent. Returns nil when the workload is fixed-replica
// (Min == Max), which signals the chart not to install an HPA.
func autoscalingFromSpec(scale deployerold.ScaleSpec) *gitops.AutoscalingConfig {
	if scale.Max <= scale.Min {
		return nil
	}
	cfg := &gitops.AutoscalingConfig{
		Enabled:     true,
		MinReplicas: scale.Min,
		MaxReplicas: scale.Max,
	}
	if scale.TargetCPU != nil {
		cfg.TargetCPU = *scale.TargetCPU
	}
	return cfg
}

// serviceMetaToSpec converts the gitops view of an environment's
// service back into a Spec. Source-level fields (framework, source
// URL) are dropped — they are project metadata, not desired state.
func serviceMetaToSpec(m gitops.ServiceInstanceMeta) deployerold.Spec {
	spec := deployerold.Spec{
		Image: domain.ImageRef{
			Repository: m.Image,
			Tag:        m.ImageTag,
		},
		Type: domain.WorkloadWeb,
	}
	if m.Port != 0 {
		spec.Ports = []deployerold.Port{{Number: m.Port}}
	}
	return spec
}

// entriesToDeployments converts gitops DeploymentEntry log entries
// into Backend Deployment records. Status is inferred from the
// commit message prefix the gitops layer captured (deploy / rollback
// / promote all become "Succeeded" — failures aren't recorded in
// git history).
func entriesToDeployments(t deployerold.Target, entries []gitops.DeploymentEntry) []deployerold.Deployment {
	out := make([]deployerold.Deployment, 0, len(entries))
	for _, e := range entries {
		out = append(out, deployerold.Deployment{
			ID:         deployerold.DeploymentID(e.Revision),
			Target:     t,
			Spec:       deployerold.Spec{Image: domain.ImageRef{Tag: e.ImageTag}},
			Status:     domain.DeploymentSucceeded,
			DeployedAt: e.Timestamp,
			DeployedBy: e.Author,
		})
	}
	return out
}

// aggregateStatus collapses the ArgoCD Application + Pod state into
// the small abstract Status the Backend interface returns.
func aggregateStatus(app *argocd.Application, pods []corev1.Pod) deployerold.Status {
	st := deployerold.Status{
		Health: argoToHealth(app),
	}
	for _, p := range pods {
		st.Pods = append(st.Pods, podSummary(p))
		if isPodReady(p) {
			st.ReadyReplicas++
		}
	}
	st.DesiredReplicas = len(pods)

	// Pod-level signals can override the ArgoCD-level health: a Synced
	// Application with crashing pods is Degraded, not Healthy.
	for _, p := range st.Pods {
		if p.Reason == "CrashLoopBackOff" || p.Reason == "ImagePullBackOff" || p.Reason == "ErrImagePull" {
			st.Health = domain.HealthDegraded
			st.Message = fmt.Sprintf("pod %s: %s", p.Name, p.Reason)
			return st
		}
	}
	if st.Health == domain.HealthHealthy && st.DesiredReplicas > 0 && st.ReadyReplicas < st.DesiredReplicas {
		st.Health = domain.HealthProgressing
	}
	return st
}

// argoToHealth maps ArgoCD's Health + Sync into the abstract
// RolloutHealth values exposed by the Backend.
func argoToHealth(app *argocd.Application) domain.RolloutHealth {
	if app == nil {
		return domain.HealthUnknown
	}
	switch strings.ToLower(app.Status.Health.Status) {
	case "healthy":
		// "Synced + Healthy" is the only state that maps to Healthy.
		// "OutOfSync + Healthy" means a deploy is pending.
		if strings.EqualFold(app.Status.Sync.Status, "Synced") {
			return domain.HealthHealthy
		}
		return domain.HealthProgressing
	case "progressing":
		return domain.HealthProgressing
	case "degraded":
		return domain.HealthDegraded
	case "missing":
		return domain.HealthDegraded
	default:
		return domain.HealthUnknown
	}
}

// podSummary builds a small per-Pod status from the K8s pod object,
// surfacing the most useful failure reason if present.
func podSummary(p corev1.Pod) deployerold.PodStatus {
	ps := deployerold.PodStatus{
		Name:  p.Name,
		Phase: string(p.Status.Phase),
		Ready: isPodReady(p),
	}
	for _, cs := range p.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			ps.Reason = cs.State.Waiting.Reason
			ps.Message = cs.State.Waiting.Message
			break
		}
		if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" && cs.State.Terminated.ExitCode != 0 {
			ps.Reason = cs.State.Terminated.Reason
			ps.Message = cs.State.Terminated.Message
			break
		}
	}
	return ps
}

// isPodReady returns true when the Pod's Ready condition is True.
func isPodReady(p corev1.Pod) bool {
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
