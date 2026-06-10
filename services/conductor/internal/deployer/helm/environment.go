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

func (e *environmentClient) Variables(ctx context.Context, id platform.EnvironmentID) (map[string]string, error) {
	env, err := e.client.loadEnv(ctx, id)

	if err != nil {
		return nil, err
	}

	out := make(map[string]string, len(env.SharedVariables))

	for k, v := range env.SharedVariables {
		out[k] = v
	}

	return out, nil
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
