package conductor

import (
	"context"
	"fmt"
	"time"

	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

const grpcTimeout = 10 * time.Second

type Project = platform.Project
type ProjectID = platform.ProjectID
type Deployment = platform.Deployment

type ScalingConfig struct {
	Replicas    int
	Autoscaling *AutoscalingConfig
}

type AutoscalingConfig struct {
	Enabled     bool
	MinReplicas int
	MaxReplicas int
	TargetCPU   int
}

func (c *Client) Projects(ctx context.Context, workspace string) ([]Project, error) {
	return c.platform.Projects(ctx, workspace)
}

func (c *Client) Project(ctx context.Context, id platform.ProjectID) (*Project, error) {
	return c.platform.Project(ctx, id)
}

func (c *Client) CreateProject(ctx context.Context, ws, id, name string) (*Project, error) {
	projects, err := c.platform.Projects(ctx, ws)

	if err != nil {
		return nil, err
	}

	taken := make(map[string]string, len(projects))

	for _, project := range projects {
		taken[project.ID.Name] = "Another project"
	}

	id, err = pickID(taken, id, name, "project")

	if err != nil {
		return nil, err
	}

	projectID := platform.ProjectID{
		Workspace: ws,
		Name:      id,
	}

	if _, err := c.createEnvironment(ctx, projectID, defaultEnvironment, defaultEnvironment, name); err != nil {
		return nil, fmt.Errorf("create default environment: %w", err)
	}

	return c.Project(ctx, projectID)
}

func (c *Client) RenameProject(ctx context.Context, id platform.ProjectID, name string) (*Project, error) {
	environments, err := c.platform.Environments(ctx, id)

	if err != nil {
		return nil, err
	}

	if len(environments) == 0 {
		return nil, usererror.Errorf("project with id %q not found", id)
	}

	for _, environment := range environments {
		if err := c.environment.SetProjectName(ctx, environment.ID, name); err != nil {
			return nil, err
		}
	}

	return c.Project(ctx, id)
}

func (c *Client) DeleteProject(ctx context.Context, project platform.ProjectID) error {
	envs, err := c.platform.Environments(ctx, project)

	if err != nil {
		return err
	}

	for _, env := range envs {
		if err := c.checkEnvironmentEmpty(ctx, env.ID); err != nil {
			return err
		}
	}

	for _, env := range envs {
		if err := c.environment.Delete(ctx, env.ID); err != nil {
			return err
		}
	}

	return nil
}
