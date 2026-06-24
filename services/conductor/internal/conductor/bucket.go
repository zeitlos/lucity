package conductor

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/objectstorage"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type BucketID = platform.BucketID
type Bucket = objectstorage.Bucket
type BucketCredentials = objectstorage.Credentials

func (c *Client) Buckets(ctx context.Context, environment EnvironmentID) ([]Bucket, error) {
	return c.objectStorage.Buckets(ctx, environment)
}

func (c *Client) Bucket(ctx context.Context, id BucketID) (*Bucket, error) {
	return c.objectStorage.Bucket(ctx, id)
}

func (c *Client) CreateBucket(ctx context.Context, environment platform.EnvironmentID, name string) (*Bucket, error) {
	bucket, err := c.objectStorage.Create(ctx, environment, name)

	if err != nil {
		return nil, fmt.Errorf("create bucket: %w", err)
	}

	return bucket, nil
}

func (c *Client) DeleteBucket(ctx context.Context, id BucketID) (bool, error) {
	if err := c.objectStorage.Delete(ctx, id); err != nil {
		return false, fmt.Errorf("delete bucket: %w", err)
	}

	return true, nil
}

func (c *Client) BucketCredentials(ctx context.Context, id BucketID) (*BucketCredentials, error) {
	return c.objectStorage.Credentials(ctx, id)
}
