package kubernetes

import (
	"context"
	"errors"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func (c *Client) Volumes(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Volume, error) {
	return nil, errors.New("unimplemented")
}

func (c *Client) Volume(ctx context.Context, id string) (*platform.Volume, error) {
	return nil, errors.New("unimplemented")
}
