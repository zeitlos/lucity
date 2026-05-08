package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/zeitlos/lucity/services/conductor/internal/data"
)

var certGVR = schema.GroupVersionResource{
	Group:    "cert-manager.io",
	Version:  "v1",
	Resource: "certificates",
}

var gatewayGVR = schema.GroupVersionResource{
	Group:    "gateway.networking.k8s.io",
	Version:  "v1",
	Resource: "gateways",
}

// customDomainResourceName converts a hostname to a Kubernetes resource name.
// For example, "api.example.com" becomes "custom-api-example-com".
func customDomainResourceName(hostname string) string {
	return "custom-" + strings.ReplaceAll(hostname, ".", "-")
}

// isCustomDomainListener returns true if the listener name belongs to a
// deployer-managed custom domain listener (HTTP or HTTPS).
func isCustomDomainListener(name string) bool {
	return strings.HasPrefix(name, "custom-") && name != "custom-http" && name != "custom-https"
}

// isCertReady checks if a cert-manager Certificate has Ready=True.
func isCertReady(cert unstructured.Unstructured) bool {
	conditions, found, _ := unstructured.NestedSlice(cert.Object, "status", "conditions")
	if !found {
		return false
	}
	for _, c := range conditions {
		cond, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if cond["type"] == "Ready" && cond["status"] == "True" {
			return true
		}
	}
	return false
}

// ProvisionCustomDomain creates a cert-manager Certificate and an HTTP
// listener for a custom domain. Returns the TLS status (always
// "PROVISIONING" right after; the HTTPS listener is added later by
// ReconcileCustomDomains once the cert is Ready).
func (c *Client) ProvisionCustomDomain(ctx context.Context, hostname string) (tlsStatus string, err error) {
	resourceName := customDomainResourceName(hostname)
	secretName := resourceName + "-tls"

	slog.Info("provisioning custom domain", "hostname", hostname, "resource", resourceName)

	cert := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]any{
				"name":      resourceName,
				"namespace": c.gatewayNamespace,
				"labels": map[string]any{
					"lucity.dev/custom-domain": "true",
					"lucity.dev/hostname":      hostname,
				},
			},
			"spec": map[string]any{
				"secretName": secretName,
				"issuerRef": map[string]any{
					"name": c.clusterIssuer,
					"kind": "ClusterIssuer",
				},
				"dnsNames": []any{hostname},
			},
		},
	}

	_, err = c.dynamic.Resource(certGVR).Namespace(c.gatewayNamespace).Create(ctx, cert, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return "", fmt.Errorf("failed to create certificate for %s: %w", hostname, err)
	}
	if apierrors.IsAlreadyExists(err) {
		slog.Info("certificate already exists", "hostname", hostname)
	}

	// Add the HTTP listener immediately so cert-manager's HTTP-01 solver
	// can create an HTTPRoute that matches this domain.
	if err := c.addGatewayListener(ctx, hostname, "HTTP", ""); err != nil {
		return "", fmt.Errorf("failed to add http listener for %s: %w", hostname, err)
	}

	slog.Info("custom domain provisioned", "hostname", hostname, "tls_status", "PROVISIONING")
	return "PROVISIONING", nil
}

// DeleteCustomDomain removes the cert-manager Certificate, TLS Secret,
// and Gateway listener pair for a custom domain.
func (c *Client) DeleteCustomDomain(ctx context.Context, hostname string) error {
	resourceName := customDomainResourceName(hostname)
	secretName := resourceName + "-tls"

	slog.Info("deleting custom domain", "hostname", hostname)

	if err := c.dynamic.Resource(certGVR).Namespace(c.gatewayNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete certificate for %s: %w", hostname, err)
	}
	if err := c.k8s.CoreV1().Secrets(c.gatewayNamespace).Delete(ctx, secretName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete tls secret for %s: %w", hostname, err)
	}
	if err := c.removeGatewayListener(ctx, resourceName+"-http"); err != nil {
		return fmt.Errorf("failed to remove http listener for %s: %w", hostname, err)
	}
	if err := c.removeGatewayListener(ctx, resourceName+"-https"); err != nil {
		return fmt.Errorf("failed to remove https listener for %s: %w", hostname, err)
	}

	slog.Info("custom domain deleted", "hostname", hostname)
	return nil
}

