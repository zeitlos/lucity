package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/cashier"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
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

	// Best-effort: sync resources to GitOps repo for ejection
	pkgCtx, pkgCancel := context.WithTimeout(ctx, grpcTimeout)
	defer pkgCancel()
	_, pkgErr := c.Packager.SetResources(pkgCtx, &packager.SetResourcesRequest{
		Project:       projectID,
		Environment:   environment,
		Tier:          strings.ToLower(tier),
		CpuMillicores: int32(cpuMillicores),
		MemoryMb:      int32(memoryMB),
		DiskMb:        int32(diskMB),
	})
	if pkgErr != nil {
		slog.Error("failed to sync resources to GitOps repo", "error", pkgErr, "project", projectID, "environment", environment)
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

// Billing types for cashier integration

type BillingSubscription struct {
	Plan              string
	Status            string
	CurrentPeriodEnd  time.Time
	CreditAmountCents int
}

type UsageSummaryResult struct {
	ResourceCostCents   int
	CreditsCents        int
	EstimatedTotalCents int
}

type BillingPortalUrlResult struct {
	URL string
}

func (c *Client) Subscription(ctx context.Context) (*BillingSubscription, error) {
	if c.Cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Cashier.Subscription(callCtx, &cashier.SubscriptionRequest{Workspace: ws})
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &BillingSubscription{
		Plan:              planProtoToString(resp.Plan),
		Status:            subscriptionStatusProtoToString(resp.Status),
		CurrentPeriodEnd:  time.Unix(resp.CurrentPeriodEnd, 0),
		CreditAmountCents: int(resp.CreditAmountCents),
	}, nil
}

func (c *Client) ChangePlan(ctx context.Context, plan string) (*BillingSubscription, error) {
	if c.Cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Cashier.ChangePlan(callCtx, &cashier.ChangePlanRequest{
		Workspace: ws,
		Plan:      stringToPlanProto(plan),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to change plan: %w", err)
	}

	return &BillingSubscription{
		Plan:              planProtoToString(resp.Plan),
		Status:            subscriptionStatusProtoToString(resp.Status),
		CurrentPeriodEnd:  time.Unix(resp.CurrentPeriodEnd, 0),
		CreditAmountCents: int(resp.CreditAmountCents),
	}, nil
}

func (c *Client) BillingPortalURL(ctx context.Context) (*BillingPortalUrlResult, error) {
	if c.Cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Cashier.BillingPortalURL(callCtx, &cashier.BillingPortalURLRequest{
		Workspace: ws,
		ReturnUrl: "", // Stripe defaults to billing portal home
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get billing portal URL: %w", err)
	}

	return &BillingPortalUrlResult{URL: resp.Url}, nil
}

func (c *Client) UsageSummary(ctx context.Context) (*UsageSummaryResult, error) {
	if c.Cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Cashier.UsageSummary(callCtx, &cashier.UsageSummaryRequest{Workspace: ws})
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	return &UsageSummaryResult{
		ResourceCostCents:   int(resp.ResourceCostCents),
		CreditsCents:        int(resp.CreditsCents),
		EstimatedTotalCents: int(resp.EstimatedTotalCents),
	}, nil
}

func planProtoToString(p cashier.Plan) string {
	switch p {
	case cashier.Plan_PLAN_PRO:
		return "PRO"
	default:
		return "HOBBY"
	}
}

func stringToPlanProto(s string) cashier.Plan {
	switch s {
	case "PRO":
		return cashier.Plan_PLAN_PRO
	default:
		return cashier.Plan_PLAN_HOBBY
	}
}

func subscriptionStatusProtoToString(s cashier.SubscriptionStatus) string {
	switch s {
	case cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE:
		return "ACTIVE"
	case cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_PAST_DUE:
		return "PAST_DUE"
	case cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED:
		return "CANCELED"
	case cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_INCOMPLETE:
		return "INCOMPLETE"
	case cashier.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIALING:
		return "TRIALING"
	default:
		return "ACTIVE"
	}
}
