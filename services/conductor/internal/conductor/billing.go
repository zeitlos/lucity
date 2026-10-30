package conductor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	pkgbilling "github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
	"github.com/zeitlos/lucity/services/conductor/internal/billing"
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

// PlanSelection is the subscription now on the selected plan, or the URL of
// the checkout to complete first.
type PlanSelection struct {
	Subscription *BillingSubscription
	CheckoutURL  string
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

	environment, err := c.Environment(ctx, environmentID)

	if err != nil {
		return nil, err
	}

	environment.ResourceTier = tier

	return environment, nil
}

// Subscription returns the workspace's subscription, or nil when it never
// had one.
func (c *Client) Subscription(ctx context.Context, ws string) (*BillingSubscription, error) {
	customerID, err := c.billingCustomer(ctx, ws)

	if err != nil {
		return nil, err
	}

	subscription, err := c.billing.Subscription(ctx, customerID)

	if err != nil {
		return nil, fmt.Errorf("read subscription: %w", err)
	}

	return convertSubscription(subscription), nil
}

func (c *Client) paid(ctx context.Context, ws string) bool {
	if c.billing == nil || c.config.PaidPlacement.IsZero() {
		return false
	}

	customerID, err := c.stripeCustomerID(ctx, ws)

	if err != nil || customerID == "" {
		return false
	}

	paid, err := c.billing.Paid(ctx, customerID)

	if err != nil {
		slog.InfoContext(ctx, "plan lookup failed, keeping the default placement", "workspace", ws, "error", err)
		return false
	}

	return paid
}

// SelectPlan puts the workspace on a plan, or returns the checkout to
// complete first when it has no card on file.
func (c *Client) SelectPlan(ctx context.Context, ws, plan string) (*PlanSelection, error) {
	customerID, err := c.billingCustomer(ctx, ws)

	if err != nil {
		return nil, err
	}

	selection, err := c.billing.SelectPlan(ctx, customerID, ws, planOf(plan), billing.Checkout{
		SuccessURL: c.config.DashboardURL + "/checkout/plan-success?session_id={CHECKOUT_SESSION_ID}",
		CancelURL:  c.config.DashboardURL + "/workspace/settings?tab=billing",
	})

	c.forgetCapabilities(ws)

	if err != nil {
		return nil, err
	}

	if selection.CheckoutURL != "" {
		slog.InfoContext(ctx, "plan checkout started", "workspace", ws, "plan", plan)
	} else {
		slog.InfoContext(ctx, "plan selected", "workspace", ws, "plan", plan)
	}

	return &PlanSelection{Subscription: convertSubscription(selection.Subscription), CheckoutURL: selection.CheckoutURL}, nil
}

func (c *Client) CompletePlanCheckout(ctx context.Context, ws, sessionID string) (*BillingSubscription, error) {
	customerID, err := c.billingCustomer(ctx, ws)

	if err != nil {
		return nil, err
	}

	subscription, err := c.billing.CompleteCheckout(ctx, customerID, ws, sessionID)

	c.forgetCapabilities(ws)

	if err != nil {
		return nil, err
	}

	if subscription == nil {
		return nil, errors.New("no subscription after a completed checkout")
	}

	c.resume(ctx, ws, customerID)

	return convertSubscription(subscription), nil
}

// AddPaymentMethod starts a checkout that saves a card for the workspace
// without charging it, and returns its URL.
func (c *Client) AddPaymentMethod(ctx context.Context, ws, returnPath string) (string, error) {
	customerID, err := c.billingCustomer(ctx, ws)

	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(returnPath, "/") || strings.HasPrefix(returnPath, "//") {
		returnPath = "/"
	}

	checkoutURL, err := c.billing.PaymentMethodCheckout(ctx, customerID, ws, billing.Checkout{
		SuccessURL: c.config.DashboardURL + "/checkout/card-success?session_id={CHECKOUT_SESSION_ID}&return=" + url.QueryEscape(returnPath),
		CancelURL:  c.config.DashboardURL + returnPath,
	})

	if err != nil {
		return "", err
	}

	slog.InfoContext(ctx, "card checkout started", "workspace", ws)

	return checkoutURL, nil
}

// CompletePaymentMethodCheckout confirms a finished card checkout. A prepaid
// card is removed again and reported with ErrPrepaidCard.
func (c *Client) CompletePaymentMethodCheckout(ctx context.Context, ws, sessionID string) (*WorkspaceDetails, error) {
	customerID, err := c.billingCustomer(ctx, ws)

	if err != nil {
		return nil, err
	}

	err = c.billing.CompletePaymentMethodCheckout(ctx, customerID, sessionID)

	c.forgetCapabilities(ws)

	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "card added", "workspace", ws)

	return c.Workspace(ctx, ws)
}

