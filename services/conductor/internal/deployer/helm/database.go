package helm

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type databaseClient struct {
	client *Client
}

func (d *databaseClient) Create(ctx context.Context, env platform.EnvironmentID, name string, spec deployer.DatabaseSpec) (deployer.RevisionID, error) {
	if err := validateResources(spec.Resources); err != nil {
		return "", err
	}

	return d.client.applyEnv(ctx, env, func(e *values.Env) error {
		return values.CreateDatabase(e, name, values.DatabaseSpec{
			Version:    spec.Version,
			Size:       spec.Size,
			Resources:  deriveRequestsAndLimtis(spec.Resources, spec.ResourceTier),
			Parameters: postgresParameters(spec.Resources.CPU, spec.Resources.Memory),
		})
	})
}

func (d *databaseClient) Restore(ctx context.Context, source platform.DatabaseID, name string, spec deployer.DatabaseSpec, targetTime *time.Time) (deployer.RevisionID, error) {
	if err := validateResources(spec.Resources); err != nil {
		return "", err
	}

	return d.client.applyEnv(ctx, source.EnvironmentID(), func(e *values.Env) error {
		return values.RestoreDatabase(e, source.Name, name, values.DatabaseSpec{
			Version:    spec.Version,
			Size:       spec.Size,
			Resources:  deriveRequestsAndLimtis(spec.Resources, spec.ResourceTier),
			Parameters: postgresParameters(spec.Resources.CPU, spec.Resources.Memory),
		}, targetTime)
	})
}

func (d *databaseClient) SetResources(ctx context.Context, id platform.DatabaseID, tier platform.ResourceTier, res deployer.Resources) (deployer.RevisionID, error) {
	if err := validateResources(res); err != nil {
		return "", err
	}

	return d.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		if err := values.SetDatabaseResources(e, id.Name, deriveRequestsAndLimtis(res, tier)); err != nil {
			return err
		}

		return values.SetDatabaseParameters(e, id.Name, postgresParameters(res.CPU, res.Memory))
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

func (d *databaseClient) AllowRules(ctx context.Context, id platform.DatabaseID) ([]deployer.AllowRule, error) {
	env, err := d.client.loadEnv(ctx, id.EnvironmentID())

	if err != nil {
		return nil, err
	}

	postgres, ok := env.Databases.Postgres[id.Name]

	if !ok {
		return nil, fmt.Errorf("database %q not found", id.Name)
	}

	if postgres.PublicAccess == nil {
		return []deployer.AllowRule{}, nil
	}

	rules := make([]deployer.AllowRule, 0, len(postgres.PublicAccess.Allow))

	for _, rule := range postgres.PublicAccess.Allow {
		prefix, err := netip.ParsePrefix(rule.Range)

		if err != nil {
			return nil, fmt.Errorf("database %q: invalid allow range %q: %w", id.Name, rule.Range, err)
		}

		rules = append(rules, deployer.AllowRule{Range: prefix, Description: rule.Description})
	}

	return rules, nil
}

func (d *databaseClient) AddAllowRule(ctx context.Context, id platform.DatabaseID, rule deployer.AllowRule) error {
	_, err := d.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.AddDatabaseAllowRule(e, id.Name, values.AllowRule{
			Range:       rule.Range.String(),
			Description: rule.Description,
		})
	})

	return err
}

func (d *databaseClient) RemoveAllowRule(ctx context.Context, id platform.DatabaseID, ipRange netip.Prefix) error {
	_, err := d.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.RemoveDatabaseAllowRule(e, id.Name, ipRange.String())
	})

	return err
}

var _ deployer.DatabaseClient = (*databaseClient)(nil)
