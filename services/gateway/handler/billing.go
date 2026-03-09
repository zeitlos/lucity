package handler

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/tenant"
)

type EnvironmentResources struct {
	Tier          string
	CpuMillicores int
	MemoryMB      int
	DiskMB        int
}

func (c *Client) EnvironmentResources(ctx context.Context, projectID, environment string) (*EnvironmentResources, error) {
	if _, err := tenant.Require(ctx); err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)
	ctx = tenant.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Deployer.ResourceQuota(callCtx, &deployer.ResourceQuotaRequest{
		Project:     projectID,
		Environment: environment,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource quota: %w", err)
	}

	return &EnvironmentResources{
		Tier:          protoTierToString(resp.Tier),
		CpuMillicores: int(resp.CpuMillicores),
		MemoryMB:      int(resp.MemoryMb),
		DiskMB:        int(resp.DiskMb),
	}, nil
}

func (c *Client) SetEnvironmentResources(ctx context.Context, projectID, environment, tier string, cpuMillicores, memoryMB, diskMB int) (*EnvironmentResources, error) {
	if _, err := tenant.Require(ctx); err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)
	ctx = tenant.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Deployer.SetResourceQuota(callCtx, &deployer.SetResourceQuotaRequest{
		Project:       projectID,
		Environment:   environment,
		Tier:          stringToProtoTier(tier),
		CpuMillicores: int32(cpuMillicores),
		MemoryMb:      int32(memoryMB),
		DiskMb:        int32(diskMB),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set resource quota: %w", err)
	}

	return &EnvironmentResources{
		Tier:          protoTierToString(resp.Tier),
		CpuMillicores: int(resp.CpuMillicores),
		MemoryMB:      int(resp.MemoryMb),
		DiskMB:        int(resp.DiskMb),
	}, nil
}

func protoTierToString(t deployer.ResourceTier) string {
	switch t {
	case deployer.ResourceTier_RESOURCE_TIER_PRODUCTION:
		return "PRODUCTION"
	default:
		return "ECO"
	}
}

func stringToProtoTier(s string) deployer.ResourceTier {
	switch s {
	case "PRODUCTION":
		return deployer.ResourceTier_RESOURCE_TIER_PRODUCTION
	default:
		return deployer.ResourceTier_RESOURCE_TIER_ECO
	}
}
