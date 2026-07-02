package conductor

import (
	"context"
	"log/slog"
	"maps"
	"slices"

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

		resolved.TLSStatus, err = c.hostname.TLSStatus(ctx, endpoint.Host)

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

	desired := make(map[string]struct{})
	complete := true

	for _, workspace := range workspaces {
		projects, err := c.platform.Projects(ctx, workspace.ID)

		if err != nil {
			slog.Warn("reconcile domains: list projects failed", "workspace", workspace.ID, "error", err)
			complete = false
			continue
		}

		for _, project := range projects {
			environments, err := c.platform.Environments(ctx, project.ID)

			if err != nil {
				slog.Warn("reconcile domains: list environments failed", "project", project.ID, "error", err)
				complete = false
				continue
			}

			for _, env := range environments {
				hosts, ok := c.reconcileEnvironmentDomains(ctx, env.ID)

				if !ok {
					complete = false
				}

				for _, host := range hosts {
					desired[host] = struct{}{}
				}
			}
		}
	}

	if err := c.gateway.Sync(ctx, slices.Collect(maps.Keys(desired)), complete); err != nil {
		slog.Warn("reconcile domains: gateway sync failed", "error", err)
	}

	return nil
}

func (c *Client) reconcileEnvironmentDomains(ctx context.Context, envID platform.EnvironmentID) ([]string, bool) {
	services, err := c.platform.Services(ctx, envID)

	if err != nil {
		slog.Warn("reconcile domains: list services failed", "env", envID, "error", err)
		return nil, false
	}

	var desired []string

	for _, service := range services {
		for _, endpoint := range service.Endpoints {
			host := endpoint.Host

			if host == "" || !c.hostname.IsCustom(host) {
				continue
			}

			enabled := endpoint.Enabled

			verified, err := c.isDomainVerified(ctx, envID.Workspace, host)

			if err != nil {
				slog.Warn("reconcile domains: dns lookup failed", "host", host, "error", err)

				if enabled {
					desired = append(desired, host)
				}

				continue
			}

			if verified != enabled {
				if _, err := c.deployer.Services().VerifyDomain(ctx, service.ID, host, verified); err != nil {
					slog.Warn("reconcile domains: verify call failed", "host", host, "error", err)
					continue
				}

				enabled = verified

				slog.Info("reconcile domains: verification changed", "service", service.ID, "host", host, "verified", verified)
			}

			if enabled {
				desired = append(desired, host)
			}
		}
	}

	return desired, true
}

func (c *Client) isDomainVerified(ctx context.Context, workspaceID, host string) (bool, error) {
	dnsStatus, err := c.hostname.DNSStatus(ctx, workspaceID, host)

	if err != nil {
		return false, err
	}

	return dnsStatus == hostname.DNSValid, nil
}
