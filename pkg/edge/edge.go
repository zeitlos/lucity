package edge

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/bunny"
	"github.com/zeitlos/lucity/pkg/deployer"
)

const (
	HeaderName       = deployer.EdgeHeaderName
	HeaderSecretName = deployer.EdgeHeaderSecretName
	HeaderSecretKey  = deployer.EdgeHeaderSecretKey
)

type Config struct {
	OriginURL    string
	HeaderSecret string
	ZonePrefix   string
}

type Certificate struct {
	Certificate []byte
	Key         []byte
}

type Host struct {
	Name        string
	Certificate *Certificate
}

type Zone struct {
	Workspace string
	Hostnames []string
}

type Client struct {
	api    *bunny.Client
	config Config

	addressesMu sync.Mutex
	addresses   []string
	addressesAt time.Time
}

var _ Interface = (*Client)(nil)

func New(api *bunny.Client, config Config) *Client {
	return &Client{
		api:    api,
		config: config,
	}
}

func (c *Client) Register(ctx context.Context, workspaceID string, host Host) error {
	zone, err := c.ensureZone(ctx, c.zoneName(workspaceID), nil)

	if err != nil {
		return err
	}

	return c.syncHost(ctx, zone, host)
}

func (c *Client) Unregister(ctx context.Context, workspaceID, name string) error {
	zone, err := c.api.PullZoneByName(ctx, c.zoneName(workspaceID))

	if err != nil || zone == nil {
		return err
	}

	if err := c.api.RemoveHostname(ctx, zone.ID, name); err != nil && !bunny.IsNotFound(err) {
		return err
	}

	return nil
}

// Zones returns the zone of every workspace, with the hostnames registered on
// it. The routing zone and zones with another origin, such as public buckets,
// are left out.
func (c *Client) Zones(ctx context.Context) ([]Zone, error) {
	zones, err := c.api.PullZones(ctx)

	if err != nil {
		return nil, err
	}

	var result []Zone

	for _, zone := range zones {
		workspaceID, ok := strings.CutPrefix(zone.Name, c.config.ZonePrefix+"-")

		if !ok || workspaceID == "" || zone.OriginURL != c.config.OriginURL {
			continue
		}

		var hostnames []string

		for _, hostname := range zone.Hostnames {
			if !hostname.IsSystemHostname {
				hostnames = append(hostnames, hostname.Value)
			}
		}

		result = append(result, Zone{Workspace: workspaceID, Hostnames: hostnames})
	}

	return result, nil
}

func (c *Client) DeleteZone(ctx context.Context, workspaceID string) error {
	zone, err := c.api.PullZoneByName(ctx, c.zoneName(workspaceID))

	if err != nil || zone == nil {
		return err
	}

	if err := c.api.DeletePullZone(ctx, zone.ID); err != nil && !bunny.IsNotFound(err) {
		return err
	}

	return nil
}
