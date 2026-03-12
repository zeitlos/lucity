package stripe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	gostripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/billing/creditbalancesummary"
	"github.com/stripe/stripe-go/v82/billing/creditgrant"
	"github.com/stripe/stripe-go/v82/billing/meterevent"
	portalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/invoice"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/subscriptionitem"
)

// PriceConfig holds all Stripe Price IDs for the billing model.
type PriceConfig struct {
	HobbyPriceID    string
	ProPriceID      string
	EcoCPUPriceID   string
	EcoMemPriceID   string
	EcoDiskPriceID  string
	ProdCPUPriceID  string
	ProdMemPriceID  string
	ProdDiskPriceID string
}

// MeterConfig holds Billing Meter event names for usage-based billing.
type MeterConfig struct {
	EcoCPUEventName  string
	EcoMemEventName  string
	EcoDiskEventName string
	ProdCPUEventName  string
	ProdMemEventName  string
	ProdDiskEventName string
}

// Client wraps the Stripe API for billing operations.
type Client struct {
	Prices PriceConfig
	Meters MeterConfig
}

// NewClient creates a Stripe client and sets the global API key.
func NewClient(secretKey string, prices PriceConfig, meters MeterConfig) *Client {
	gostripe.Key = secretKey
	return &Client{Prices: prices, Meters: meters}
}

// CreateCustomer creates a Stripe Customer for a workspace.
func (c *Client) CreateCustomer(ctx context.Context, workspace, name, email string) (string, error) {
	params := &gostripe.CustomerParams{
		Name:  gostripe.String(name),
		Email: gostripe.String(email),
	}
	params.AddMetadata("workspace", workspace)

	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe customer: %w", err)
	}
	return cust.ID, nil
}

// CreateSubscription creates a subscription with a plan + 6 metered resource line items.
// trialDays > 0 starts the subscription with a free trial period.
// After creation, an initial credit grant is created for the billing period.
func (c *Client) CreateSubscription(ctx context.Context, customerID, workspace string, planPriceID string, trialDays int) (string, error) {
	params := &gostripe.SubscriptionParams{
		Customer:        gostripe.String(customerID),
		PaymentBehavior: gostripe.String("default_incomplete"),
		Items: []*gostripe.SubscriptionItemsParams{
			{Price: gostripe.String(planPriceID)},
			{Price: gostripe.String(c.Prices.EcoCPUPriceID)},
			{Price: gostripe.String(c.Prices.EcoMemPriceID)},
			{Price: gostripe.String(c.Prices.EcoDiskPriceID)},
			{Price: gostripe.String(c.Prices.ProdCPUPriceID)},
			{Price: gostripe.String(c.Prices.ProdMemPriceID)},
			{Price: gostripe.String(c.Prices.ProdDiskPriceID)},
		},
	}
	if trialDays > 0 {
		params.TrialPeriodDays = gostripe.Int64(int64(trialDays))
	}
	params.AddMetadata("workspace", workspace)

	sub, err := subscription.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe subscription: %w", err)
	}

	// Create initial credit grant so credits exist from day one.
	if len(sub.Items.Data) > 0 {
		var creditCents int64
		var expiresAt int64
		var name string
		if sub.TrialEnd > 0 {
			creditCents = 500 // EUR 5 trial credit, regardless of plan
			expiresAt = sub.TrialEnd
			name = "Trial credit"
		} else {
			creditCents = PlanCreditCents(planPriceID, c.Prices)
			expiresAt = sub.Items.Data[0].CurrentPeriodEnd
			name = "Plan credit"
		}
		if err := c.CreateCreditGrant(ctx, customerID, creditCents, name, expiresAt); err != nil {
			// Non-fatal — metering worker will create it on next tick if missing.
			slog.Warn("failed to create initial credit grant", "error", err)
		}
	}

	return sub.ID, nil
}

// EndTrial ends a trial immediately by setting trial_end to now.
func (c *Client) EndTrial(ctx context.Context, subscriptionID string) error {
	_, err := subscription.Update(subscriptionID, &gostripe.SubscriptionParams{
		TrialEnd: gostripe.Int64(time.Now().Unix()),
	})
	if err != nil {
		return fmt.Errorf("failed to end trial: %w", err)
	}
	return nil
}

