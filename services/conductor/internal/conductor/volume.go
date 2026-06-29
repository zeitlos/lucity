package conductor

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type VolumeID = platform.VolumeID
type Volume = platform.Volume

func (c *Client) Volumes(ctx context.Context, environment EnvironmentID) ([]Volume, error) {
	return c.platform.Volumes(ctx, environment)
}

func (c *Client) Volume(ctx context.Context, id VolumeID) (*Volume, error) {
	return c.platform.Volume(ctx, id)
}

func (c *Client) CreateVolume(ctx context.Context, environment platform.EnvironmentID, name, size string) (*Volume, error) {
	parsedSize, err := resource.ParseQuantity(size)

	if err != nil {
		return nil, fmt.Errorf("parse size: %w", err)
	}

	if _, err := c.deployer.Volumes().Create(ctx, environment, name, parsedSize); err != nil {
		return nil, fmt.Errorf("create volume: %w", err)
	}

	return &Volume{
		ID: platform.VolumeID{
			Workspace:   environment.Workspace,
			Project:     environment.Project,
			Environment: environment.Name,
			Name:        name,
		},
		Name:   name,
		Size:   parsedSize,
		Status: platform.VolumePending,
	}, nil
}

func (c *Client) DeleteVolume(ctx context.Context, id platform.VolumeID) (bool, error) {
	if err := c.deployer.Volumes().Delete(ctx, id); err != nil {
		return false, fmt.Errorf("delete volume: %w", err)
	}

	return true, nil
}

func (c *Client) MountVolume(ctx context.Context, volume platform.VolumeID, service platform.ServiceID, path string) (*Volume, error) {
	if volume.EnvironmentID() != service.EnvironmentID() {
		return nil, fmt.Errorf("volume and service must be in the same environment")
	}

	if _, err := c.deployer.Services().Mount(ctx, service, volume, path); err != nil {
		return nil, fmt.Errorf("mount volume: %w", err)
	}

	return c.platform.Volume(ctx, volume)
}

func (c *Client) UnmountVolume(ctx context.Context, volume platform.VolumeID) (*Volume, error) {
	if _, err := c.deployer.Services().Unmount(ctx, volume); err != nil {
		return nil, fmt.Errorf("unmount volume: %w", err)
	}

	return c.platform.Volume(ctx, volume)
}
