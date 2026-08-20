package conductor

import (
	"context"
	"log/slog"

	"github.com/zeitlos/lucity/services/conductor/internal/hostname"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Endpoint struct {
	Host               string
	Port               int
	Protocol           platform.Protocol
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

func (c *Client) Endpoints(ctx context.Context, serviceID ServiceID, endpoints []platform.Endpoint) ([]Endpoint, error) {
	result := make([]Endpoint, 0, len(endpoints))

	for _, endpoint := range endpoints {
		var err error
		resolved := Endpoint{
			Host:               endpoint.Host,
			Port:               endpoint.Port,
			Protocol:           endpoint.Protocol,
			RequiredDNSRecords: c.hostname.DNSRecords(serviceID.Workspace, endpoint.Host),
			Type:               CustomDomainEndpoint,
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
				c.reconcileEnvironmentDomains(ctx, env.ID)
			}
		}
	}

	return nil
}

func (c *Client) reconcileEnvironmentDomains(ctx context.Context, envID platform.EnvironmentID) {
	services, err := c.platform.Services(ctx, envID)

	if err != nil {
		slog.Warn("reconcile domains: list services failed", "env", envID, "error", err)
		return
	}

	for _, service := range services {
		for _, endpoint := range service.Endpoints {
			host := endpoint.Host

			if host == "" || !c.hostname.IsCustom(host) {
				continue
			}

			verified, err := c.isDomainVerified(ctx, envID.Workspace, host)

			if err != nil {
				slog.Warn("reconcile domains: dns lookup failed", "host", host, "error", err)
				continue
			}

			if verified == endpoint.Enabled {
				continue
			}

			if _, err := c.deployer.Services().AttachDomain(ctx, service.ID, host, verified); err != nil {
				slog.Warn("reconcile domains: verify call failed", "host", host, "error", err)
				continue
			}

			slog.Info("reconcile domains: verification changed", "service", service.ID, "host", host, "verified", verified)
		}
	}
}

func (c *Client) isDomainVerified(ctx context.Context, workspaceID, host string) (bool, error) {
	dnsStatus, err := c.hostname.DNSStatus(ctx, workspaceID, host)

	if err != nil {
		return false, err
	}

	return dnsStatus == hostname.DNSValid, nil
}
