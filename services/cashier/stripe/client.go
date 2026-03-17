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
	checkoutsession "github.com/stripe/stripe-go/v82/checkout/session"
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
// creditDays > 0 creates a promotional credit grant that expires after that many days.
// After creation, an initial credit grant is created for the billing period.
func (c *Client) CreateSubscription(ctx context.Context, customerID, workspace string, planPriceID string, creditDays int) (string, error) {
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
	params.AddMetadata("workspace", workspace)

	sub, err := subscription.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe subscription: %w", err)
	}

	// Create initial credit grant so credits exist from day one.
	if len(sub.Items.Data) > 0 {
		creditCents := PlanCreditCents(planPriceID, c.Prices)
		expiresAt := sub.Items.Data[0].CurrentPeriodEnd
		name := "Plan credit"

		if creditDays > 0 {
			// Promotional credit for new signups — expires after creditDays.
			expiresAt = time.Now().Add(time.Duration(creditDays) * 24 * time.Hour).Unix()
			name = "Signup credit"
		}

		if err := c.CreateCreditGrant(ctx, customerID, creditCents, name, expiresAt); err != nil {
			// Non-fatal — metering worker will create it on next tick if missing.
			slog.Warn("failed to create initial credit grant", "error", err)
		}
	}

	return sub.ID, nil
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

// CreditGrantExpiry returns the expires_at timestamp of the earliest active credit grant
// for a customer. Returns 0 if no active grants exist.
func (c *Client) CreditGrantExpiry(ctx context.Context, customerID string) (int64, error) {
	params := &gostripe.BillingCreditGrantListParams{
		Customer: gostripe.String(customerID),
	}

	var earliest int64
	now := time.Now().Unix()
	iter := creditgrant.List(params)
	for iter.Next() {
		grant := iter.BillingCreditGrant()
		if grant.VoidedAt > 0 || grant.ExpiresAt <= now {
			continue
		}
		if earliest == 0 || grant.ExpiresAt < earliest {
			earliest = grant.ExpiresAt
		}
	}
	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to list credit grants: %w", err)
	}
	return earliest, nil
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

// CustomerByWorkspace searches for an existing Stripe customer with matching workspace metadata.
// Returns the customer ID if found, empty string if not. Used for idempotent customer creation.
func (c *Client) CustomerByWorkspace(ctx context.Context, workspace string) (string, error) {
	params := &gostripe.CustomerSearchParams{
		SearchParams: gostripe.SearchParams{
			Query: fmt.Sprintf("metadata['workspace']:'%s'", workspace),
		},
	}

	iter := customer.Search(params)
	if iter.Next() {
		return iter.Customer().ID, nil
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("failed to search customers: %w", err)
	}
	return "", nil
}

// ActiveSubscriptionForCustomer returns the ID of an active subscription
// for the given customer that has the specified workspace in its metadata.
// Returns empty string if no matching subscription exists.
func (c *Client) ActiveSubscriptionForCustomer(ctx context.Context, customerID, workspace string) (string, error) {
	params := &gostripe.SubscriptionListParams{
		ListParams: gostripe.ListParams{},
	}
	params.Filters.AddFilter("customer", "", customerID)
	params.Filters.AddFilter("status", "", "active")

	iter := subscription.List(params)
	for iter.Next() {
		sub := iter.Subscription()
		if sub.Metadata["workspace"] == workspace {
			return sub.ID, nil
		}
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return "", nil
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

// CreateCheckoutSession creates a Stripe Checkout Session in subscription mode
// with the plan price + 6 metered resource prices.
func (c *Client) CreateCheckoutSession(ctx context.Context, workspace, name, planPriceID, email, successURL, cancelURL, userID string) (string, string, error) {
	params := &gostripe.CheckoutSessionParams{
		Mode:                    gostripe.String(string(gostripe.CheckoutSessionModeSubscription)),
		CustomerEmail:           gostripe.String(email),
		SuccessURL:              gostripe.String(successURL),
		CancelURL:               gostripe.String(cancelURL),
		PaymentMethodCollection: gostripe.String("always"),
		LineItems: []*gostripe.CheckoutSessionLineItemParams{
			{Price: gostripe.String(planPriceID), Quantity: gostripe.Int64(1)},
			{Price: gostripe.String(c.Prices.EcoCPUPriceID)},
			{Price: gostripe.String(c.Prices.EcoMemPriceID)},
			{Price: gostripe.String(c.Prices.EcoDiskPriceID)},
			{Price: gostripe.String(c.Prices.ProdCPUPriceID)},
			{Price: gostripe.String(c.Prices.ProdMemPriceID)},
			{Price: gostripe.String(c.Prices.ProdDiskPriceID)},
		},
		SubscriptionData: &gostripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"workspace": workspace,
			},
		},
	}
	params.AddMetadata("workspace_id", workspace)
	params.AddMetadata("workspace_name", name)
	params.AddMetadata("user_id", userID)

	session, err := checkoutsession.New(params)
	if err != nil {
		return "", "", fmt.Errorf("failed to create checkout session: %w", err)
	}
	return session.URL, session.ID, nil
}

// RetrieveCheckoutSession retrieves a Checkout Session and extracts customer/subscription IDs.
func (c *Client) RetrieveCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSessionResult, error) {
	session, err := checkoutsession.Get(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve checkout session: %w", err)
	}

	result := &CheckoutSessionResult{
		SessionID: session.ID,
		Status:    string(session.Status),
	}

	if session.Customer != nil {
		result.CustomerID = session.Customer.ID
	}
	if session.Subscription != nil {
		result.SubscriptionID = session.Subscription.ID
	}

	result.Workspace = session.Metadata["workspace_id"]
	result.Name = session.Metadata["workspace_name"]
	result.UserID = session.Metadata["user_id"]
	result.Email = session.CustomerEmail

	return result, nil
}

// CheckoutSessionResult holds the relevant fields from a completed Checkout Session.
type CheckoutSessionResult struct {
	SessionID      string
	Status         string
	Workspace      string
	Name           string
	UserID         string
	Email          string
	CustomerID     string
	SubscriptionID string
}
