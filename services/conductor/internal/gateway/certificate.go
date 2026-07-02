package gateway

import (
	"context"

	"github.com/zeitlos/lucity/pkg/labels"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (c *Client) ensureCertificate(ctx context.Context, host string) error {
	name := ResourceNameFor(host)

	_, err := c.dyn.Resource(certificateGVR).Namespace(c.gatewayNamespace).Get(ctx, name, metav1.GetOptions{})

	if err == nil {
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return err
	}

	cert := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]any{
				"name":      name,
				"namespace": c.gatewayNamespace,
				"labels": map[string]any{
					labels.ManagedBy: labels.ManagedByLucity,
				},
				"annotations": map[string]any{
					labels.Prefix + "custom-domain": host,
				},
			},
			"spec": map[string]any{
				"secretName": name + "-tls",
				"dnsNames":   []any{host},
				"issuerRef": map[string]any{
					"kind": "ClusterIssuer",
					"name": c.clusterIssuer,
				},
			},
		},
	}

	_, err = c.dyn.Resource(certificateGVR).Namespace(c.gatewayNamespace).Create(ctx, cert, metav1.CreateOptions{})

	return err
}

func (c *Client) removeCertificate(ctx context.Context, host string) error {
	err := c.dyn.Resource(certificateGVR).Namespace(c.gatewayNamespace).Delete(ctx, ResourceNameFor(host), metav1.DeleteOptions{})

	if apierrors.IsNotFound(err) {
		return nil
	}

	return err
}

func (c *Client) secretExists(ctx context.Context, name string) (bool, error) {
	_, err := c.dyn.Resource(secretGVR).Namespace(c.gatewayNamespace).Get(ctx, name, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
