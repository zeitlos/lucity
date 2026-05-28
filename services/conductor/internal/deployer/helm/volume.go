package helm

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type volumeClient struct {
	client *Client
}

func (v *volumeClient) Create(ctx context.Context, env platform.EnvironmentID, name string, spec deployer.VolumeSpec) (deployer.RevisionID, error) {
	return "", fmt.Errorf("Create: chart does not support volumes yet")
}

func (v *volumeClient) Delete(ctx context.Context, id platform.VolumeID) error {
	return fmt.Errorf("Delete: chart does not support volumes yet")
}

var _ deployer.VolumeClient = (*volumeClient)(nil)
