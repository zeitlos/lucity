package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/zeitlos/lucity/pkg/deployer"
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

func (s *Server) ProvisionCustomDomain(ctx context.Context, req *deployer.ProvisionCustomDomainRequest) (*deployer.ProvisionCustomDomainResponse, error) {
	hostname := req.GetHostname()
	resourceName := customDomainResourceName(hostname)
	secretName := resourceName + "-tls"

	slog.Info("provisioning custom domain", "hostname", hostname, "resource", resourceName)

	cert := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]any{
				"name":      resourceName,
				"namespace": s.gatewayNamespace,
				"labels": map[string]any{
					"lucity.dev/custom-domain": "true",
					"lucity.dev/hostname":      hostname,
				},
			},
			"spec": map[string]any{
				"secretName": secretName,
				"issuerRef": map[string]any{
					"name": s.clusterIssuer,
					"kind": "ClusterIssuer",
				},
				"dnsNames": []any{hostname},
			},
		},
	}

	_, err := s.dynamic.Resource(certGVR).Namespace(s.gatewayNamespace).Create(ctx, cert, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("failed to create certificate for %s: %w", hostname, err)
	}
	if apierrors.IsAlreadyExists(err) {
		slog.Info("certificate already exists", "hostname", hostname)
	}

	if err := s.addGatewayCertRef(ctx, secretName); err != nil {
		return nil, fmt.Errorf("failed to add gateway cert ref for %s: %w", hostname, err)
	}

	slog.Info("custom domain provisioned", "hostname", hostname, "tls_status", "PROVISIONING")
	return &deployer.ProvisionCustomDomainResponse{TlsStatus: "PROVISIONING"}, nil
}

func (s *Server) DeleteCustomDomain(ctx context.Context, req *deployer.DeleteCustomDomainRequest) (*deployer.DeleteCustomDomainResponse, error) {
	hostname := req.GetHostname()
	resourceName := customDomainResourceName(hostname)
	secretName := resourceName + "-tls"

	slog.Info("deleting custom domain", "hostname", hostname)

	// Delete the Certificate.
	err := s.dynamic.Resource(certGVR).Namespace(s.gatewayNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to delete certificate for %s: %w", hostname, err)
	}

	// Delete the TLS Secret.
	err = s.k8s.CoreV1().Secrets(s.gatewayNamespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to delete tls secret for %s: %w", hostname, err)
	}

	if err := s.removeGatewayCertRef(ctx, secretName); err != nil {
		return nil, fmt.Errorf("failed to remove gateway cert ref for %s: %w", hostname, err)
	}

	slog.Info("custom domain deleted", "hostname", hostname)
	return &deployer.DeleteCustomDomainResponse{}, nil
}

func (s *Server) CustomDomainStatus(ctx context.Context, req *deployer.CustomDomainStatusRequest) (*deployer.CustomDomainStatusResponse, error) {
	hostname := req.GetHostname()
	resourceName := customDomainResourceName(hostname)

	cert, err := s.dynamic.Resource(certGVR).Namespace(s.gatewayNamespace).Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return &deployer.CustomDomainStatusResponse{TlsStatus: "NONE"}, nil
		}
		return nil, fmt.Errorf("failed to get certificate for %s: %w", hostname, err)
	}

	conditions, found, err := unstructured.NestedSlice(cert.Object, "status", "conditions")
	if err != nil || !found {
		return &deployer.CustomDomainStatusResponse{TlsStatus: "PROVISIONING"}, nil
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
			return &deployer.CustomDomainStatusResponse{TlsStatus: "ACTIVE"}, nil
		}

		if strings.Contains(reason, "Failed") || strings.Contains(reason, "Error") ||
			strings.Contains(message, "Failed") || strings.Contains(message, "Error") {
			return &deployer.CustomDomainStatusResponse{
				TlsStatus: "ERROR",
				Message:   fmt.Sprintf("%s: %s", reason, message),
			}, nil
		}

		return &deployer.CustomDomainStatusResponse{TlsStatus: "PROVISIONING"}, nil
	}

	return &deployer.CustomDomainStatusResponse{TlsStatus: "PROVISIONING"}, nil
}

func (s *Server) addGatewayCertRef(ctx context.Context, secretName string) error {
	gw, err := s.dynamic.Resource(gatewayGVR).Namespace(s.gatewayNamespace).Get(ctx, s.gatewayName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get gateway %s: %w", s.gatewayName, err)
	}

	listeners, found, err := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	if err != nil || !found {
		return fmt.Errorf("failed to read gateway listeners: %w", err)
	}

	for i, l := range listeners {
		listener, ok := l.(map[string]any)
		if !ok {
			continue
		}
		name, _ := listener["name"].(string)
		if name != "custom-https" {
			continue
		}

		tls, _ := listener["tls"].(map[string]any)
		if tls == nil {
			tls = map[string]any{}
			listener["tls"] = tls
		}

		certRefs, _ := tls["certificateRefs"].([]any)

		// Check if already present.
		for _, ref := range certRefs {
			r, ok := ref.(map[string]any)
			if !ok {
				continue
			}
			if r["name"] == secretName {
				return nil
			}
		}

		certRefs = append(certRefs, map[string]any{"name": secretName})
		tls["certificateRefs"] = certRefs
		listener["tls"] = tls
		listeners[i] = listener

		if err := unstructured.SetNestedSlice(gw.Object, listeners, "spec", "listeners"); err != nil {
			return fmt.Errorf("failed to set gateway listeners: %w", err)
		}

		_, err = s.dynamic.Resource(gatewayGVR).Namespace(s.gatewayNamespace).Update(ctx, gw, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update gateway %s: %w", s.gatewayName, err)
		}

		slog.Info("added cert ref to gateway", "secret", secretName, "gateway", s.gatewayName)
		return nil
	}

	return fmt.Errorf("listener 'custom-https' not found on gateway %s", s.gatewayName)
}

