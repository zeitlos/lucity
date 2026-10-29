// Package domains keeps workload hostnames attached, verified and registered
// with the edge CDN.
package domains

import (
	"context"
	"log/slog"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/edge"
	"github.com/zeitlos/lucity/pkg/hostname"
	"github.com/zeitlos/lucity/pkg/platform"
)

type Client struct {
	hostname         *hostname.Client
	edge             edge.Interface
	platform         platform.Interface
	deployer         deployer.Interface
	kubernetes       kubernetes.Interface
	workloadDomain   string
	gatewayNamespace string
	wildcardSecret   string
}

func New(hostname *hostname.Client, edge edge.Interface, platform platform.Interface, deployer deployer.Interface, kubernetes kubernetes.Interface, workloadDomain, gatewayNamespace, wildcardSecret string) *Client {
	return &Client{
		hostname:         hostname,
		edge:             edge,
		platform:         platform,
		deployer:         deployer,
		kubernetes:       kubernetes,
		workloadDomain:   workloadDomain,
		gatewayNamespace: gatewayNamespace,
		wildcardSecret:   wildcardSecret,
	}
}

// Reconcile attaches custom domains whose DNS verifies and detaches those that
// stopped verifying, then syncs every attached host of the given environments
// to the edge.
func (c *Client) Reconcile(ctx context.Context, environments []platform.EnvironmentID) error {
	wildcard := c.wildcardCertificate(ctx)
	attached := map[string][]edge.Host{}
	detached := map[string][]string{}

	for _, environment := range environments {
		if err := ctx.Err(); err != nil {
			return err
		}

		hosts, direct := c.reconcileEnvironment(ctx, environment, wildcard)
		attached[environment.Workspace] = append(attached[environment.Workspace], hosts...)
		detached[environment.Workspace] = append(detached[environment.Workspace], direct...)
	}

	var routing []edge.Host

	if wildcard != nil {
		routing = []edge.Host{{Name: "*." + c.workloadDomain, Certificate: wildcard}}
	}

	return c.edge.Sync(ctx, attached, detached, routing)
}

func (c *Client) reconcileEnvironment(ctx context.Context, envID platform.EnvironmentID, wildcard *edge.Certificate) ([]edge.Host, []string) {
	services, err := c.platform.Services(ctx, envID)

	if err != nil {
		slog.Warn("reconcile domains: list services failed", "env", envID, "error", err)
		return nil, nil
	}

	var (
		attached []edge.Host
		detached []string
	)

	for _, service := range services {
		for _, endpoint := range service.Endpoints {
			host := endpoint.Host

			if host == "" || c.hostname.IsInternal(host) {
				continue
			}

			enabled := endpoint.Enabled

			if c.hostname.IsCustom(host) {
				verified, err := c.Verified(ctx, envID.Workspace, host, endpoint.RedirectTo != "")

				if err != nil {
					slog.Warn("reconcile domains: dns lookup failed", "host", host, "error", err)
				} else if verified != endpoint.Enabled {
					if _, err := c.deployer.Services().AttachDomain(ctx, service.ID, host, verified); err != nil {
						slog.Warn("reconcile domains: verify call failed", "host", host, "error", err)
					} else {
						slog.Info("reconcile domains: verification changed", "service", service.ID, "host", host, "verified", verified)
						enabled = verified
					}
				}
			}

			if !enabled {
				continue
			}

			if !c.UsesEdge(ctx, host, endpoint.RedirectTo != "") {
				detached = append(detached, host)
				continue
			}

			if edgeHost, ok := c.edgeHost(ctx, envID.Namespace(), host, wildcard); ok {
				attached = append(attached, edgeHost)
			}
		}
	}

	return attached, detached
}

// AttachPending attaches every not yet attached custom domain of the given
// workspaces whose DNS now verifies, and registers it with the edge.
func (c *Client) AttachPending(ctx context.Context, workspaces map[string]bool) error {
	pending, err := c.platform.PendingDomains(ctx)

	if err != nil {
		return err
	}

	for _, domain := range pending {
		if err := ctx.Err(); err != nil {
			return err
		}

		if !workspaces[domain.Service.Workspace] || !c.hostname.IsCustom(domain.Host) {
			continue
		}

		redirect := domain.RedirectTo != ""
		verified, err := c.Verified(ctx, domain.Service.Workspace, domain.Host, redirect)

		if err != nil {
			slog.DebugContext(ctx, "attach domains: dns lookup failed", "host", domain.Host, "error", err)
			continue
		}

		if !verified {
			continue
		}

		if _, err := c.deployer.Services().AttachDomain(ctx, domain.Service, domain.Host, true); err != nil {
			slog.WarnContext(ctx, "attach domains: attach failed", "service", domain.Service, "host", domain.Host, "error", err)
			continue
		}

		slog.InfoContext(ctx, "attach domains: domain verified and attached", "service", domain.Service, "host", domain.Host)

		if c.UsesEdge(ctx, domain.Host, redirect) {
			c.Register(ctx, domain.Service, domain.Host)
		}
	}

	return nil
}