func (c *Client) BillingPortalURL(ctx context.Context, ws string) (string, error) {
	customerID, err := c.billingCustomer(ctx, ws)

	if err != nil {
		return "", err
	}

	return c.billing.PortalURL(ctx, customerID, c.config.DashboardURL+"/workspace/settings?tab=billing")
}

// UsageSummary returns the costs of the current billing period so far, or
// nil when the workspace never had a subscription.
func (c *Client) UsageSummary(ctx context.Context, ws string) (*UsageSummaryResult, error) {
	customerID, err := c.billingCustomer(ctx, ws)

	if err != nil {
		return nil, err
	}

	usage, err := c.billing.Usage(ctx, customerID)

	if err != nil || usage == nil {
		return nil, err
	}

	return &UsageSummaryResult{
		ResourceCostCents:   int(usage.ResourceCents),
		CreditsCents:        int(usage.CreditCents),
		EstimatedTotalCents: int(usage.EstimatedTotalCents),
	}, nil
}

func (c *Client) resume(ctx context.Context, ws, customerID string) {
	standing, judged, err := c.billing.Standing(ctx, customerID)

	if err != nil {
		slog.WarnContext(ctx, "resume after checkout: failed to read standing", "workspace", ws, "error", err)
		return
	}

	if !judged || !standing.Good {
		return
	}

	orgID, err := c.orgID(ctx, ws)

	if err != nil {
		slog.WarnContext(ctx, "resume after checkout: failed to resolve org", "workspace", ws, "error", err)
		return
	}

	org, err := c.logto.Organization(ctx, orgID)

	if err != nil {
		slog.WarnContext(ctx, "resume after checkout: failed to read org", "workspace", ws, "error", err)
		return
	}

	if abuse, _ := org.CustomData["abuse"].(bool); abuse {
		paid, err := c.billing.Paid(ctx, customerID)

		if err != nil {
			slog.WarnContext(ctx, "resume after checkout: failed to read plan", "workspace", ws, "error", err)
			return
		}

		if !paid {
			return
		}
	}

	if suspended, _ := org.CustomData["suspended"].(bool); !suspended {
		return
	}

	delete(org.CustomData, "suspended")

	if _, err := c.logto.UpdateOrganizationCustomData(ctx, orgID, org.CustomData); err != nil {
		slog.WarnContext(ctx, "resume after checkout: failed to update org", "workspace", ws, "error", err)
		return
	}

	slog.InfoContext(ctx, "workspace resumed after checkout", "workspace", ws)
}

func (c *Client) billingCustomer(ctx context.Context, ws string) (string, error) {
	if c.billing == nil {
		return "", errors.New("billing not configured")
	}

	customerID, err := c.stripeCustomerID(ctx, ws)

	if err != nil {
		return "", err
	}

	if customerID == "" {
		return "", usererror.New("billing is not configured for this workspace")
	}

	return customerID, nil
}

// stripeCustomerID reads the Stripe customer ID from the Logto org customData.
func (c *Client) stripeCustomerID(ctx context.Context, ws string) (string, error) {
	orgID, err := c.orgID(ctx, ws)
	if err != nil {
		return "", fmt.Errorf("failed to resolve org: %w", err)
	}
	org, err := c.logto.Organization(ctx, orgID)
	if err != nil {
		return "", fmt.Errorf("failed to get organization: %w", err)
	}
	customerID, _ := org.CustomData["stripeCustomerId"].(string)
	return customerID, nil
}

func convertSubscription(subscription *billing.Subscription) *BillingSubscription {
	if subscription == nil {
		return nil
	}

	result := &BillingSubscription{
		Status:            string(subscription.Status),
		CurrentPeriodEnd:  subscription.CurrentPeriodEnd,
		CreditAmountCents: int(subscription.CreditCents),
		CreditExpiry:      subscription.CreditExpiry,
		HasPaymentMethod:  subscription.HasPaymentMethod,
	}

	if subscription.Plan.Paid() {
		plan := strings.ToUpper(string(subscription.Plan))
		result.Plan = &plan
	}

	return result
}

func planOf(plan string) pkgbilling.Plan {
	switch plan {
	case "HOBBY":
		return pkgbilling.PlanHobby
	case "PRO":
		return pkgbilling.PlanPro
	}

	return pkgbilling.Plan(strings.ToLower(plan))
}
