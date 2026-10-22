package edge

import (
	"context"
	"log/slog"
	"slices"

	"github.com/zeitlos/lucity/pkg/bunny"
)

func (c *Client) Sync(ctx context.Context, workspaces map[string][]Host, detached map[string][]string, routing []Host) error {
	zones, err := c.api.PullZones(ctx)

	if err != nil {
		return err
	}

	byName := make(map[string]*bunny.PullZone, len(zones))

	for i := range zones {
		byName[zones[i].Name] = &zones[i]
	}

	for workspaceID, names := range detached {
		zone := byName[c.zoneName(workspaceID)]

		if zone == nil {
			continue
		}

		for _, name := range names {
			c.removeHostLogged(ctx, zone, name)
		}
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

	zone, err := c.ensureZone(ctx, c.config.ZonePrefix, byName[c.config.ZonePrefix])

	if err != nil {
		slog.WarnContext(ctx, "edge sync: routing zone failed", "zone", c.config.ZonePrefix, "error", err)
		return nil
	}

	for _, host := range routing {
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

func (c *Client) removeHostLogged(ctx context.Context, zone *bunny.PullZone, name string) {
	if hostnameOn(zone, name) == nil {
		return
	}

	if err := c.api.RemoveHostname(ctx, zone.ID, name); err != nil && !bunny.IsNotFound(err) {
		slog.WarnContext(ctx, "edge sync: host removal failed", "zone", zone.Name, "host", name, "error", err)
		return
	}

	zone.Hostnames = slices.DeleteFunc(zone.Hostnames, func(hostname bunny.Hostname) bool {
		return hostname.Value == name
	})

	slog.InfoContext(ctx, "edge sync: host detached", "zone", zone.Name, "host", name)
}
