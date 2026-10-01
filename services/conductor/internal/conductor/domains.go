package conductor

import (
	"context"
	"log/slog"

	"github.com/zeitlos/lucity/services/conductor/internal/edge"
	"github.com/zeitlos/lucity/services/conductor/internal/hostname"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Endpoint struct {
	Host               string
	Port               int
	Protocol           platform.Protocol
	RedirectTo         string
	RequiredDNSRecords []hostname.DNSRecord
	DNSStatus          hostname.DNSStatus
	TLSStatus          hostname.TLSStatus
	Type               EndpointType
}

type EndpointType = string

const (
	InternalEndpoint     EndpointType = "internal"
	PlatformEndpoint     EndpointType = "platform"
	CustomDomainEndpoint EndpointType = "custom"
)

func (c *Client) ResolveEndpoints(ctx context.Context, serviceID ServiceID, endpoints []platform.Endpoint) ([]Endpoint, error) {
	result := make([]Endpoint, 0, len(endpoints))

	for _, endpoint := range endpoints {
		var err error
		resolved := Endpoint{
			Host:               endpoint.Host,
			Port:               endpoint.Port,
			Protocol:           endpoint.Protocol,
			RedirectTo:         endpoint.RedirectTo,
			RequiredDNSRecords: c.hostname.DNSRecords(serviceID.Workspace, endpoint.Host, endpoint.RedirectTo != ""),
		}

		resolved.DNSStatus, err = c.hostname.DNSStatus(ctx, serviceID.Workspace, endpoint.Host)

		if err != nil {
			slog.ErrorContext(ctx, "failed to lookup dns status", "error", err, "service", serviceID.String(), "host", endpoint.Host)
		}

		resolved.TLSStatus, err = c.hostname.TLSStatus(ctx, serviceID.Namespace(), endpoint.Host)

		if err != nil {
			slog.ErrorContext(ctx, "failed to lookup tls status", "error", err, "service", serviceID.String(), "host", endpoint.Host)
		}

		if c.hostname.IsInternal(endpoint.Host) {
			resolved.Type = InternalEndpoint
		} else if c.hostname.IsPlatform(endpoint.Host) {
			resolved.Type = PlatformEndpoint
		} else {
			resolved.Type = CustomDomainEndpoint
		}

		result = append(result, resolved)
	}

	return result, nil
}

func (c *Client) ReconcileDomains(ctx context.Context) error {
	workspaces, err := c.directory.Workspaces(ctx)

	if err != nil {
		return err
	}

	wildcard := c.wildcardCertificate(ctx)
	attached := map[string][]edge.Host{}

	for _, workspace := range workspaces {
		projects, err := c.platform.Projects(ctx, workspace.ID)

		if err != nil {
			slog.Warn("reconcile domains: list projects failed", "workspace", workspace.ID, "error", err)
			continue
		}

		for _, project := range projects {
			environments, err := c.platform.Environments(ctx, project.ID)

			if err != nil {
				slog.Warn("reconcile domains: list environments failed", "project", project.ID, "error", err)
				continue
			}

			for _, env := range environments {
				attached[workspace.ID] = append(attached[workspace.ID], c.reconcileEnvironmentDomains(ctx, env.ID, wildcard)...)
			}
		}
	}

	var routing []edge.Host

	if wildcard != nil {
		routing = []edge.Host{{Name: "*." + c.config.WorkloadDomain, Certificate: wildcard}}
	}

	return c.edge.Sync(ctx, attached, routing)
}

func (c *Client) reconcileEnvironmentDomains(ctx context.Context, envID platform.EnvironmentID, wildcard *edge.Certificate) []edge.Host {
	services, err := c.platform.Services(ctx, envID)

	if err != nil {
		slog.Warn("reconcile domains: list services failed", "env", envID, "error", err)
		return nil
	}

	var attached []edge.Host

	for _, service := range services {
		for _, endpoint := range service.Endpoints {
			host := endpoint.Host

			if host == "" || c.hostname.IsInternal(host) {
				continue
			}

			enabled := endpoint.Enabled

			if c.hostname.IsCustom(host) {
				verified, err := c.isDomainVerified(ctx, envID.Workspace, host)

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

			if edgeHost, ok := c.edgeHost(ctx, envID.Namespace(), host, wildcard); ok {
				attached = append(attached, edgeHost)
			}
		}
	}

	return attached
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

func (c *Client) customCertificate(ctx context.Context, namespace, host string) *edge.Certificate {
	if c.config.CustomCertificate == nil {
		return nil
	}

	certificate, err := c.config.CustomCertificate(ctx, namespace, host)

	if err != nil {
		slog.DebugContext(ctx, "edge: custom certificate unavailable", "host", host, "error", err)
		return nil
	}

	return certificate
}

func (c *Client) wildcardCertificate(ctx context.Context) *edge.Certificate {
	if c.config.WildcardCertificate == nil {
		return nil
	}

	certificate, err := c.config.WildcardCertificate(ctx)

	if err != nil {
		slog.WarnContext(ctx, "edge: wildcard certificate unavailable", "error", err)
		return nil
	}

	return certificate
}

func (c *Client) registerEdge(ctx context.Context, serviceID platform.ServiceID, host string) {
	edgeHost, ok := c.edgeHost(ctx, serviceID.Namespace(), host, c.wildcardCertificate(ctx))

	if !ok {
		return
	}

	if err := c.edge.Register(ctx, serviceID.Workspace, edgeHost); err != nil {
		slog.WarnContext(ctx, "edge register failed", "workspace", serviceID.Workspace, "host", host, "error", err)
	}
}

func (c *Client) unregisterEdge(ctx context.Context, workspaceID, host string) {
	if c.hostname.IsInternal(host) {
		return
	}

	if err := c.edge.Unregister(ctx, workspaceID, host); err != nil {
		slog.WarnContext(ctx, "edge unregister failed", "workspace", workspaceID, "host", host, "error", err)
	}
}

func (c *Client) isDomainVerified(ctx context.Context, workspaceID, host string) (bool, error) {
	dnsStatus, err := c.hostname.DNSStatus(ctx, workspaceID, host)

	if err != nil {
		return false, err
	}

	return dnsStatus == hostname.DNSValid, nil
}
