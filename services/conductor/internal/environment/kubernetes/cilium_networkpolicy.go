package kubernetes

import (
	"context"

	pkglabels "github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const ciliumNetworkPolicyName = "lucity-allow-ingress"

var ciliumNetworkPolicyGVR = schema.GroupVersionResource{
	Group:    "cilium.io",
	Version:  "v2",
	Resource: "ciliumnetworkpolicies",
}

func (c *Client) ensureCiliumNetworkPolicy(ctx context.Context, id platform.EnvironmentID) error {
	namespace := id.Namespace()
	policy := buildCiliumNetworkPolicy(namespace)

	client := c.dyn.Resource(ciliumNetworkPolicyGVR).Namespace(namespace)

	existing, err := client.Get(ctx, ciliumNetworkPolicyName, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		if _, err := client.Create(ctx, policy, metav1.CreateOptions{}); err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	policy.SetResourceVersion(existing.GetResourceVersion())

	if _, err := client.Update(ctx, policy, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func buildCiliumNetworkPolicy(namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "cilium.io/v2",
			"kind":       "CiliumNetworkPolicy",
			"metadata": map[string]any{
				"name":      ciliumNetworkPolicyName,
				"namespace": namespace,
				"labels": map[string]any{
					pkglabels.ManagedBy: pkglabels.ManagedByLucity,
				},
			},
			"spec": map[string]any{
				"endpointSelector": map[string]any{},
				"ingress": []any{
					map[string]any{
						"fromEntities": []any{"ingress"},
					},
				},
			},
		},
	}
}
