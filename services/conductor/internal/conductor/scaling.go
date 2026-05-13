package conductor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func (c *Client) SetServiceScaling(ctx context.Context, service platform.ServiceID, replicas int, autoscaling *AutoscalingConfig) (*ScalingConfig, error) {
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
	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	if err := c.Deployer.SetServiceScaling(callCtx, ws, service.Project, service.Environment, service.Name, replicas, as); err != nil {
		return nil, fmt.Errorf("failed to set service scaling: %w", err)
	}

	// 2. Best-effort: sync to GitOps repo for ejection
	pkgCtx, pkgCancel := context.WithTimeout(ctx, grpcTimeout)
	defer pkgCancel()
	if pkgErr := c.Packager.SetServiceScaling(pkgCtx, ws, service.Project, service.Environment, service.Name, replicas, as); pkgErr != nil {
		slog.Error("failed to sync scaling to GitOps repo", "error", pkgErr, "project", service.Project, "environment", service.Environment, "service", service.Name)
	}

	result := &ScalingConfig{Replicas: replicas}
	if autoscaling != nil && autoscaling.Enabled {
		result.Autoscaling = autoscaling
	}
	return result, nil
}
