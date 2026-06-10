package conductor

import (
	"context"
	"errors"
	"fmt"

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

// CreateEnvironment provisions a new env namespace at the requested tier.
// The helm release materializes lazily on the first service/db/volume/
// variable write — no explicit install is needed here.
//
// fromEnvironment duplication is NOT yet supported on the helm-only path:
// it requires walking the source env's services/databases/volumes via
// platform.* and re-creating each via the deployer, but platform.Service
// doesn't yet expose Image and Port fields. Variables are also deferred
// (see variable.go TODO). Until those land, callers must omit
// fromEnvironment.
func (c *Client) CreateEnvironment(ctx context.Context, project ProjectID, name string, fromEnvironment *EnvironmentID, tier ResourceTier) (*Environment, error) {
	if fromEnvironment != nil {
		return nil, errors.New("creating an environment from a source is not yet supported")
	}

	envID := platform.EnvironmentID{
		Workspace: project.Workspace,
		Project:   project.Name,
		Name:      name,
	}

	if err := c.environment.Ensure(ctx, envID, tier); err != nil {
		return nil, err
	}

	return c.Environment(ctx, envID)
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
		return fmt.Errorf("environment %q has %d service(s); remove them first", id, len(services))
	}

	databases, err := c.platform.Databases(ctx, id)

	if err != nil {
		return err
	}

	if len(databases) > 0 {
		return fmt.Errorf("environment %q has %d database(s); remove them first", id, len(databases))
	}

	volumes, err := c.platform.Volumes(ctx, id)

	if err != nil {
		return err
	}

	if len(volumes) > 0 {
		return fmt.Errorf("environment %q has %d volume(s); remove them first", id, len(volumes))
	}

	return nil
}

func (c *Client) checkNotLastEnvironment(ctx context.Context, id platform.EnvironmentID) error {
	envs, err := c.platform.Environments(ctx, id.ProjectID())

	if err != nil {
		return fmt.Errorf("list project environments: %w", err)
	}

	if len(envs) <= 1 {
		return fmt.Errorf("cannot delete %q: a project must have at least one environment", id)
	}

	return nil
}