// CustomDomainStatus returns the TLS provisioning state of a custom domain.
func (c *Client) CustomDomainStatus(ctx context.Context, hostname string) (data.CustomDomainStatus, error) {
	resourceName := customDomainResourceName(hostname)

	cert, err := c.dynamic.Resource(certGVR).Namespace(c.gatewayNamespace).Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return data.CustomDomainStatus{TLSStatus: "NONE"}, nil
		}
		return data.CustomDomainStatus{}, fmt.Errorf("failed to get certificate for %s: %w", hostname, err)
	}

	conditions, found, err := unstructured.NestedSlice(cert.Object, "status", "conditions")
	if err != nil || !found {
		return data.CustomDomainStatus{TLSStatus: "PROVISIONING"}, nil
	}

	for _, c := range conditions {
		cond, ok := c.(map[string]any)
		if !ok {
			continue
		}
		condType, _ := cond["type"].(string)
		if condType != "Ready" {
			continue
		}

		condStatus, _ := cond["status"].(string)
		reason, _ := cond["reason"].(string)
		message, _ := cond["message"].(string)

		if condStatus == "True" {
			return data.CustomDomainStatus{TLSStatus: "ACTIVE"}, nil
		}
		if strings.Contains(reason, "Failed") || strings.Contains(reason, "Error") ||
			strings.Contains(message, "Failed") || strings.Contains(message, "Error") {
			return data.CustomDomainStatus{
				TLSStatus: "ERROR",
				Message:   fmt.Sprintf("%s: %s", reason, message),
			}, nil
		}
		return data.CustomDomainStatus{TLSStatus: "PROVISIONING"}, nil
	}

	return data.CustomDomainStatus{TLSStatus: "PROVISIONING"}, nil
}

