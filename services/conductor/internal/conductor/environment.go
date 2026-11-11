package conductor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

type Environment = platform.Environment
type EnvironmentID = platform.EnvironmentID

func (c *Client) Environments(ctx context.Context, projectID ProjectID) ([]Environment, error) {
	return c.platform.Environments(ctx, projectID)
}

func (c *Client) Environment(ctx context.Context, id EnvironmentID) (*Environment, error) {
	return c.platform.Environment(ctx, id)
}

const updateSettle = time.Second

// EnvironmentUpdates sends the environment, then sends it again after the
// objects behind it change, at most once per second. It ends when ctx is done
// or the environment is gone.
func (c *Client) EnvironmentUpdates(ctx context.Context, id EnvironmentID) (<-chan Environment, error) {
	environment, err := c.Environment(ctx, id)

	if err != nil {
		return nil, err
	}

	changes := c.changes.Watch(ctx, id)
	updates := make(chan Environment)

	go func() {
		defer close(updates)

		for {
			select {
			case updates <- *environment:
			case <-ctx.Done():
				return
			}

			select {
			case <-changes:
			case <-ctx.Done():
				return
			}

			select {
			case <-time.After(updateSettle):
			case <-ctx.Done():
				return
			}

			select {
			case <-changes:
			default:
			}

			environment, err = c.Environment(ctx, id)

			if err != nil {
				return
			}
		}
	}()

	return updates, nil
}

const defaultEnvironment = "development"

// CreateEnvironment provisions a new env namespace.
// The helm release materializes lazily on the first service/db/volume/
// variable write — no explicit install is needed here.
func (c *Client) CreateEnvironment(ctx context.Context, project ProjectID, id, name string) (*Environment, error) {
	existing, err := c.platform.Project(ctx, project)

	if err != nil {
		return nil, err
	}

	taken := make(map[string]string, len(existing.Environments))

	for _, environment := range existing.Environments {
		taken[environment.ID.Name] = "Another environment in this project"
	}

	id, err = pickID(taken, id, name, "environment")

	if err != nil {
		return nil, err
	}

	return c.createEnvironment(ctx, project, id, name, existing.Name)
}

func (c *Client) createEnvironment(ctx context.Context, project ProjectID, id, name, projectName string) (*Environment, error) {
	envID := platform.EnvironmentID{
		Workspace: project.Workspace,
		Project:   project.Name,
		Name:      id,
	}

	if err := c.environment.Ensure(ctx, envID); err != nil {
		return nil, err
	}

	if err := c.environment.SetName(ctx, envID, name); err != nil {
		return nil, err
	}

	if err := c.environment.SetProjectName(ctx, envID, projectName); err != nil {
		return nil, err
	}

	if c.paid(ctx, envID.Workspace) {
		if _, err := c.deployer.Environments().SetState(ctx, envID, deployer.EnvironmentState{Placement: c.config.PaidPlacement}); err != nil {
			slog.WarnContext(ctx, "install on the paid placement failed", "environment", envID, "error", err)
		}
	}

	return c.Environment(ctx, envID)
}

func (c *Client) RenameEnvironment(ctx context.Context, id platform.EnvironmentID, name string) (*Environment, error) {
	if _, err := c.platform.Environment(ctx, id); err != nil {
		return nil, err
	}

	if err := c.environment.SetName(ctx, id, name); err != nil {
		return nil, err
	}

	return c.Environment(ctx, id)
}

func (c *Client) DeleteEnvironment(ctx context.Context, environment platform.EnvironmentID) error {
	if err := c.checkEnvironmentEmpty(ctx, environment); err != nil {
		return err
	}

	if err := c.checkNotLastEnvironment(ctx, environment); err != nil {
		return err
	}

	if err := c.environment.Delete(ctx, environment); err != nil {
		return err
	}

	return nil
}

func (c *Client) checkEnvironmentEmpty(ctx context.Context, id platform.EnvironmentID) error {
	services, err := c.platform.Services(ctx, id)

	if err != nil {
		return err
	}

	if len(services) > 0 {
		return usererror.Errorf("environment %q has %d service(s); remove them first", id, len(services))
	}

	databases, err := c.platform.Databases(ctx, id)

	if err != nil {
		return err
	}

	if len(databases) > 0 {
		return usererror.Errorf("environment %q has %d database(s); remove them first", id, len(databases))
	}

	keyValueStores, err := c.platform.KeyValueStores(ctx, id)

	if err != nil {
		return err
	}

	if len(keyValueStores) > 0 {
		return usererror.Errorf("environment %q has %d key-value store(s); remove them first", id, len(keyValueStores))
	}

	volumes, err := c.platform.Volumes(ctx, id)

	if err != nil {
		return err
	}

	if len(volumes) > 0 {
		return usererror.Errorf("environment %q has %d volume(s); remove them first", id, len(volumes))
	}

	buckets, err := c.objectStorage.Buckets(ctx, id, false)

	if err != nil {
		return err
	}

	if len(buckets) > 0 {
		return usererror.Errorf("environment %q has %d bucket(s); remove them first", id, len(buckets))
	}

	return nil
}

func (c *Client) checkNotLastEnvironment(ctx context.Context, id platform.EnvironmentID) error {
	envs, err := c.platform.Environments(ctx, id.ProjectID())

	if err != nil {
		return fmt.Errorf("list project environments: %w", err)
	}

	if len(envs) <= 1 {
		return usererror.Errorf("cannot delete %q: a project must have at least one environment", id)
	}

	return nil
}
