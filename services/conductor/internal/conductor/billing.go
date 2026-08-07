package conductor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/cashier"
)

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

type ResourceTier = platform.ResourceTier

const ProductionTier = platform.ProductionTier
const EcoTier = platform.EcoTier

// SetEnvironmentResources changes the environment's resource tier, which flips
// the LimitRange's default container requests: Eco runs burstable (~50% of the
// limit as request), Production runs guaranteed (request == limit). The
// namespace ResourceQuota is fixed for every environment at the defaults in
// resources.go and is not settable per environment.
//
// Existing services keep the requests baked into their Helm values until their
// next deploy; the tier only reaches them through deriveRequestsAndLimits.
func (c *Client) SetEnvironmentResources(ctx context.Context, environmentID platform.EnvironmentID, tier ResourceTier) (*Environment, error) {
	if err := c.environment.Ensure(ctx, environmentID, tier); err != nil {
		return nil, err
	}

	return c.Environment(ctx, environmentID)
}

func (c *Client) Subscription(ctx context.Context, ws string) (*BillingSubscription, error) {
	if c.cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}

	customerID, subscriptionID, err := c.stripeIDs(ctx, ws)

	if err != nil {
		return nil, err
	}

	if subscriptionID == "" {
		return nil, fmt.Errorf("billing is not configured for this workspace")
	}

	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.cashier.Subscription(callCtx, &cashier.SubscriptionRequest{
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

func (c *Client) ChangePlan(ctx context.Context, ws, plan string) (*BillingSubscription, error) {
	customerID, subscriptionID, err := c.stripeIDs(ctx, ws)

	if err != nil {
		return nil, err
	}

	if subscriptionID == "" {
		return nil, fmt.Errorf("billing is not configured for this workspace")
	}

	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.cashier.ChangePlan(callCtx, &cashier.ChangePlanRequest{
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

func (c *Client) BillingPortalURL(ctx context.Context, ws string) (string, error) {
	customerID, _, err := c.stripeIDs(ctx, ws)

	if err != nil {
		return "", err
	}

	if customerID == "" {
		return "", fmt.Errorf("billing is not configured for this workspace")
	}

	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()

	resp, err := c.cashier.BillingPortalURL(callCtx, &cashier.BillingPortalURLRequest{
		Workspace:  ws,
		ReturnUrl:  "", // Stripe defaults to billing portal home
		CustomerId: customerID,
	})

	if err != nil {
		return "", fmt.Errorf("failed to get billing portal URL: %w", err)
	}

	return resp.Url, nil
}

func (c *Client) UsageSummary(ctx context.Context, ws string) (*UsageSummaryResult, error) {
	if c.cashier == nil {
		return nil, fmt.Errorf("billing not configured")
	}
	customerID, subscriptionID, err := c.stripeIDs(ctx, ws)
	if err != nil {
		return nil, err
	}
	if customerID == "" {
		return nil, fmt.Errorf("billing is not configured for this workspace")
	}
	ctx = auth.OutgoingContext(ctx)

	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	resp, err := c.cashier.UsageSummary(callCtx, &cashier.UsageSummaryRequest{
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
func (c *Client) stripeIDs(ctx context.Context, ws string) (customerID, subscriptionID string, err error) {
	orgID, err := c.orgID(ctx, ws)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve org: %w", err)
	}
	org, err := c.logto.Organization(ctx, orgID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get organization: %w", err)
	}
	if org.CustomData != nil {
		customerID, _ = org.CustomData["stripeCustomerId"].(string)
		subscriptionID, _ = org.CustomData["stripeSubscriptionId"].(string)
	}
	return customerID, subscriptionID, nil
}

func (c *Client) CreatePlanCheckout(ctx context.Context, ws, plan string) (string, error) {
	claims, err := auth.FromContext(ctx)

	if err != nil {
		return "", err
	}

	customerID, _, err := c.stripeIDs(ctx, ws)
	if err != nil {
		return "", err
	}
	if customerID == "" {
		return "", fmt.Errorf("billing is not configured for this workspace")
	}

	successURL := fmt.Sprintf("%s/checkout/plan-success?session_id={CHECKOUT_SESSION_ID}", c.config.DashboardURL)
	cancelURL := fmt.Sprintf("%s/settings", c.config.DashboardURL)

	ctx = auth.OutgoingContext(ctx)
	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()

	resp, err := c.cashier.CreatePlanCheckoutSession(callCtx, &cashier.CreatePlanCheckoutSessionRequest{
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

func (c *Client) CompletePlanCheckout(ctx context.Context, ws, sessionID string) (*BillingSubscription, error) {
	customerID, subscriptionID, err := c.stripeIDs(ctx, ws)

	if err != nil {
		return nil, err
	}

	if customerID == "" || subscriptionID == "" {
		return nil, fmt.Errorf("billing is not configured for this workspace")
	}

	ctx = auth.OutgoingContext(ctx)
	callCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()

	resp, err := c.cashier.AddPlan(callCtx, &cashier.AddPlanRequest{
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