// addGatewayListener adds an HTTP or HTTPS listener for a custom domain to the Gateway.
// For HTTPS, secretName must be the TLS secret name. For HTTP, secretName is ignored.
// Idempotent: returns nil if the listener already exists.
func (c *Client) addGatewayListener(ctx context.Context, hostname, protocol, secretName string) error {
	resourceName := customDomainResourceName(hostname)
	var listenerName string
	if protocol == "HTTPS" {
		listenerName = resourceName + "-https"
	} else {
		listenerName = resourceName + "-http"
	}

	gw, err := c.dynamic.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Get(ctx, c.gatewayName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get gateway %s: %w", c.gatewayName, err)
	}

	listeners, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	for _, l := range listeners {
		listener, ok := l.(map[string]any)
		if !ok {
			continue
		}
		if listener["name"] == listenerName {
			return nil // Already exists.
		}
	}

	newListener := map[string]any{
		"name":     listenerName,
		"hostname": hostname,
		"protocol": protocol,
		"port":     int64(443),
		"allowedRoutes": map[string]any{
			"namespaces": map[string]any{"from": "All"},
		},
	}
	if protocol == "HTTP" {
		newListener["port"] = int64(80)
	} else {
		newListener["tls"] = map[string]any{
			"mode": "Terminate",
			"certificateRefs": []any{
				map[string]any{"name": secretName},
			},
		}
	}

	patchOps := []map[string]any{
		{"op": "add", "path": "/spec/listeners/-", "value": newListener},
	}
	patchData, err := json.Marshal(patchOps)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	_, err = c.dynamic.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Patch(
		ctx, c.gatewayName, types.JSONPatchType, patchData, metav1.PatchOptions{
			FieldManager: "deployer",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to patch gateway %s: %w", c.gatewayName, err)
	}

	slog.Info("added gateway listener", "listener", listenerName, "hostname", hostname, "protocol", protocol)
	return nil
}

// removeGatewayListener removes a listener by name from the Gateway.
// Idempotent: returns nil if the listener doesn't exist.
func (c *Client) removeGatewayListener(ctx context.Context, listenerName string) error {
	gw, err := c.dynamic.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Get(ctx, c.gatewayName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get gateway %s: %w", c.gatewayName, err)
	}

	listeners, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	idx := -1
	for i, l := range listeners {
		listener, ok := l.(map[string]any)
		if !ok {
			continue
		}
		if listener["name"] == listenerName {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil // Not found, nothing to remove.
	}

	patchOps := []map[string]any{
		{"op": "remove", "path": fmt.Sprintf("/spec/listeners/%d", idx)},
	}
	patchData, err := json.Marshal(patchOps)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	_, err = c.dynamic.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Patch(
		ctx, c.gatewayName, types.JSONPatchType, patchData, metav1.PatchOptions{
			FieldManager: "deployer",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to patch gateway %s: %w", c.gatewayName, err)
	}

	slog.Info("removed gateway listener", "listener", listenerName)
	return nil
}

// ReconcileCustomDomains ensures the Gateway has per-domain listener pairs
// matching the set of cert-manager Certificates labeled with lucity.dev/custom-domain=true.
// HTTP listeners are created for all certs (needed for ACME challenges).
// HTTPS listeners are created only for Ready certs (Secret must exist).
func (c *Client) ReconcileCustomDomains(ctx context.Context) error {
	certs, err := c.dynamic.Resource(certGVR).Namespace(c.gatewayNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: "lucity.dev/custom-domain=true",
	})
	if err != nil {
		return fmt.Errorf("failed to list custom domain certificates: %w", err)
	}

	// Build maps of expected state.
	type certInfo struct {
		hostname   string
		secretName string
		ready      bool
	}
	certsByHostname := make(map[string]certInfo)
	for _, cert := range certs.Items {
		hostname, _ := cert.GetLabels()["lucity.dev/hostname"]
		if hostname == "" {
			continue
		}
		secretName, found, _ := unstructured.NestedString(cert.Object, "spec", "secretName")
		if !found {
			continue
		}
		certsByHostname[hostname] = certInfo{
			hostname:   hostname,
			secretName: secretName,
			ready:      isCertReady(cert),
		}
	}

	// Read current Gateway listeners.
	gw, err := c.dynamic.Resource(gatewayGVR).Namespace(c.gatewayNamespace).Get(ctx, c.gatewayName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get gateway %s: %w", c.gatewayName, err)
	}

	listeners, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")

	// Track existing custom domain listeners.
	existingHTTP := make(map[string]bool)  // hostname → exists
	existingHTTPS := make(map[string]bool) // hostname → exists
	for _, l := range listeners {
		listener, ok := l.(map[string]any)
		if !ok {
			continue
		}
		name, _ := listener["name"].(string)
		if !isCustomDomainListener(name) {
			continue
		}
		hostname, _ := listener["hostname"].(string)
		protocol, _ := listener["protocol"].(string)
		if protocol == "HTTP" {
			existingHTTP[hostname] = true
		} else if protocol == "HTTPS" {
			existingHTTPS[hostname] = true
		}
	}

	changed := false

	// Add missing listeners for certs.
	for hostname, info := range certsByHostname {
		if !existingHTTP[hostname] {
			slog.Info("reconcile: adding http listener", "hostname", hostname)
			if err := c.addGatewayListener(ctx, hostname, "HTTP", ""); err != nil {
				slog.Error("reconcile: failed to add http listener", "hostname", hostname, "error", err)
				continue
			}
			changed = true
		}
		if info.ready && !existingHTTPS[hostname] {
			slog.Info("reconcile: adding https listener", "hostname", hostname)
			if err := c.addGatewayListener(ctx, hostname, "HTTPS", info.secretName); err != nil {
				slog.Error("reconcile: failed to add https listener", "hostname", hostname, "error", err)
				continue
			}
			changed = true
		}
		if !info.ready && existingHTTPS[hostname] {
			// Cert is no longer Ready (e.g. expired). Remove HTTPS listener.
			resourceName := customDomainResourceName(hostname)
			slog.Info("reconcile: removing https listener (cert not ready)", "hostname", hostname)
			if err := c.removeGatewayListener(ctx, resourceName+"-https"); err != nil {
				slog.Error("reconcile: failed to remove https listener", "hostname", hostname, "error", err)
			}
			changed = true
		}
	}

	// Remove orphaned listeners (no matching cert).
	for hostname := range existingHTTP {
		if _, ok := certsByHostname[hostname]; !ok {
			resourceName := customDomainResourceName(hostname)
			slog.Info("reconcile: removing orphaned http listener", "hostname", hostname)
			if err := c.removeGatewayListener(ctx, resourceName+"-http"); err != nil {
				slog.Error("reconcile: failed to remove orphaned http listener", "hostname", hostname, "error", err)
			}
			changed = true
		}
	}
	for hostname := range existingHTTPS {
		if _, ok := certsByHostname[hostname]; !ok {
			resourceName := customDomainResourceName(hostname)
			slog.Info("reconcile: removing orphaned https listener", "hostname", hostname)
			if err := c.removeGatewayListener(ctx, resourceName+"-https"); err != nil {
				slog.Error("reconcile: failed to remove orphaned https listener", "hostname", hostname, "error", err)
			}
			changed = true
		}
	}

	if changed {
		slog.Info("reconciled custom domain listeners", "certs", len(certsByHostname))
	} else {
		slog.Debug("custom domain listeners in sync", "certs", len(certsByHostname))
	}
	return nil
}