func (s *Server) removeGatewayCertRef(ctx context.Context, secretName string) error {
	gw, err := s.dynamic.Resource(gatewayGVR).Namespace(s.gatewayNamespace).Get(ctx, s.gatewayName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get gateway %s: %w", s.gatewayName, err)
	}

	listeners, found, err := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	if err != nil || !found {
		return fmt.Errorf("failed to read gateway listeners: %w", err)
	}

	for i, l := range listeners {
		listener, ok := l.(map[string]any)
		if !ok {
			continue
		}
		name, _ := listener["name"].(string)
		if name != "custom-https" {
			continue
		}

		tls, _ := listener["tls"].(map[string]any)
		if tls == nil {
			return nil
		}

		certRefs, _ := tls["certificateRefs"].([]any)
		filtered := make([]any, 0, len(certRefs))
		for _, ref := range certRefs {
			r, ok := ref.(map[string]any)
			if !ok {
				continue
			}
			if r["name"] != secretName {
				filtered = append(filtered, ref)
			}
		}

		tls["certificateRefs"] = filtered
		listener["tls"] = tls
		listeners[i] = listener

		if err := unstructured.SetNestedSlice(gw.Object, listeners, "spec", "listeners"); err != nil {
			return fmt.Errorf("failed to set gateway listeners: %w", err)
		}

		_, err = s.dynamic.Resource(gatewayGVR).Namespace(s.gatewayNamespace).Update(ctx, gw, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update gateway %s: %w", s.gatewayName, err)
		}

		slog.Info("removed cert ref from gateway", "secret", secretName, "gateway", s.gatewayName)
		return nil
	}

	// Listener not found is fine during deletion.
	return nil
}

// ReconcileCustomDomains ensures the Gateway's custom-https listener cert refs
// match the set of cert-manager Certificates labeled with lucity.dev/custom-domain=true.
func (s *Server) ReconcileCustomDomains(ctx context.Context) error {
	certs, err := s.dynamic.Resource(certGVR).Namespace(s.gatewayNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: "lucity.dev/custom-domain=true",
	})
	if err != nil {
		return fmt.Errorf("failed to list custom domain certificates: %w", err)
	}

	// Collect expected secret names from certificates.
	expected := make([]any, 0, len(certs.Items))
	for _, cert := range certs.Items {
		secretName, found, err := unstructured.NestedString(cert.Object, "spec", "secretName")
		if err != nil || !found {
			continue
		}
		expected = append(expected, map[string]any{"name": secretName})
	}

	gw, err := s.dynamic.Resource(gatewayGVR).Namespace(s.gatewayNamespace).Get(ctx, s.gatewayName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get gateway %s: %w", s.gatewayName, err)
	}

	listeners, found, err := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	if err != nil || !found {
		return fmt.Errorf("failed to read gateway listeners: %w", err)
	}

	updated := false
	for i, l := range listeners {
		listener, ok := l.(map[string]any)
		if !ok {
			continue
		}
		name, _ := listener["name"].(string)
		if name != "custom-https" {
			continue
		}

		tls, _ := listener["tls"].(map[string]any)
		if tls == nil {
			tls = map[string]any{}
			listener["tls"] = tls
		}

		current, _ := tls["certificateRefs"].([]any)
		if certRefsEqual(current, expected) {
			slog.Debug("custom domain cert refs already in sync", "count", len(expected))
			return nil
		}

		tls["certificateRefs"] = expected
		listener["tls"] = tls
		listeners[i] = listener
		updated = true
		break
	}

	if !updated {
		slog.Debug("no custom-https listener found on gateway, skipping reconciliation")
		return nil
	}

	if err := unstructured.SetNestedSlice(gw.Object, listeners, "spec", "listeners"); err != nil {
		return fmt.Errorf("failed to set gateway listeners: %w", err)
	}

	_, err = s.dynamic.Resource(gatewayGVR).Namespace(s.gatewayNamespace).Update(ctx, gw, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update gateway %s: %w", s.gatewayName, err)
	}

	slog.Info("reconciled custom domain cert refs", "count", len(expected), "gateway", s.gatewayName)
	return nil
}

// certRefsEqual checks if two cert ref slices contain the same secret names.
func certRefsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	names := make(map[string]bool, len(a))
	for _, ref := range a {
		r, ok := ref.(map[string]any)
		if !ok {
			continue
		}
		n, _ := r["name"].(string)
		names[n] = true
	}
	for _, ref := range b {
		r, ok := ref.(map[string]any)
		if !ok {
			continue
		}
		n, _ := r["name"].(string)
		if !names[n] {
			return false
		}
	}
	return true
}
