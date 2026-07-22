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
type BucketObjectListing = objectstorage.ObjectListing

const maxBucketsPerEnvironment = 10

func (c *Client) Buckets(ctx context.Context, environment EnvironmentID) ([]Bucket, error) {
	return c.objectStorage.Buckets(ctx, environment)
}

func (c *Client) Bucket(ctx context.Context, id BucketID) (*Bucket, error) {
	return c.objectStorage.Bucket(ctx, id)
}

func (c *Client) CreateBucket(ctx context.Context, environment platform.EnvironmentID, name string) (*Bucket, error) {
	existing, err := c.objectStorage.Buckets(ctx, environment)

	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}

	if len(existing) >= maxBucketsPerEnvironment {
		return nil, fmt.Errorf("bucket limit reached: %d per environment", maxBucketsPerEnvironment)
	}

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

func (c *Client) SetBucketPublic(ctx context.Context, id BucketID, public bool) (*Bucket, error) {
	bucket, err := c.objectStorage.SetPublic(ctx, id, public)

	if err != nil {
		return nil, fmt.Errorf("set bucket public: %w", err)
	}

	return bucket, nil
}

func (c *Client) BucketObjects(ctx context.Context, id BucketID, prefix string) (*BucketObjectListing, error) {
	return c.objectStorage.Objects(ctx, id, prefix)
}

func (c *Client) BucketObjectDownloadURL(ctx context.Context, id BucketID, key string) (string, error) {
	return c.objectStorage.PresignDownload(ctx, id, key)
}

func (c *Client) BucketObjectUploadURL(ctx context.Context, id BucketID, key string) (string, error) {
	return c.objectStorage.PresignUpload(ctx, id, key)
}

func (c *Client) DeleteBucketObject(ctx context.Context, id BucketID, key string) (bool, error) {
	if err := c.objectStorage.DeleteObject(ctx, id, key); err != nil {
		return false, fmt.Errorf("delete bucket object: %w", err)
	}

	return true, nil
}
