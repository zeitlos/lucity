package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/pkg/tenant"
)

func (c *Client) SetServiceScaling(ctx context.Context, projectID, environment, service string, replicas int, autoscaling *AutoscalingConfig) (*ScalingConfig, error) {
	if _, err := tenant.Require(ctx); err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)
	ctx = tenant.OutgoingContext(ctx)

	// 1. Apply to K8s immediately via deployer
	req := &deployer.SetServiceScalingRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		Replicas:    int32(replicas),
	}
	if autoscaling != nil {
		req.Autoscaling = &deployer.DeployerAutoscalingConfig{
			Enabled:     autoscaling.Enabled,
			MinReplicas: int32(autoscaling.MinReplicas),
			MaxReplicas: int32(autoscaling.MaxReplicas),
			TargetCpu:   int32(autoscaling.TargetCPU),
		}
	}

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := c.Deployer.SetServiceScaling(callCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to set service scaling: %w", err)
	}

	// 2. Best-effort: sync to GitOps repo for ejection
	pkgReq := &packager.SetServiceScalingRequest{
		Project:     projectID,
		Environment: environment,
		Service:     service,
		Replicas:    int32(replicas),
	}
	if autoscaling != nil {
		pkgReq.Autoscaling = &packager.PackagerAutoscalingConfig{
			Enabled:     autoscaling.Enabled,
			MinReplicas: int32(autoscaling.MinReplicas),
			MaxReplicas: int32(autoscaling.MaxReplicas),
			TargetCpu:   int32(autoscaling.TargetCPU),
		}
	}

	pkgCtx, pkgCancel := context.WithTimeout(ctx, grpcTimeout)
	defer pkgCancel()
	_, pkgErr := c.Packager.SetServiceScaling(pkgCtx, pkgReq)
	if pkgErr != nil {
		slog.Error("failed to sync scaling to GitOps repo", "error", pkgErr, "project", projectID, "environment", environment, "service", service)
	}

	result := &ScalingConfig{
		Replicas: replicas,
	}
	if autoscaling != nil && autoscaling.Enabled {
		result.Autoscaling = autoscaling
	}
	return result, nil
}
