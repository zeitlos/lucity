package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	pkglabels "github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/platform"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Client) ensureNamespace(ctx context.Context, id platform.EnvironmentID) error {
	name := id.Namespace()
	desired := namespaceLabels(id)

	existing, err := c.k8s.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: desired,
			},
		}

		if _, err := c.k8s.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	if owner := existing.Labels[pkglabels.Workspace]; owner != "" && owner != id.Workspace {
		return fmt.Errorf("namespace %q is owned by workspace %q, refusing to operate for workspace %q", name, owner, id.Workspace)
	}

	if labelsMatch(existing.Labels, desired) {
		return nil
	}

	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}

	for k, v := range desired {
		existing.Labels[k] = v
	}

	if _, err := c.k8s.CoreV1().Namespaces().Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *Client) SetName(ctx context.Context, id platform.EnvironmentID, name string) error {
	return c.annotateNamespace(ctx, id, pkglabels.EnvironmentName, name)
}

func (c *Client) SetProjectName(ctx context.Context, id platform.EnvironmentID, name string) error {
	return c.annotateNamespace(ctx, id, pkglabels.ProjectName, name)
}

func (c *Client) annotateNamespace(ctx context.Context, id platform.EnvironmentID, key, value string) error {
	patch, err := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]string{key: value},
		},
	})

	if err != nil {
		return err
	}

	_, err = c.k8s.CoreV1().Namespaces().Patch(ctx, id.Namespace(), types.MergePatchType, patch, metav1.PatchOptions{})

	return err
}

func (c *Client) deleteNamespace(ctx context.Context, id platform.EnvironmentID) error {
	name := id.Namespace()

	if err := c.k8s.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}

func namespaceLabels(id platform.EnvironmentID) map[string]string {
	return map[string]string{
		pkglabels.Workspace:   id.Workspace,
		pkglabels.Project:     id.Project,
		pkglabels.Environment: id.Name,
		pkglabels.ManagedBy:   pkglabels.ManagedByLucity,

		"pod-security.kubernetes.io/enforce":         "baseline",
		"pod-security.kubernetes.io/enforce-version": "latest",
		"pod-security.kubernetes.io/warn":            "baseline",
		"pod-security.kubernetes.io/warn-version":    "latest",
	}
}

func labelsMatch(have, want map[string]string) bool {
	for k, v := range want {
		if have[k] != v {
			return false
		}
	}

	return true
}
