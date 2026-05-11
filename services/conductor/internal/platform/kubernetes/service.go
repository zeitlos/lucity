package kubernetes

import (
	"context"
	"errors"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func (c *Client) Services(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Service, error) {
	return nil, errors.New("unimplemented")
}

func (c *Client) Service(ctx context.Context, id string) (*platform.Service, error) {
	return nil, errors.New("unimplemented")
}
