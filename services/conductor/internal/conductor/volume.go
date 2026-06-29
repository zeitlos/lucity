package conductor

import (
	"context"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type VolumeID = platform.VolumeID
type Volume = platform.Volume

func (c *Client) Volumes(ctx context.Context, environment EnvironmentID) ([]Volume, error) {
	return c.platform.Volumes(ctx, environment)
}

func (c *Client) Volume(ctx context.Context, id VolumeID) (*Volume, error) {
	return c.platform.Volume(ctx, id)
}
