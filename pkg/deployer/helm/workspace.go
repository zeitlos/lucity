package helm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/kube"
	"helm.sh/helm/v4/pkg/release"
	"helm.sh/helm/v4/pkg/storage/driver"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

const workspaceReleasePrefix = "lucity-workspace-"

var errAgentsDisabled = usererror.New("agents are not enabled on this platform")

type workspaceClient struct {
	client *Client
}

func workspaceReleaseName(workspace string) string {
	return workspaceReleasePrefix + workspace
}

func (c *Client) workspaceConfig() (*action.Configuration, error) {
	if c.agents.WorkspaceChart == nil || c.agents.Namespace == "" {
		return nil, errAgentsDisabled
	}

	config := action.NewConfiguration()

	if err := config.Init(restGetterFor(c.agents.Namespace), c.agents.Namespace, "secret"); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Client) applyWorkspace(ctx context.Context, workspace string, mutate func(*values.Workspace) error) error {
	config, err := c.workspaceConfig()

	if err != nil {
		return err
	}

	ws, err := loadWorkspace(config, workspace)

	if err != nil {
		return err
	}

	ws.Workspace = workspace
	ws.Gateway = values.WorkspaceGateway{Name: AgentGatewayName, Listener: AgentGatewayListener}

	if err := mutate(ws); err != nil {
		return err
	}

	if err := values.ValidateWorkspace(ws, workspace); err != nil {
		return err
	}

	data, err := values.MarshalWorkspace(ws)

	if err != nil {
		return err
	}

	var vals map[string]any

	if err := yaml.Unmarshal(data, &vals); err != nil {
		return err
	}

	_, err = installOrUpgrade(ctx, config, workspaceReleaseName(workspace), c.agents.Namespace, c.agents.WorkspaceChart, vals, workspaceKinds)

	return err
}

func loadWorkspace(config *action.Configuration, workspace string) (*values.Workspace, error) {
	get := action.NewGetValues(config)
	get.AllValues = true

	raw, err := get.Run(workspaceReleaseName(workspace))

	if errors.Is(err, driver.ErrReleaseNotFound) {
		return values.NewWorkspace(workspace), nil
	}

	if err != nil {
		return nil, err
	}

	data, err := yaml.Marshal(raw)

	if err != nil {
		return nil, err
	}

	return values.ParseWorkspace(data)
}

func (w *workspaceClient) Workspace(_ context.Context, workspace string) (*values.Workspace, error) {
	config, err := w.client.workspaceConfig()

	if err != nil {
		return nil, err
	}

	return loadWorkspace(config, workspace)
}

func (w *workspaceClient) SetModelProvider(ctx context.Context, id platform.ModelProviderID, provider values.ModelProvider) error {
	return w.client.applyWorkspace(ctx, id.Workspace, func(ws *values.Workspace) error {
		return values.SetModelProvider(ws, id.Name, provider)
	})
}

func (w *workspaceClient) DeleteModelProvider(ctx context.Context, id platform.ModelProviderID) error {
	return w.client.applyWorkspace(ctx, id.Workspace, func(ws *values.Workspace) error {
		return values.DeleteModelProvider(ws, id.Name)
	})
}

func (w *workspaceClient) SetConnector(ctx context.Context, id platform.ConnectorID, connector values.Connector) error {
	return w.client.applyWorkspace(ctx, id.Workspace, func(ws *values.Workspace) error {
		return values.SetConnector(ws, id.Name, connector)
	})
}

func (w *workspaceClient) DeleteConnector(ctx context.Context, id platform.ConnectorID) error {
	return w.client.applyWorkspace(ctx, id.Workspace, func(ws *values.Workspace) error {
		return values.DeleteConnector(ws, id.Name)
	})
}

func (w *workspaceClient) SetAgentGrant(ctx context.Context, id platform.AgentID, grant values.AgentGrant) error {
	return w.client.applyWorkspace(ctx, id.Workspace, func(ws *values.Workspace) error {
		return values.SetAgentGrant(ws, id, grant)
	})
}

func (w *workspaceClient) DeleteAgentGrant(ctx context.Context, id platform.AgentID) error {
	return w.client.applyWorkspace(ctx, id.Workspace, func(ws *values.Workspace) error {
		return values.DeleteAgentGrant(ws, id)
	})
}

// Delete uninstalls a workspace's release. A workspace without one is
// already gone, so that is not an error.
func (w *workspaceClient) Delete(_ context.Context, workspace string) error {
	config, err := w.client.workspaceConfig()

	if errors.Is(err, errAgentsDisabled) {
		return nil
	}

	if err != nil {
		return err
	}

	uninstall := action.NewUninstall(config)
	uninstall.IgnoreNotFound = true
	uninstall.WaitStrategy = kube.HookOnlyStrategy

	if _, err := uninstall.Run(workspaceReleaseName(workspace)); err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		return fmt.Errorf("uninstall workspace release: %w", err)
	}

	return nil
}

func (w *workspaceClient) Workspaces(_ context.Context) ([]string, error) {
	config, err := w.client.workspaceConfig()

	if err != nil {
		return nil, err
	}

	list := action.NewList(config)
	list.Filter = "^" + workspaceReleasePrefix

	releases, err := list.Run()

	if err != nil {
		return nil, err
	}

	workspaces := make([]string, 0, len(releases))

	for _, rel := range releases {
		accessor, err := release.NewAccessor(rel)

		if err != nil {
			return nil, err
		}

		workspaces = append(workspaces, strings.TrimPrefix(accessor.Name(), workspaceReleasePrefix))
	}

	return workspaces, nil
}

func (w *workspaceClient) Outdated(_ context.Context, workspace string) (bool, error) {
	config, err := w.client.workspaceConfig()

	if err != nil {
		return false, err
	}

	metadata, err := action.NewGetMetadata(config).Run(workspaceReleaseName(workspace))

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

	current, err := semver.Parse(w.client.agents.WorkspaceChart.Metadata.Version)

	if err != nil {
		return false, err
	}

	return current.GT(installed), nil
}

// Upgrade re-renders a workspace's release with the current chart and the
// values it already has.
func (w *workspaceClient) Upgrade(ctx context.Context, workspace string) error {
	return w.client.applyWorkspace(ctx, workspace, func(*values.Workspace) error {
		return nil
	})
}

var _ deployer.WorkspaceClient = (*workspaceClient)(nil)
