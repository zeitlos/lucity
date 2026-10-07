package ovh

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	goovh "github.com/ovh/go-ovh/ovh"
)

type Client struct {
	api       *goovh.Client
	projectID string
}

type Bucket struct {
	Name        string            `json:"name"`
	ObjectsSize int64             `json:"objectsSize"`
	Tags        map[string]string `json:"tags"`
}

func New(endpoint, applicationKey, applicationSecret, consumerKey, projectID string) (*Client, error) {
	api, err := goovh.NewClient(endpoint, applicationKey, applicationSecret, consumerKey)

	if err != nil {
		return nil, fmt.Errorf("create ovh client: %w", err)
	}

	return &Client{
		api:       api,
		projectID: projectID,
	}, nil
}

func (c *Client) Buckets(ctx context.Context, region string) ([]Bucket, error) {
	var buckets []Bucket

	if err := c.api.GetWithContext(ctx, c.storagePath(region), &buckets); err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}

	return buckets, nil
}

func (c *Client) Bucket(ctx context.Context, region, name string) (*Bucket, error) {
	var bucket Bucket

	if err := c.api.GetWithContext(ctx, c.storagePath(region)+"/"+url.PathEscape(name)+"?noObjects=true", &bucket); err != nil {
		return nil, fmt.Errorf("get bucket %s: %w", name, err)
	}

	return &bucket, nil
}

func (c *Client) storagePath(region string) string {
	return fmt.Sprintf("/cloud/project/%s/region/%s/storage", url.PathEscape(c.projectID), url.PathEscape(strings.ToUpper(region)))
}
