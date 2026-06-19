package helm

import (
	"context"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type keyValueStoreClient struct {
	client *Client
}

func (k *keyValueStoreClient) Create(ctx context.Context, env platform.EnvironmentID, name string, spec deployer.KeyValueStoreSpec) (deployer.RevisionID, error) {
	return k.client.applyEnv(ctx, env, func(e *values.Env) error {
		return values.CreateKeyValueStore(e, name, values.KeyValueStoreSpec{
			Version:  spec.Version,
			Size:     spec.Size,
			Password: spec.Password,
		})
	})
}

func (k *keyValueStoreClient) Delete(ctx context.Context, id platform.KeyValueStoreID) error {
	_, err := k.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.DeleteKeyValueStore(e, id.Name)
	})

	return err
}

var _ deployer.KeyValueStoreClient = (*keyValueStoreClient)(nil)
