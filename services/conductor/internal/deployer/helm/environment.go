package helm

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/blang/semver/v4"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"helm.sh/helm/v3/pkg/action"
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

func (e *environmentClient) Export(_ context.Context, id platform.EnvironmentID) ([]byte, error) {
	namespace := id.Namespace()

	config := new(action.Configuration)

	if err := config.Init(restGetterFor(namespace), namespace, "secret", debugLog); err != nil {
		return nil, err
	}

	exists, _, err := releaseState(config, releaseName)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	env, err := loadCurrent(config)

	if err != nil {
		return nil, err
	}

	return values.Marshal(env)
}

func (e *environmentClient) Reconcile(ctx context.Context, id platform.EnvironmentID) (deployer.RevisionID, error) {
	namespace := id.Namespace()

	config := new(action.Configuration)

	if err := config.Init(restGetterFor(namespace), namespace, "secret", debugLog); err != nil {
		return "", err
	}

	metadata, err := action.NewGetMetadata(config).Run(releaseName)

	if err != nil {
		return "", err
	}

	installed, err := semver.Parse(metadata.Version)

	if err != nil {
		return "", err
	}

	updateRequired := e.client.chartVersion.GT(installed)

	if !updateRequired {
		return deployer.RevisionID(strconv.Itoa(metadata.Revision)), nil
	}

	slog.Info("reconcile environment", "environment", id)

	return e.client.applyEnv(ctx, id, func(env *values.Env) error {
		// Empty apply to ensure a fresh install.
		return nil
	})
}

var _ deployer.EnvironmentClient = (*environmentClient)(nil)
