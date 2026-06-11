package kubernetes

import (
	"bytes"
	"context"
	"log/slog"

	pkglabels "github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) ensurePullSecret(ctx context.Context, id platform.EnvironmentID) error {
	if c.systemPullSecret == "" {
		return nil
	}

	namespace := id.Namespace()

	source, err := c.k8s.CoreV1().Secrets(c.systemNamespace).Get(ctx, c.systemPullSecret, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		slog.Warn("registry pull secret not found, skipping", "secret", c.systemPullSecret, "namespace", c.systemNamespace)
		return nil
	}

	if err != nil {
		return err
	}

	target := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PullSecretName,
			Namespace: namespace,
			Labels: map[string]string{
				pkglabels.ManagedBy: pkglabels.ManagedByLucity,
			},
		},
		Type: source.Type,
		Data: source.Data,
	}

	existing, err := c.k8s.CoreV1().Secrets(namespace).Get(ctx, PullSecretName, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		if _, err := c.k8s.CoreV1().Secrets(namespace).Create(ctx, target, metav1.CreateOptions{}); err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	if existing.Type == source.Type && secretDataEqual(existing.Data, source.Data) {
		return nil
	}

	existing.Type = source.Type
	existing.Data = source.Data
	existing.Labels = target.Labels

	if _, err := c.k8s.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func secretDataEqual(a, b map[string][]byte) bool {
	if len(a) != len(b) {
		return false
	}

	for k, av := range a {
		bv, ok := b[k]

		if !ok || !bytes.Equal(av, bv) {
			return false
		}
	}

	return true
}
