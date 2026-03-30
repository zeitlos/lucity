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
	Plan              *string
	Status            string
	CurrentPeriodEnd  time.Time
	CreditAmountCents int
	CreditExpiry      *time.Time
	HasPaymentMethod  bool
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
	customerID, subscriptionID, err := c.stripeIDs(ctx)
	if err != nil {
		return nil, err
	}
	if subscriptionID == "" {
		return nil, fmt.Errorf("billing is not configured for this workspace")
	}
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Cashier.Subscription(callCtx, &cashier.SubscriptionRequest{
		Workspace:      ws,
		CustomerId:     customerID,
		SubscriptionId: subscriptionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	result := &BillingSubscription{
		Plan:              planProtoToPtr(resp.Plan),
		Status:            subscriptionStatusProtoToString(resp.Status),
		CurrentPeriodEnd:  time.Unix(resp.CurrentPeriodEnd, 0),
		CreditAmountCents: int(resp.CreditAmountCents),
		HasPaymentMethod:  resp.HasPaymentMethod,
	}
	if resp.CreditExpiry > 0 {
		t := time.Unix(resp.CreditExpiry, 0)
		result.CreditExpiry = &t
	}
	return result, nil
}

func (c *Client) ChangePlan(ctx context.Context, plan string) (*BillingSubscription, error) {
	if c.Cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	customerID, subscriptionID, err := c.stripeIDs(ctx)
	if err != nil {
		return nil, err
	}
	if subscriptionID == "" {
		return nil, fmt.Errorf("billing is not configured for this workspace")
	}
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Cashier.ChangePlan(callCtx, &cashier.ChangePlanRequest{
		Workspace:      ws,
		Plan:           stringToPlanProto(plan),
		CustomerId:     customerID,
		SubscriptionId: subscriptionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to change plan: %w", err)
	}

	result := &BillingSubscription{
		Plan:              planProtoToPtr(resp.Plan),
		Status:            subscriptionStatusProtoToString(resp.Status),
		CurrentPeriodEnd:  time.Unix(resp.CurrentPeriodEnd, 0),
		CreditAmountCents: int(resp.CreditAmountCents),
		HasPaymentMethod:  resp.HasPaymentMethod,
	}
	if resp.CreditExpiry > 0 {
		t := time.Unix(resp.CreditExpiry, 0)
		result.CreditExpiry = &t
	}
	return result, nil
}

func (c *Client) BillingPortalURL(ctx context.Context) (*BillingPortalUrlResult, error) {
	if c.Cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}
	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}
	customerID, _, err := c.stripeIDs(ctx)
	if err != nil {
		return nil, err
	}
	if customerID == "" {
		return nil, fmt.Errorf("billing is not configured for this workspace")
	}
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Cashier.BillingPortalURL(callCtx, &cashier.BillingPortalURLRequest{
		Workspace:  ws,
		ReturnUrl:  "", // Stripe defaults to billing portal home
		CustomerId: customerID,
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
	customerID, subscriptionID, err := c.stripeIDs(ctx)
	if err != nil {
		return nil, err
	}
	if customerID == "" {
		return nil, fmt.Errorf("billing is not configured for this workspace")
	}
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.Cashier.UsageSummary(callCtx, &cashier.UsageSummaryRequest{
		Workspace:      ws,
		CustomerId:     customerID,
		SubscriptionId: subscriptionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	return &UsageSummaryResult{
		ResourceCostCents:   int(resp.ResourceCostCents),
		CreditsCents:        int(resp.CreditsCents),
		EstimatedTotalCents: int(resp.EstimatedTotalCents),
	}, nil
}

// stripeIDs reads the Stripe customer and subscription IDs from the Logto org customData.
func (c *Client) stripeIDs(ctx context.Context) (customerID, subscriptionID string, err error) {
	ws, err := tenant.Require(ctx)
	if err != nil {
		return "", "", err
	}
	orgID, err := c.resolveOrgID(ctx, ws)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve org: %w", err)
	}
	org, err := c.Logto.Organization(ctx, orgID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get organization: %w", err)
	}
	if org.CustomData != nil {
		customerID, _ = org.CustomData["stripeCustomerId"].(string)
		subscriptionID, _ = org.CustomData["stripeSubscriptionId"].(string)
	}
	return customerID, subscriptionID, nil
}

func (c *Client) CreatePlanCheckout(ctx context.Context, plan string) (string, error) {
	if c.Cashier == nil {
		return "", fmt.Errorf("billing not configured")
	}
	claims := auth.FromContext(ctx)
	if claims == nil {
		return "", fmt.Errorf("unauthenticated")
	}
	customerID, _, err := c.stripeIDs(ctx)
	if err != nil {
		return "", err
	}
	if customerID == "" {
		return "", fmt.Errorf("billing is not configured for this workspace")
	}

	successURL := fmt.Sprintf("%s/checkout/plan-success?session_id={CHECKOUT_SESSION_ID}", c.DashboardURL)
	cancelURL := fmt.Sprintf("%s/settings", c.DashboardURL)

	ctx = auth.OutgoingContext(ctx)
	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()

	resp, err := c.Cashier.CreatePlanCheckoutSession(callCtx, &cashier.CreatePlanCheckoutSessionRequest{
		CustomerId: customerID,
		Plan:       stringToPlanProto(plan),
		SuccessUrl: successURL,
		CancelUrl:  cancelURL,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create plan checkout: %w", err)
	}

	slog.Info("plan checkout initiated", "plan", plan, "user", claims.Email)
	return resp.Url, nil
}

func (c *Client) CompletePlanCheckout(ctx context.Context, sessionID string) (*BillingSubscription, error) {
	if c.Cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}
	customerID, subscriptionID, err := c.stripeIDs(ctx)
	if err != nil {
		return nil, err
	}
	if customerID == "" || subscriptionID == "" {
		return nil, fmt.Errorf("billing is not configured for this workspace")
	}

	ws, err := tenant.Require(ctx)
	if err != nil {
		return nil, err
	}

	ctx = auth.OutgoingContext(ctx)
	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()

	resp, err := c.Cashier.AddPlan(callCtx, &cashier.AddPlanRequest{
		Workspace:         ws,
		CustomerId:        customerID,
		SubscriptionId:    subscriptionID,
		CheckoutSessionId: sessionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add plan: %w", err)
	}

	return &BillingSubscription{
		Plan:              planProtoToPtr(resp.Plan),
		Status:            subscriptionStatusProtoToString(resp.Status),
		CurrentPeriodEnd:  time.Unix(resp.CurrentPeriodEnd, 0),
		CreditAmountCents: int(resp.CreditAmountCents),
		HasPaymentMethod:  resp.HasPaymentMethod,
	}, nil
}

func planProtoToString(p cashier.Plan) string {
	switch p {
	case cashier.Plan_PLAN_PRO:
		return "PRO"
	case cashier.Plan_PLAN_HOBBY:
		return "HOBBY"
	default:
		return ""
	}
}

// planProtoToPtr returns a pointer to the plan string, or nil if no plan is set.
func planProtoToPtr(p cashier.Plan) *string {
	s := planProtoToString(p)
	if s == "" {
		return nil
	}
	return &s
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
