package conductor

import (
	"context"
	"log/slog"

	"github.com/zeitlos/lucity/services/conductor/internal/hostname"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type DNSRecord = hostname.DNSRecord

func (c *Client) RequiredDNSRecords(ctx context.Context, workspaceID, host string) ([]DNSRecord, error) {
	return c.hostname.DNSRecords(workspaceID, host), nil
}

func (c *Client) ReconcileDomains(ctx context.Context) error {
	workspaces, err := c.directory.Workspaces(ctx)

	if err != nil {
		return err
	}

	seen := make(map[string]struct{})

	for _, workspace := range workspaces {
		hosts, err := c.reconcileWorkspaceDomains(ctx, workspace.ID)

		if err != nil {
			slog.Warn("reconcile domains: workspace failed", "workspace", workspace.ID, "error", err)
			continue
		}

		for _, host := range hosts {
			seen[host] = struct{}{}
		}
	}

	verified := make([]string, 0, len(seen))

	for host := range seen {
		verified = append(verified, host)
	}

	if err := c.gateway.Sync(ctx, verified); err != nil {
		slog.Warn("reconcile domains: gateway sync failed", "error", err)
	}

	return nil
}

func (c *Client) reconcileWorkspaceDomains(ctx context.Context, workspaceID string) ([]string, error) {
	projects, err := c.platform.Projects(ctx, workspaceID)

	if err != nil {
		return nil, err
	}

	var verified []string

	for _, project := range projects {
		environments, err := c.platform.Environments(ctx, project.ID)

		if err != nil {
			slog.Warn("reconcile domains: list environments failed", "project", project.ID, "error", err)
			continue
		}

		for _, env := range environments {
			hosts, err := c.reconcileEnvironmentDomains(ctx, workspaceID, env.ID)

			if err != nil {
				slog.Warn("reconcile domains: env failed", "env", env.ID, "error", err)
				continue
			}

			verified = append(verified, hosts...)
		}
	}

	return verified, nil
}

func (c *Client) reconcileEnvironmentDomains(ctx context.Context, workspaceID string, envID platform.EnvironmentID) ([]string, error) {
	services, err := c.platform.Services(ctx, envID)

	if err != nil {
		return nil, err
	}

	namespace := envID.Namespace()
	var verified []string

	for _, service := range services {
		for _, endpoint := range service.Endpoints {
			host := endpoint.Host

			if host == "" {
				continue
			}

			currentVerified := endpoint.Protocol == platform.ProtocolHTTPS

			status, err := c.hostname.Status(ctx, namespace, workspaceID, host)

			if err != nil {
				slog.Warn("reconcile domains: status check failed", "host", host, "error", err)
				continue
			}

			desiredVerified := status.DNS == hostname.DNSValid

			if currentVerified != desiredVerified {
				if _, err := c.deployer.Services().VerifyDomain(ctx, service.ID, host, desiredVerified); err != nil {
					slog.Warn("reconcile domains: verify call failed", "host", host, "error", err)
					continue
				}

				slog.Info("reconcile domains: flipped",
					"service", service.ID,
					"host", host,
					"verified", desiredVerified,
				)
			}

			if desiredVerified && !c.hostname.IsPlatform(host) {
				verified = append(verified, host)
			}
		}
	}

	return verified, nil
}
