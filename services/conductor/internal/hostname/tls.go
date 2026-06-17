package hostname

import (
	"context"

	"github.com/zeitlos/lucity/services/conductor/internal/gateway"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var certificateGVR = schema.GroupVersionResource{
	Group:    "cert-manager.io",
	Version:  "v1",
	Resource: "certificates",
}

func (c *Client) TLSStatus(ctx context.Context, host string) (TLSStatus, error) {
	if c.IsInternal(host) {
		return TLSNone, nil
	}

	if c.IsPlatform(host) {
		return TLSActive, nil
	}

	cert, err := c.dyn.Resource(certificateGVR).
		Namespace(c.gatewayNamespace).
		Get(ctx, gateway.ResourceNameFor(host), metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		return TLSProvisioning, nil
	}

	if err != nil {
		return TLSError, err
	}

	conditions, _, _ := unstructured.NestedSlice(cert.Object, "status", "conditions")

	for _, raw := range conditions {
		condition, ok := raw.(map[string]any)

		if !ok {
			continue
		}

		if condition["type"] == "Ready" && condition["status"] == "True" {
			return TLSActive, nil
		}
	}

	if attempts, _, _ := unstructured.NestedInt64(cert.Object, "status", "failedIssuanceAttempts"); attempts > 0 {
		return TLSError, nil
	}

	return TLSProvisioning, nil
}
