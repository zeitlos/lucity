package conductor

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type KeyValueStoreID = platform.KeyValueStoreID
type KeyValueStore = platform.KeyValueStore

type KeyValueStoreCredentials struct {
	Type     EndpointType
	Host     string
	Port     string
	Password string
	URI      string
}

func (c *Client) KeyValueStores(ctx context.Context, environment EnvironmentID) ([]KeyValueStore, error) {
	return c.platform.KeyValueStores(ctx, environment)
}

func (c *Client) KeyValueStore(ctx context.Context, id KeyValueStoreID) (*KeyValueStore, error) {
	return c.platform.KeyValueStore(ctx, id)
}

func (c *Client) CreateKeyValueStore(ctx context.Context, environment platform.EnvironmentID, name, version, size string) (*KeyValueStore, error) {
	if version == "" {
		version = "8"
	}

	if size == "" {
		size = "1Gi"
	}

	parsedSize, err := resource.ParseQuantity(size)

	if err != nil {
		return nil, fmt.Errorf("parse size: %w", err)
	}

	if _, err := c.deployer.KeyValueStores().Create(ctx, environment, name, deployer.KeyValueStoreSpec{
		Version:  version,
		Size:     parsedSize,
		Password: randCrockford32(32),
	}); err != nil {
		return nil, fmt.Errorf("create key-value store: %w", err)
	}

	return &KeyValueStore{
		Name:    name,
		Version: version,
		Size:    parsedSize,
	}, nil
}

func (c *Client) DeleteKeyValueStore(ctx context.Context, store platform.KeyValueStoreID) (bool, error) {
	if err := c.deployer.KeyValueStores().Delete(ctx, store); err != nil {
		return false, fmt.Errorf("delete key-value store: %w", err)
	}

	return true, nil
}

func (c *Client) KeyValueStoreCredentials(ctx context.Context, store platform.KeyValueStoreID) ([]KeyValueStoreCredentials, error) {
	creds, err := c.platform.KeyValueStoreCredentials(ctx, store)

	if errors.Is(err, platform.ErrDatabaseProvisioning) {
		return nil, &DatabaseProvisioningError{}
	}

	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	return []KeyValueStoreCredentials{{
		Type:     InternalEndpoint,
		Host:     creds.Host,
		Port:     creds.Port,
		Password: creds.Password,
		URI:      creds.URI,
	}}, nil
}
