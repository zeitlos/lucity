package helm

import (
	"context"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type databaseClient struct {
	client *Client
}

func (d *databaseClient) Create(ctx context.Context, env platform.EnvironmentID, name string, spec deployer.DatabaseSpec) (deployer.RevisionID, error) {
	return d.client.applyEnv(ctx, env, func(e *values.Env) error {
		return values.CreateDatabase(e, name, values.DatabaseSpec{
			Version:   spec.Version,
			Size:      spec.Size,
			Resources: deriveRequestsAndLimtis(spec.Resources, spec.ResourceTier),
		})
	})
}

func (d *databaseClient) SetResources(ctx context.Context, id platform.DatabaseID, tier platform.ResourceTier, res deployer.Resources) (deployer.RevisionID, error) {
	return d.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetDatabaseResources(e, id.Name, deriveRequestsAndLimtis(res, tier))
	})
}

func (d *databaseClient) SetStorage(ctx context.Context, id platform.DatabaseID, size resource.Quantity) (deployer.RevisionID, error) {
	return d.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.SetDatabaseStorage(e, id.Name, size.String())
	})
}

func (d *databaseClient) Delete(ctx context.Context, id platform.DatabaseID) error {
	_, err := d.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.DeleteDatabase(e, id.Name)
	})

	return err
}

func (d *databaseClient) Expose(ctx context.Context, id platform.DatabaseID, host string) error {
	_, err := d.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.ExposeDatabase(e, id.Name, host)
	})

	return err
}

func (d *databaseClient) Unexpose(ctx context.Context, id platform.DatabaseID) error {
	_, err := d.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.UnexposeDatabase(e, id.Name)
	})

	return err
}

var _ deployer.DatabaseClient = (*databaseClient)(nil)
