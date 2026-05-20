package conductor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func (c *Client) SetServiceScaling(ctx context.Context, serviceID platform.ServiceID, replicas int, autoscaling *AutoscalingConfig) (*Service, error) {
	ws, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var as *data.AutoscalingConfig
	if autoscaling != nil {
		as = &data.AutoscalingConfig{
			Enabled:     autoscaling.Enabled,
			MinReplicas: autoscaling.MinReplicas,
			MaxReplicas: autoscaling.MaxReplicas,
			TargetCPU:   autoscaling.TargetCPU,
		}
	}

	// 1. Apply to K8s immediately via deployer
	if err := c.Deployer.SetServiceScaling(ctx, ws, serviceID.Project, serviceID.Environment, serviceID.Name, replicas, as); err != nil {
		return nil, fmt.Errorf("failed to set service scaling: %w", err)
	}

	// 2. Best-effort: sync to GitOps repo for ejection
	if pkgErr := c.Packager.SetServiceScaling(ctx, ws, serviceID.Project, serviceID.Environment, serviceID.Name, replicas, as); pkgErr != nil {
		slog.Error("failed to sync scaling to GitOps repo", "error", pkgErr, "project", serviceID.Project, "environment", serviceID.Environment, "service", serviceID.Name)
	}

	return c.Service(ctx, serviceID)
}
