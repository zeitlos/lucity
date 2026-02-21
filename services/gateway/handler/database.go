package handler

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/packager"
)

type Database struct {
	Name      string
	Version   string
	Instances int
	Size      string
}

func (c *Client) CreateDatabase(ctx context.Context, projectID, name, version string, instances int, size string) (*Database, error) {
	ctx = auth.OutgoingContext(ctx)

	if version == "" {
		version = "16"
	}
	if instances == 0 {
		instances = 1
	}
	if size == "" {
		size = "10Gi"
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.AddDatabase(callCtx, &packager.AddDatabaseRequest{
		Project:   projectID,
		Name:      name,
		Version:   version,
		Instances: int32(instances),
		Size:      size,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return &Database{
		Name:      name,
		Version:   version,
		Instances: instances,
		Size:      size,
	}, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, projectID, name string) (bool, error) {
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Packager.RemoveDatabase(callCtx, &packager.RemoveDatabaseRequest{
		Project: projectID,
		Name:    name,
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete database: %w", err)
	}
	return true, nil
}

func (c *Client) Databases(ctx context.Context, projectID string) ([]Database, error) {
	proj, err := c.Project(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return proj.Databases, nil
}
