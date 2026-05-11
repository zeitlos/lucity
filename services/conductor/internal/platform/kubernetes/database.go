package kubernetes

import (
	"context"
	"errors"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func (c *Client) Databases(ctx context.Context, environmentID platform.EnvironmentID) ([]platform.Database, error) {
	return nil, errors.New("unimplemented")
}

func (c *Client) Database(ctx context.Context, id string) (*platform.Database, error) {
	return nil, errors.New("unimplemented")
}
