package helm

import (
	"context"
	"errors"

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

func (e *environmentClient) State(ctx context.Context, id platform.EnvironmentID) (*deployer.EnvironmentState, error) {
	namespace := id.Namespace()

	var newest types.UID

	if releases := e.client.releases; releases != nil {
		uid, _, synced := releases.newest(namespace)

		if synced && uid == "" {
			return nil, nil
		}

		if synced {
			if state, ok := releases.state(namespace, uid); ok {
				return &state, nil
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

	state := deployer.EnvironmentState{Placement: env.Placement, Suspended: env.Suspended}

	if newest != "" {
		e.client.releases.markState(namespace, newest, state)
	}

	return &state, nil
}

func (e *environmentClient) SetState(ctx context.Context, id platform.EnvironmentID, state deployer.EnvironmentState) (deployer.RevisionID, error) {
	return e.client.applyEnv(ctx, id, func(env *values.Env) error {
		if err := values.SetPlacement(env, state.Placement); err != nil {
			return err
		}

		return values.SetSuspended(env, state.Suspended)
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

func (e *environmentClient) Outdated(ctx context.Context, id platform.EnvironmentID) (bool, error) {
	namespace := id.Namespace()

	var newest types.UID

	if releases := e.client.releases; releases != nil {
		uid, _, synced := releases.newest(namespace)

		if synced && uid == "" {
			return false, nil
		}

		if synced && releases.isVerified(namespace, uid) {
			return false, nil
		}

		newest = uid
	}

	config := action.NewConfiguration()

	if err := config.Init(restGetterFor(namespace), namespace, "secret"); err != nil {
		return false, err
	}

	metadata, err := action.NewGetMetadata(config).Run(releaseName)

	if errors.Is(err, driver.ErrReleaseNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	installed, err := semver.Parse(metadata.Version)

	if err != nil {
		return false, err
	}

	if e.client.chartVersion.GT(installed) {
		return true, nil
	}

	if newest != "" {
		e.client.releases.markVerified(namespace, newest)
	}

	return false, nil
}

var _ deployer.EnvironmentClient = (*environmentClient)(nil)
