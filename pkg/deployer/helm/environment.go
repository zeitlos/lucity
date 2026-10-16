package helm

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/blang/semver/v4"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/platform"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/types"
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

func (e *environmentClient) Placement(ctx context.Context, id platform.EnvironmentID) (*values.Placement, error) {
	namespace := id.Namespace()

	var newest types.UID

	if releases := e.client.releases; releases != nil {
		uid, _, synced := releases.newest(namespace)

		if synced && uid == "" {
			return nil, nil
		}

		if synced {
			if placement, ok := releases.placement(namespace, uid); ok {
				return &placement, nil
			}
		}

		newest = uid
	}

	config := action.NewConfiguration()

	if err := config.Init(restGetterFor(namespace), namespace, "secret"); err != nil {
		return nil, err
	}

	env, installed, err := loadRelease(config)

	if err != nil {
		return nil, err
	}

	if !installed {
		return nil, nil
	}

	if newest != "" {
		e.client.releases.markPlacement(namespace, newest, env.Placement)
	}

	return &env.Placement, nil
}

func (e *environmentClient) Place(ctx context.Context, id platform.EnvironmentID, placement values.Placement) (deployer.RevisionID, error) {
	return e.client.applyEnv(ctx, id, func(env *values.Env) error {
		return values.SetPlacement(env, placement)
	})
}

func (e *environmentClient) Export(_ context.Context, id platform.EnvironmentID) ([]byte, error) {
	namespace := id.Namespace()

	config := action.NewConfiguration()

	if err := config.Init(restGetterFor(namespace), namespace, "secret"); err != nil {
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

	values.StripBackups(env)
	values.StripPlacement(env)
	env.Gateway.HeaderMatches = nil

	return values.Marshal(env)
}

func (e *environmentClient) Reconcile(ctx context.Context, id platform.EnvironmentID) (deployer.RevisionID, error) {
	namespace := id.Namespace()

	var newest types.UID

	if releases := e.client.releases; releases != nil {
		uid, revision, synced := releases.newest(namespace)

		if synced && uid == "" {
			return "", nil
		}

		if synced && releases.isVerified(namespace, uid) {
			return deployer.RevisionID(strconv.Itoa(revision)), nil
		}

		newest = uid
	}

	config := action.NewConfiguration()

	if err := config.Init(restGetterFor(namespace), namespace, "secret"); err != nil {
		return "", err
	}

	metadata, err := action.NewGetMetadata(config).Run(releaseName)

	if errors.Is(err, driver.ErrReleaseNotFound) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	installed, err := semver.Parse(metadata.Version)

	if err != nil {
		return "", err
	}

	updateRequired := e.client.chartVersion.GT(installed)

	if !updateRequired {
		if newest != "" {
			e.client.releases.markVerified(namespace, newest)
		}

		return deployer.RevisionID(strconv.Itoa(metadata.Revision)), nil
	}

	slog.Info("reconcile environment", "environment", id)

	return e.client.applyEnv(ctx, id, func(env *values.Env) error {
		// Empty apply to ensure a fresh install.
		return nil
	})
}

var _ deployer.EnvironmentClient = (*environmentClient)(nil)