// Verified reports whether the DNS records of a custom domain point at the
// platform.
func (c *Client) Verified(ctx context.Context, workspaceID, host string, redirect bool) (bool, error) {
	dnsStatus, err := c.hostname.DNSStatus(ctx, workspaceID, host, redirect)

	if err != nil {
		return false, err
	}

	return dnsStatus == hostname.DNSValid, nil
}

// UsesEdge reports whether traffic for host goes through the edge. Only
// redirect hosts whose apex points straight at the load balancer bypass it.
func (c *Client) UsesEdge(ctx context.Context, host string, redirect bool) bool {
	if !redirect {
		return true
	}

	direct, err := c.hostname.DirectToLoadBalancer(ctx, host)

	if err != nil {
		slog.WarnContext(ctx, "edge: dns lookup failed", "host", host, "error", err)
		return true
	}

	return !direct
}

func (c *Client) Register(ctx context.Context, serviceID platform.ServiceID, host string) {
	edgeHost, ok := c.edgeHost(ctx, serviceID.Namespace(), host, c.wildcardCertificate(ctx))

	if !ok {
		return
	}

	if err := c.edge.Register(ctx, serviceID.Workspace, edgeHost); err != nil {
		slog.WarnContext(ctx, "edge register failed", "workspace", serviceID.Workspace, "host", host, "error", err)
	}
}

func (c *Client) Unregister(ctx context.Context, workspaceID, host string) {
	if c.hostname.IsInternal(host) {
		return
	}

	if err := c.edge.Unregister(ctx, workspaceID, host); err != nil {
		slog.WarnContext(ctx, "edge unregister failed", "workspace", workspaceID, "host", host, "error", err)
	}
}

func (c *Client) DeleteZone(ctx context.Context, workspaceID string) error {
	return c.edge.DeleteZone(ctx, workspaceID)
}

// ReleaseZones deletes the edge zone of every workspace that has no route left
// and holds only platform hostnames, which the routing zone's wildcard serves
// as well. Zones are listed before routes are read: every caller creates a
// route before its zone, so a zone created in between always has its route
// seen.
func (c *Client) ReleaseZones(ctx context.Context) error {
	zones, err := c.edge.Zones(ctx)

	if err != nil || len(zones) == 0 {
		return err
	}

	routed, err := c.platform.RoutedWorkspaces(ctx)

	if err != nil {
		return err
	}

	for _, zone := range zones {
		if routed[zone.Workspace] || slices.ContainsFunc(zone.Hostnames, c.hostname.IsCustom) {
			continue
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if err := c.edge.DeleteZone(ctx, zone.Workspace); err != nil {
			slog.WarnContext(ctx, "edge: zone release failed", "workspace", zone.Workspace, "error", err)
			continue
		}

		slog.InfoContext(ctx, "edge: zone released", "workspace", zone.Workspace, "hosts", len(zone.Hostnames))
	}

	return nil
}

func (c *Client) edgeHost(ctx context.Context, namespace, host string, wildcard *edge.Certificate) (edge.Host, bool) {
	if c.hostname.IsPlatform(host) {
		if wildcard == nil {
			return edge.Host{}, false
		}

		return edge.Host{Name: host, Certificate: wildcard}, true
	}

	return edge.Host{Name: host, Certificate: c.customCertificate(ctx, namespace, host)}, true
}

func (c *Client) edgeEnabled() bool {
	_, disabled := c.edge.(edge.Disabled)

	return !disabled
}

func (c *Client) customCertificate(ctx context.Context, namespace, host string) *edge.Certificate {
	if !c.edgeEnabled() {
		return nil
	}

	certificate, err := c.certificate(ctx, namespace, values.TLSSecretName(host))

	if err != nil {
		slog.DebugContext(ctx, "edge: custom certificate unavailable", "host", host, "error", err)
		return nil
	}

	return certificate
}

func (c *Client) wildcardCertificate(ctx context.Context) *edge.Certificate {
	if !c.edgeEnabled() {
		return nil
	}

	certificate, err := c.certificate(ctx, c.gatewayNamespace, c.wildcardSecret)

	if err != nil {
		slog.WarnContext(ctx, "edge: wildcard certificate unavailable", "error", err)
		return nil
	}

	return certificate
}

func (c *Client) certificate(ctx context.Context, namespace, name string) (*edge.Certificate, error) {
	secret, err := c.kubernetes.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return &edge.Certificate{Certificate: secret.Data["tls.crt"], Key: secret.Data["tls.key"]}, nil
}
