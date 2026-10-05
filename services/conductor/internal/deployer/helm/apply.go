package helm

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v4/pkg/action"
	helmchart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/kube"
	"helm.sh/helm/v4/pkg/release"
	"helm.sh/helm/v4/pkg/release/common"
	"helm.sh/helm/v4/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/environment/kubernetes"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/resources"
)

// Name releases the same as the chart to produce minimal resource names.
const releaseName = "lucity-app"

func (c *Client) applyEnv(ctx context.Context, envID platform.EnvironmentID, mutate func(*values.Env) error) (deployer.RevisionID, error) {
	namespace := envID.Namespace()

	config := action.NewConfiguration()

	if err := config.Init(restGetterFor(namespace), namespace, "secret"); err != nil {
		return "", err
	}

	env, err := loadCurrent(config)

	if err != nil {
		return "", err
	}

	env.CommonLabels = values.CommonLabels(envID.Workspace, envID.Project, envID.Name)
	env.SharedVariableLabels = values.SharedVariableLabels()
	env.ImagePullSecrets = []values.PullSecret{{Name: kubernetes.PullSecretName}}
	env.Gateway = values.Gateway{Name: c.gatewayName, Namespace: c.gatewayNamespace, HTTPSListener: c.gatewayHTTPSListener, HeaderMatches: c.headerMatches}
	env.VerticalPodAutoscaler = values.VerticalPodAutoscaler{
		Enabled:      resources.VerticalPodAutoscalerUpdateMode != "",
		UpdatePolicy: values.UpdatePolicy{UpdateMode: resources.VerticalPodAutoscalerUpdateMode},
	}

	if c.backups.Enabled {
		env.Databases.BackupStore = c.backupStore(envID)
	}

	if err := mutate(env); err != nil {
		return "", err
	}

	values.EnsureBackupServerNames(env, namespace)

	if err := values.Validate(env); err != nil {
		return "", err
	}

	valsMap, err := envToMap(env)

	if err != nil {
		return "", err
	}

	rel, err := installOrUpgrade(ctx, config, releaseName, namespace, c.chart, valsMap)

	if err != nil {
		return "", err
	}

	accessor, err := release.NewAccessor(rel)

	if err != nil {
		return "", err
	}

	return deployer.RevisionID(strconv.Itoa(accessor.Version())), nil
}

func (c *Client) backupStore(envID platform.EnvironmentID) values.BackupStore {
	return values.BackupStore{
		Enabled:         true,
		Endpoint:        c.backups.Endpoint,
		DestinationPath: fmt.Sprintf("s3://%s/%s", c.backups.Bucket, envID.Workspace),
		Retention:       resources.DefaultBackupRetention,
		Schedule:        resources.DefaultBackupSchedule,
		Credentials: values.BackupCredentials{
			ExistingSecret: kubernetes.BackupSecretName,
			SecretKeys: values.BackupSecretKeys{
				AccessKeyIDKey:     kubernetes.BackupAccessKeyIDKey,
				SecretAccessKeyKey: kubernetes.BackupSecretAccessKeyKey,
				RegionKey:          kubernetes.BackupRegionKey,
			},
		},
	}
}

// installOrUpgrade dispatches to action.Install or action.Upgrade based on
// whether the release already exists. The helm SDK's Upgrade.Install field
// is purely informational — it does NOT auto-install missing releases.
// Callers must check history and route manually.
func installOrUpgrade(ctx context.Context, config *action.Configuration, releaseName, namespace string, chart *helmchart.Chart, vals map[string]any) (release.Releaser, error) {
	exists, replace, err := releaseState(config, releaseName)

	if err != nil {
		return nil, err
	}

	if !exists {
		install := action.NewInstall(config)
		install.ReleaseName = releaseName
		install.Namespace = namespace
		install.Replace = replace
		install.ServerSideApply = false
		install.WaitStrategy = kube.HookOnlyStrategy
		install.PostRenderer = releaseGuard{namespace: namespace}
		install.PostRenderStrategy = action.PostRenderStrategyCombined

		rel, err := install.RunWithContext(ctx, chart, vals)

		if err != nil {
			return nil, err
		}

		return rel, nil
	}

	upgrade := action.NewUpgrade(config)
	upgrade.Namespace = namespace
	upgrade.ServerSideApply = "false"
	upgrade.WaitStrategy = kube.HookOnlyStrategy
	upgrade.PostRenderer = releaseGuard{namespace: namespace}
	upgrade.PostRenderStrategy = action.PostRenderStrategyCombined

	rel, err := upgrade.RunWithContext(ctx, releaseName, chart, vals)

	if err != nil {
		return nil, err
	}

	return rel, nil
}

// releaseState reports whether a release exists in an installable state.
// Returns (exists, replace, err):
//   - exists=false → no release at all; call install
//   - exists=false, replace=true → prior release left as Uninstalled status;
//     install with Replace=true to clobber the dead marker
//   - exists=true → call upgrade
func releaseState(config *action.Configuration, name string) (exists, replace bool, err error) {
	history := action.NewHistory(config)
	history.Max = 1

	versions, err := history.Run(name)

	if errors.Is(err, driver.ErrReleaseNotFound) {
		return false, false, nil
	}

	if err != nil {
		return false, false, err
	}

	if len(versions) == 0 {
		return true, false, nil
	}

	latest, err := release.NewAccessor(versions[len(versions)-1])

	if err != nil {
		return false, false, err
	}

	if latest.Status() == common.StatusUninstalled.String() {
		return false, true, nil
	}

	return true, false, nil
}

func (c *Client) loadEnv(_ context.Context, envID platform.EnvironmentID) (*values.Env, error) {
	namespace := envID.Namespace()

	cfg := action.NewConfiguration()

	if err := cfg.Init(restGetterFor(namespace), namespace, "secret"); err != nil {
		return nil, err
	}

	return loadCurrent(cfg)
}

func loadCurrent(cfg *action.Configuration) (*values.Env, error) {
	get := action.NewGetValues(cfg)
	get.AllValues = true

	raw, err := get.Run(releaseName)

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

// restGetterFor returns a Helm RESTClientGetter scoped to namespace. Helm uses
// the getter's namespace as the default for rendered resources that omit one in
// their metadata; without it, those resources land in the kubeconfig's default
// namespace instead of the environment's namespace.
func restGetterFor(namespace string) genericclioptions.RESTClientGetter {
	flags := genericclioptions.NewConfigFlags(false)
	flags.Namespace = &namespace

	return flags
}
