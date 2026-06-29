package helm

import (
	"context"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/values"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type volumeClient struct {
	client *Client
}

func (v *volumeClient) Create(ctx context.Context, env platform.EnvironmentID, name string, size resource.Quantity) (deployer.RevisionID, error) {
	return v.client.applyEnv(ctx, env, func(e *values.Env) error {
		return values.CreateVolume(e, name, size)
	})
}

func (v *volumeClient) Delete(ctx context.Context, id platform.VolumeID) error {
	_, err := v.client.applyEnv(ctx, id.EnvironmentID(), func(e *values.Env) error {
		return values.DeleteVolume(e, id.Name)
	})

	return err
}

var _ deployer.VolumeClient = (*volumeClient)(nil)
