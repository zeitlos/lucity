package kubernetes

import (
	"context"
	"errors"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func (c *Client) Deployments(ctx context.Context, serviceID string) ([]platform.Deployment, error) {
	return nil, errors.New("unimplemented")
}

func (c *Client) Deployment(ctx context.Context, id string) (*platform.Deployment, error) {
	return nil, errors.New("unimplemented")
}
