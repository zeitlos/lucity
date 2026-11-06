package conductor

import (
	"context"

	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

// pickID returns the requested id when it's free, or derives one from name
// when none was requested. taken maps every used id to what uses it, such as
// "a database".
func pickID(taken map[string]string, requested, name, fallback string) (string, error) {
	if requested == "" {
		return platform.DeriveID(name, fallback, func(id string) bool {
			_, ok := taken[id]
			return ok
		}), nil
	}

	if user, ok := taken[requested]; ok {
		return "", usererror.Errorf("%s already uses the ID %q.", user, requested)
	}

	return requested, nil
}

// resourceIDs maps the id of everything inside an environment to what it is.
// Services, databases, key-value stores, buckets, volumes and agents share
// one set of ids.
func (c *Client) resourceIDs(ctx context.Context, environment platform.EnvironmentID) (map[string]string, error) {
	ids := map[string]string{}

	services, err := c.platform.Services(ctx, environment)

	if err != nil {
		return nil, err
	}

	for _, service := range services {
		ids[service.ID.Name] = "A service in this environment"
	}

	databases, err := c.platform.Databases(ctx, environment)

	if err != nil {
		return nil, err
	}

	for _, database := range databases {
		ids[database.ID.Name] = "A database in this environment"
	}

	stores, err := c.platform.KeyValueStores(ctx, environment)

	if err != nil {
		return nil, err
	}

	for _, store := range stores {
		ids[store.ID.Name] = "A key-value store in this environment"
	}

	volumes, err := c.platform.Volumes(ctx, environment)

	if err != nil {
		return nil, err
	}

	for _, volume := range volumes {
		ids[volume.ID.Name] = "A volume in this environment"
	}

	buckets, err := c.objectStorage.Buckets(ctx, environment, false)

	if err != nil {
		return nil, err
	}

	for _, bucket := range buckets {
		ids[bucket.ID.Name] = "A bucket in this environment"
	}

	if c.agentsEnabled() {
		agents, err := c.deployer.Agents().Agents(ctx, environment)

		if err != nil {
			return nil, err
		}

		for name := range agents {
			ids[name] = "An agent in this environment"
		}
	}

	return ids, nil
}

func (c *Client) newResourceID(ctx context.Context, environment platform.EnvironmentID, requested, name, fallback string) (string, error) {
	taken, err := c.resourceIDs(ctx, environment)

	if err != nil {
		return "", err
	}

	return pickID(taken, requested, name, fallback)
}

func (c *Client) RenameService(ctx context.Context, id platform.ServiceID, name string) (*Service, error) {
	if _, err := c.deployer.Services().SetDisplayName(ctx, id, name); err != nil {
		return nil, err
	}

	return c.Service(ctx, id)
}

func (c *Client) RenameDatabase(ctx context.Context, id platform.DatabaseID, name string) (*Database, error) {
	if _, err := c.deployer.Databases().SetDisplayName(ctx, id, name); err != nil {
		return nil, err
	}

	return c.Database(ctx, id)
}

func (c *Client) RenameKeyValueStore(ctx context.Context, id platform.KeyValueStoreID, name string) (*KeyValueStore, error) {
	if _, err := c.deployer.KeyValueStores().SetDisplayName(ctx, id, name); err != nil {
		return nil, err
	}

	return c.KeyValueStore(ctx, id)
}

func (c *Client) RenameVolume(ctx context.Context, id platform.VolumeID, name string) (*Volume, error) {
	if _, err := c.deployer.Volumes().SetDisplayName(ctx, id, name); err != nil {
		return nil, err
	}

	return c.Volume(ctx, id)
}

func (c *Client) RenameBucket(ctx context.Context, id platform.BucketID, name string) (*Bucket, error) {
	return c.objectStorage.SetName(ctx, id, name)
}
