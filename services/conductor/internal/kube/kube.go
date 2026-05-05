// Package kube holds small Kubernetes API helpers shared across the
// conductor's internal packages: namespace lifecycle, label
// application, and the workspace+project+environment lookup that
// turns a deployer.Target's Namespace string into structured
// identifiers.
//
// This package never reads workspace from request context. Callers
// pass identifiers explicitly; helpers either set labels (for create)
// or read them back (for resolve).
package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/services/conductor/internal/domain"
)

// Standard label keys applied to namespaces hosting user workloads.
// Backends rely on these to discover ownership when they only have a
// Target.Namespace string in hand.
const (
	LabelWorkspace   = "lucity.dev/workspace"
	LabelProject     = "lucity.dev/project"
	LabelEnvironment = "lucity.dev/environment"
	LabelEphemeral   = "lucity.dev/ephemeral"
)

// NamespaceContext is the structured identity of a workload namespace,
// recovered from its labels.
type NamespaceContext struct {
	Name        string
	Workspace   domain.WorkspaceID
	Project     domain.ProjectID
	Environment domain.EnvName
	Ephemeral   bool
}

// EnsureNamespace creates the namespace with the given labels if it
// does not exist. If it does exist, the labels are merged onto it
// (existing labels are preserved unless they collide). Idempotent.
func EnsureNamespace(ctx context.Context, k8s kubernetes.Interface, name string, labels map[string]string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}

	_, err := k8s.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create namespace %s: %w", name, err)
	}

	// Merge labels onto the existing namespace.
	existing, err := k8s.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get namespace %s: %w", name, err)
	}
	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	changed := false
	for k, v := range labels {
		if existing.Labels[k] != v {
			existing.Labels[k] = v
			changed = true
		}
	}
	if !changed {
		return nil
	}
	if _, err := k8s.CoreV1().Namespaces().Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update namespace labels for %s: %w", name, err)
	}
	return nil
}

// ResolveNamespace returns the structured identity of the namespace.
// Returns ErrNoLabels if any of the three required labels are missing
// — callers can treat this as "not a Lucity-managed namespace".
func ResolveNamespace(ctx context.Context, k8s kubernetes.Interface, name string) (NamespaceContext, error) {
	ns, err := k8s.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return NamespaceContext{}, fmt.Errorf("get namespace %s: %w", name, err)
	}

	ws := ns.Labels[LabelWorkspace]
	project := ns.Labels[LabelProject]
	env := ns.Labels[LabelEnvironment]

	if ws == "" || project == "" || env == "" {
		return NamespaceContext{}, fmt.Errorf("namespace %s: %w", name, ErrNoLabels)
	}

	return NamespaceContext{
		Name:        name,
		Workspace:   domain.WorkspaceID(ws),
		Project:     domain.ProjectID(project),
		Environment: domain.EnvName(env),
		Ephemeral:   ns.Labels[LabelEphemeral] == "true",
	}, nil
}

// LabelsFor produces the standard label set for a workload namespace.
// ephemeral is the lucity.dev/ephemeral label, set on PR preview
// environments and similar.
func LabelsFor(ws domain.WorkspaceID, project domain.ProjectID, env domain.EnvName, ephemeral bool) map[string]string {
	labels := map[string]string{
		LabelWorkspace:   string(ws),
		LabelProject:     string(project),
		LabelEnvironment: string(env),
	}
	if ephemeral {
		labels[LabelEphemeral] = "true"
	}
	return labels
}