// ChangePlan swaps the plan subscription item from one price to another.
func (c *Client) ChangePlan(ctx context.Context, subscriptionID, newPlanPriceID, oldPlanPriceID string) (*gostripe.Subscription, error) {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Find the current plan item (matches old plan price)
	var planItemID string
	for _, item := range sub.Items.Data {
		if item.Price.ID == oldPlanPriceID {
			planItemID = item.ID
			break
		}
	}
	if planItemID == "" {
		return nil, fmt.Errorf("plan item not found on subscription")
	}

	// Update the plan item's price
	_, err = subscriptionitem.Update(planItemID, &gostripe.SubscriptionItemParams{
		Price:             gostripe.String(newPlanPriceID),
		ProrationBehavior: gostripe.String("create_prorations"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to change plan: %w", err)
	}

	// Return updated subscription
	return subscription.Get(subscriptionID, nil)
}

// Subscription retrieves a subscription by ID.
func (c *Client) Subscription(ctx context.Context, subscriptionID string) (*gostripe.Subscription, error) {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

// BillingPortalURL creates a billing portal session and returns the URL.
func (c *Client) BillingPortalURL(ctx context.Context, customerID, returnURL string) (string, error) {
	params := &gostripe.BillingPortalSessionParams{
		Customer: gostripe.String(customerID),
	}
	if returnURL != "" {
		params.ReturnURL = gostripe.String(returnURL)
	}

	session, err := portalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create billing portal session: %w", err)
	}
	return session.URL, nil
}

// UpcomingInvoice retrieves a preview of the upcoming invoice for a subscription.
func (c *Client) UpcomingInvoice(ctx context.Context, customerID, subscriptionID string) (*gostripe.Invoice, error) {
	params := &gostripe.InvoiceCreatePreviewParams{
		Customer:     gostripe.String(customerID),
		Subscription: gostripe.String(subscriptionID),
	}

	inv, err := invoice.CreatePreview(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming invoice: %w", err)
	}
	return inv, nil
}

// ReportMeterEvent reports a billing meter event for usage-based billing.
// eventName corresponds to a Billing Meter's event_name in Stripe.
// identifier enables Stripe's 24-hour deduplication window — same identifier = rejected as duplicate.
func (c *Client) ReportMeterEvent(ctx context.Context, eventName, customerID string, value int64, timestamp int64, identifier string) error {
	params := &gostripe.BillingMeterEventParams{
		EventName:  gostripe.String(eventName),
		Identifier: gostripe.String(identifier),
		Payload: map[string]string{
			"stripe_customer_id": customerID,
			"value":              fmt.Sprintf("%d", value),
		},
		Timestamp: gostripe.Int64(timestamp),
	}

	_, err := meterevent.New(params)
	if err != nil {
		var stripeErr *gostripe.Error
		if errors.As(err, &stripeErr) && stripeErr.HTTPStatusCode == 400 &&
			strings.Contains(stripeErr.Msg, "An event already exists") {
			slog.Debug("metering: event already exists, skipping", "event", eventName, "identifier", identifier)
			return nil
		}
		return fmt.Errorf("failed to report meter event %q: %w", eventName, err)
	}
	return nil
}

// CreateCreditGrant creates a Stripe Billing Credit Grant for a customer.
// Credits apply to all metered prices and expire at expiresAt (unix timestamp).
func (c *Client) CreateCreditGrant(ctx context.Context, customerID string, amountCents int64, name string, expiresAt int64) error {
	params := &gostripe.BillingCreditGrantParams{
		Customer: gostripe.String(customerID),
		Name:     gostripe.String(name),
		Category: gostripe.String(string(gostripe.BillingCreditGrantCategoryPromotional)),
		Amount: &gostripe.BillingCreditGrantAmountParams{
			Type: gostripe.String(string(gostripe.BillingCreditGrantAmountTypeMonetary)),
			Monetary: &gostripe.BillingCreditGrantAmountMonetaryParams{
				Currency: gostripe.String(string(gostripe.CurrencyEUR)),
				Value:    gostripe.Int64(amountCents),
			},
		},
		ApplicabilityConfig: &gostripe.BillingCreditGrantApplicabilityConfigParams{
			Scope: &gostripe.BillingCreditGrantApplicabilityConfigScopeParams{
				PriceType: gostripe.String("metered"),
			},
		},
		ExpiresAt: gostripe.Int64(expiresAt),
	}

	_, err := creditgrant.New(params)
	if err != nil {
		return fmt.Errorf("failed to create credit grant: %w", err)
	}
	return nil
}

// CreditBalanceCents returns the available credit balance in cents for a customer.
func (c *Client) CreditBalanceCents(ctx context.Context, customerID string) (int64, error) {
	params := &gostripe.BillingCreditBalanceSummaryParams{
		Customer: gostripe.String(customerID),
		Filter: &gostripe.BillingCreditBalanceSummaryFilterParams{
			Type: gostripe.String("applicability_scope"),
			ApplicabilityScope: &gostripe.BillingCreditBalanceSummaryFilterApplicabilityScopeParams{
				PriceType: gostripe.String("metered"),
			},
		},
	}

	summary, err := creditbalancesummary.Get(params)
	if err != nil {
		return 0, fmt.Errorf("failed to get credit balance: %w", err)
	}

	for _, bal := range summary.Balances {
		if bal.AvailableBalance != nil && bal.AvailableBalance.Monetary != nil {
			return bal.AvailableBalance.Monetary.Value, nil
		}
	}
	return 0, nil
}

// ActiveCreditGrantForPeriod checks if a credit grant already exists for the given billing period.
// Returns true if a grant with matching metadata exists.
func (c *Client) ActiveCreditGrantForPeriod(ctx context.Context, customerID string, periodStart int64) (bool, error) {
	params := &gostripe.BillingCreditGrantListParams{
		Customer: gostripe.String(customerID),
	}

	iter := creditgrant.List(params)
	for iter.Next() {
		grant := iter.BillingCreditGrant()
		if grant.Metadata["billing_period_start"] == fmt.Sprintf("%d", periodStart) {
			return true, nil
		}
	}
	if err := iter.Err(); err != nil {
		return false, fmt.Errorf("failed to list credit grants: %w", err)
	}
	return false, nil
}

// CreateCreditGrantForPeriod creates a credit grant for a specific billing period, with idempotency
// via metadata. Returns without error if a grant for this period already exists.
func (c *Client) CreateCreditGrantForPeriod(ctx context.Context, customerID string, amountCents int64, periodStart, periodEnd int64) error {
	exists, err := c.ActiveCreditGrantForPeriod(ctx, customerID, periodStart)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	params := &gostripe.BillingCreditGrantParams{
		Customer: gostripe.String(customerID),
		Name:     gostripe.String("Plan credit"),
		Category: gostripe.String(string(gostripe.BillingCreditGrantCategoryPromotional)),
		Amount: &gostripe.BillingCreditGrantAmountParams{
			Type: gostripe.String(string(gostripe.BillingCreditGrantAmountTypeMonetary)),
			Monetary: &gostripe.BillingCreditGrantAmountMonetaryParams{
				Currency: gostripe.String(string(gostripe.CurrencyEUR)),
				Value:    gostripe.Int64(amountCents),
			},
		},
		ApplicabilityConfig: &gostripe.BillingCreditGrantApplicabilityConfigParams{
			Scope: &gostripe.BillingCreditGrantApplicabilityConfigScopeParams{
				PriceType: gostripe.String("metered"),
			},
		},
		ExpiresAt: gostripe.Int64(periodEnd),
		Metadata: map[string]string{
			"billing_period_start": fmt.Sprintf("%d", periodStart),
		},
	}

	_, err = creditgrant.New(params)
	if err != nil {
		return fmt.Errorf("failed to create credit grant: %w", err)
	}
	return nil
}

// HasPaymentMethod checks if a Stripe customer has a default payment method set.
// This checks the customer's invoice settings, which is where the billing portal sets
// the default card — not on the subscription itself.
func (c *Client) HasPaymentMethod(ctx context.Context, customerID string) (bool, error) {
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return false, fmt.Errorf("failed to get customer: %w", err)
	}
	return cust.InvoiceSettings != nil && cust.InvoiceSettings.DefaultPaymentMethod != nil, nil
}

// PlanPriceID returns the Stripe Price ID for a plan.
func (c *Client) PlanPriceID(plan string) string {
	switch plan {
	case "PRO":
		return c.Prices.ProPriceID
	default:
		return c.Prices.HobbyPriceID
	}
}

// PlanCreditCents returns the monthly credit amount in cents for a plan.
func PlanCreditCents(planPriceID string, prices PriceConfig) int64 {
	switch planPriceID {
	case prices.ProPriceID:
		return 2500 // EUR 25
	default:
		return 500 // EUR 5
	}
}
