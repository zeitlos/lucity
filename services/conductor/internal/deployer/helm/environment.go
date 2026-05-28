package helm

import (
	"context"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type environmentClient struct {
	client *Client
}

func (e *environmentClient) SetVariables(ctx context.Context, id platform.EnvironmentID, vars map[string]string) (deployer.RevisionID, error) {
	return e.client.applyEnv(ctx, id, func(env *values.Env) error {
		return values.SetEnvironmentVariables(env, vars)
	})
}

func (e *environmentClient) Suspend(ctx context.Context, id platform.EnvironmentID, suspended bool) (deployer.RevisionID, error) {
	return e.client.applyEnv(ctx, id, func(env *values.Env) error {
		return values.SetSuspended(env, suspended)
	})
}

var _ deployer.EnvironmentClient = (*environmentClient)(nil)
