package helm

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func (c *Client) applyEnv(ctx context.Context, envID platform.EnvironmentID, mutate func(*values.Env) error) (deployer.RevisionID, error) {
	ns := envID.Namespace()

	cfg := new(action.Configuration)

	if err := cfg.Init(c.restGetter, ns, "secret", debugLog); err != nil {
		return "", err
	}

	env, err := loadCurrent(cfg, ns)

	if err != nil {
		return "", err
	}

	env.Workspace = envID.Workspace
	env.Project = envID.Project
	env.Environment = envID.Name

	if err := mutate(env); err != nil {
		return "", err
	}

	if err := values.Validate(env); err != nil {
		return "", fmt.Errorf("validate: %w", err)
	}

	valsMap, err := envToMap(env)

	if err != nil {
		return "", err
	}

	upgrade := action.NewUpgrade(cfg)
	upgrade.Namespace = ns
	upgrade.Install = true

	rel, err := upgrade.RunWithContext(ctx, ns, c.chart, valsMap)

	if err != nil {
		return "", fmt.Errorf("helm upgrade %s: %w", ns, err)
	}

	return deployer.RevisionID(strconv.Itoa(rel.Version)), nil
}

func loadCurrent(cfg *action.Configuration, name string) (*values.Env, error) {
	get := action.NewGetValues(cfg)
	get.AllValues = true

	raw, err := get.Run(name)

	if errors.Is(err, driver.ErrReleaseNotFound) {
		return values.New(), nil
	}

	if err != nil {
		return nil, err
	}

	data, err := yaml.Marshal(raw)

	if err != nil {
		return nil, err
	}

	return values.Parse(data)
}

func envToMap(env *values.Env) (map[string]any, error) {
	data, err := values.Marshal(env)

	if err != nil {
		return nil, err
	}

	var out map[string]any

	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func debugLog(format string, v ...any) {}
