package edge

import (
	"context"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/bunny"
)

const (
	HeaderName       = "X-Lucity-Edge"
	HeaderSecretName = "lucity-edge"
	HeaderSecretKey  = "value"
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
