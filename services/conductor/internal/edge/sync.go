package edge

import (
	"context"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/bunny"
)

func (c *Client) Sync(ctx context.Context, workspaces map[string][]Host, routing []Host) error {
	zones, err := c.api.PullZones(ctx)

	if err != nil {
		return err
	}

	byName := make(map[string]*bunny.PullZone, len(zones))

	for i := range zones {
		byName[zones[i].Name] = &zones[i]
	}

	for workspaceID, hosts := range workspaces {
		if len(hosts) == 0 {
			continue
		}

		name := c.zoneName(workspaceID)
		zone, err := c.ensureZone(ctx, name, byName[name])

		if err != nil {
			slog.WarnContext(ctx, "edge sync: zone failed", "workspace", workspaceID, "error", err)
			continue
		}

		for _, host := range hosts {
			c.syncHostLogged(ctx, zone, host)
		}
	}

	if len(routing) == 0 {
		return nil
	}

	zone, ok := byName[c.config.ZonePrefix]

	if !ok {
		slog.ErrorContext(ctx, "edge sync: routing zone missing", "zone", c.config.ZonePrefix)
		return nil
	}

	if _, err := c.ensureZone(ctx, zone.Name, zone); err != nil {
		slog.WarnContext(ctx, "edge sync: routing zone rules failed", "error", err)
	}

	for _, host := range routing {
		if hostnameOn(zone, host.Name) == nil {
			slog.WarnContext(ctx, "edge sync: routing hostname not registered", "host", host.Name)
			continue
		}

		c.syncHostLogged(ctx, zone, host)
	}

	return nil
}

func (c *Client) syncHostLogged(ctx context.Context, zone *bunny.PullZone, host Host) {
	if err := c.syncHost(ctx, zone, host); err != nil {
		slog.WarnContext(ctx, "edge sync: host failed", "zone", zone.Name, "host", host.Name, "error", err)
	}
}

func (c *Client) syncHost(ctx context.Context, zone *bunny.PullZone, host Host) error {
	if hostnameOn(zone, host.Name) == nil {
		if err := c.api.AddHostname(ctx, zone.ID, host.Name); err != nil {
			return err
		}

		zone.Hostnames = append(zone.Hostnames, bunny.Hostname{Value: host.Name})
	}

	if host.Certificate == nil {
		return nil
	}

	return c.ensureUploaded(ctx, zone, host.Name, host.Certificate)
}
