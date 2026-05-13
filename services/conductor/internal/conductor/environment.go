package conductor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type EnvironmentNew = platform.Environment
type EnvironmentID = platform.EnvironmentID

type Environment struct {
	ID         string
	Name       string
	Namespace  string
	Ephemeral  bool
	SyncStatus string
	Services   []Service
	Databases  []Database
}

func (c *Client) Environments(ctx context.Context, projectID string) ([]EnvironmentNew, error) {
	id, err := platform.ParseProjectID(projectID)

	if err != nil {
		return nil, err
	}

	return c.platform.Environments(ctx, id)
}

func (c *Client) Environment(ctx context.Context, id string) (*EnvironmentNew, error) {
	envID, err := platform.ParseEnvironmentID(id)

	if err != nil {
		return nil, err
	}

	return c.platform.Environment(ctx, envID)
}

func (c *Client) CreateEnvironment(ctx context.Context, project ProjectID, name string, fromEnvironment *EnvironmentID, tier string) (*EnvironmentNew, error) {
	projectID := project.Name
	fromEnvName := ""
	if fromEnvironment != nil {
		fromEnvName = fromEnvironment.Name
	}
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	createCtx, createCancel := context.WithTimeout(ctx, grpcLongTimeout)
	defer createCancel()
	namespace, err := c.Packager.CreateEnvironment(createCtx, ws, projectID, name, fromEnvName)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Deploy the new environment via ArgoCD
	envDeployCtx, envDeployCancel := context.WithTimeout(ctx, grpcTimeout)
	defer envDeployCancel()
	if _, err := c.Deployer.DeployEnvironment(envDeployCtx, ws, projectID, name, "", namespace); err != nil {
		slog.Warn("failed to deploy environment", "project", projectID, "environment", name, "error", err)
	}

	// If PRODUCTION tier was requested, set up ResourceQuota with default allocations.
	if tier == "PRODUCTION" {
		envID, _ := platform.ParseEnvironmentID(projectID + "/" + name)
		c.SetEnvironmentResources(ctx, envID, tier, 1000, 1024, 1024)
	}

	// Trigger immediate sync so the environment deploys right away
	syncCtx, syncCancel := context.WithTimeout(ctx, grpcTimeout)
	defer syncCancel()
	if _, err := c.Deployer.SyncDeployment(syncCtx, ws, projectID, name); err != nil {
		slog.Warn("failed to trigger sync after environment create", "project", projectID, "environment", name, "error", err)
	}

	return c.Environment(ctx, projectID+"/"+name)
}

func (c *Client) DeleteEnvironment(ctx context.Context, environment platform.EnvironmentID) (bool, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}
	// Remove ArgoCD Application first (cascade deletes managed resources)
	rmCtx, rmCancel := context.WithTimeout(ctx, grpcTimeout)
	defer rmCancel()
	if err := c.Deployer.RemoveDeployment(rmCtx, ws, environment.Project, environment.Name); err != nil {
		slog.Warn("failed to remove deployment", "project", environment.Project, "environment", environment.Name, "error", err)
	}

	// Then remove from GitOps repo
	delEnvCtx, delEnvCancel := context.WithTimeout(ctx, grpcTimeout)
	defer delEnvCancel()
	if err := c.Packager.DeleteEnvironment(delEnvCtx, ws, environment.Project, environment.Name); err != nil {
		return false, fmt.Errorf("failed to delete environment: %w", err)
	}
	return true, nil
}
