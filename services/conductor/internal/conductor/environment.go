package conductor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Environment = platform.Environment
type EnvironmentID = platform.EnvironmentID

func (c *Client) Environments(ctx context.Context, projectID ProjectID) ([]Environment, error) {
	return c.platform.Environments(ctx, projectID)
}

func (c *Client) Environment(ctx context.Context, id EnvironmentID) (*Environment, error) {
	return c.platform.Environment(ctx, id)
}

func (c *Client) CreateEnvironment(ctx context.Context, project ProjectID, name string, fromEnvironment *EnvironmentID, tier string) (*Environment, error) {
	projectID := project.Name
	fromEnvName := ""
	if fromEnvironment != nil {
		fromEnvName = fromEnvironment.Name
	}
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	envID, err := platform.ParseEnvironmentID(project.String() + "/" + name)

	if err != nil {
		return nil, err
	}

	namespace, err := c.Packager.CreateEnvironment(ctx, ws, projectID, name, fromEnvName)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Deploy the new environment via ArgoCD
	if _, err := c.Deployer.DeployEnvironment(ctx, ws, projectID, name, "", namespace); err != nil {
		slog.Warn("failed to deploy environment", "project", projectID, "environment", name, "error", err)
	}

	// If PRODUCTION tier was requested, set up ResourceQuota with default allocations.
	if tier == "PRODUCTION" {
		envID, _ := platform.ParseEnvironmentID(projectID + "/" + name)
		c.SetEnvironmentResources(ctx, envID, tier, 1000, 1024, 1024)
	}

	// Trigger immediate sync so the environment deploys right away
	if _, err := c.Deployer.SyncDeployment(ctx, ws, projectID, name); err != nil {
		slog.Warn("failed to trigger sync after environment create", "project", projectID, "environment", name, "error", err)
	}

	return c.Environment(ctx, envID)
}

func (c *Client) DeleteEnvironment(ctx context.Context, environment platform.EnvironmentID) (bool, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}
	// Remove ArgoCD Application first (cascade deletes managed resources)
	if err := c.Deployer.RemoveDeployment(ctx, ws, environment.Project, environment.Name); err != nil {
		slog.Warn("failed to remove deployment", "project", environment.Project, "environment", environment.Name, "error", err)
	}

	// Then remove from GitOps repo
	if err := c.Packager.DeleteEnvironment(ctx, ws, environment.Project, environment.Name); err != nil {
		return false, fmt.Errorf("failed to delete environment: %w", err)
	}
	return true, nil
}
