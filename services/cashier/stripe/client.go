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

// TrialCreditCents is the fixed trial credit amount (EUR 5).
const TrialCreditCents = 500

// CreateSubscription creates a subscription with only metered resource line items.
// No plan is included — the user adds a plan later via a setup checkout + AddPlan.
// creditDays > 0 creates a promotional trial credit grant that expires after that many days.
func (c *Client) CreateSubscription(ctx context.Context, customerID, workspace string, creditDays int) (string, error) {
	params := &gostripe.SubscriptionParams{
		Customer:        gostripe.String(customerID),
		PaymentBehavior: gostripe.String("allow_incomplete"),
		Items: []*gostripe.SubscriptionItemsParams{
			{Price: gostripe.String(c.Prices.EcoCPUPriceID)},
			{Price: gostripe.String(c.Prices.EcoMemPriceID)},
			{Price: gostripe.String(c.Prices.EcoDiskPriceID)},
			{Price: gostripe.String(c.Prices.ProdCPUPriceID)},
			{Price: gostripe.String(c.Prices.ProdMemPriceID)},
			{Price: gostripe.String(c.Prices.ProdDiskPriceID)},
		},
	}
	params.AddMetadata("workspace", workspace)

	// Trial billing: threshold invoice at €5 usage, interval invoice at creditDays.
	// Whichever fires first ends the trial via handlePaymentSucceeded.
	if creditDays > 0 {
		params.BillingThresholds = &gostripe.SubscriptionBillingThresholdsParams{
			AmountGTE: gostripe.Int64(int64(TrialCreditCents)),
		}
		params.PendingInvoiceItemInterval = &gostripe.SubscriptionPendingInvoiceItemIntervalParams{
			Interval:      gostripe.String("day"),
			IntervalCount: gostripe.Int64(int64(creditDays)),
		}
	}

	sub, err := subscription.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe subscription: %w", err)
	}

	// Create trial credit grant so credits exist from day one.
	// The 1-hour buffer past creditDays ensures credits are still valid when the
	// interval invoice finalizes at exactly day N (prevents race with expiry).
	if creditDays > 0 {
		expiresAt := time.Now().Add(time.Duration(creditDays)*24*time.Hour + time.Hour).Unix()
		if err := c.CreateCreditGrant(ctx, customerID, TrialCreditCents, "Trial credit", expiresAt); err != nil {
			slog.Warn("failed to create trial credit grant", "error", err)
		}
	}

	return sub.ID, nil
}

// ChangePlan swaps the plan subscription item from one price to another.
// Requires a plan to already exist on the subscription (use AddPlanToSubscription for first-time).
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
		return nil, fmt.Errorf("no plan on subscription: add a plan first")
	}

	// Update the plan item's price
	_, err = subscriptionitem.Update(planItemID, &gostripe.SubscriptionItemParams{
		Price:             gostripe.String(newPlanPriceID),
		ProrationBehavior: gostripe.String("create_prorations"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to change plan: %w", err)
	}

	return subscription.Get(subscriptionID, nil)
}

// AddPlanToSubscription adds a plan price as a new subscription item.
// Used when a trial user adds their first plan after completing a setup checkout.
func (c *Client) AddPlanToSubscription(ctx context.Context, subscriptionID, planPriceID string) (*gostripe.Subscription, error) {
	_, err := subscriptionitem.New(&gostripe.SubscriptionItemParams{
		Subscription:      gostripe.String(subscriptionID),
		Price:             gostripe.String(planPriceID),
		ProrationBehavior: gostripe.String("create_prorations"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add plan to subscription: %w", err)
	}

	return subscription.Get(subscriptionID, nil)
}

// ClearTrialBillingParams removes billing thresholds and pending invoice item interval
// from a subscription, and resets the billing cycle anchor to now. Called when a trial
// user converts to a paid plan so the paid billing period starts fresh.
func (c *Client) ClearTrialBillingParams(ctx context.Context, subscriptionID string) error {
	params := &gostripe.SubscriptionParams{
		BillingCycleAnchor: gostripe.Int64(time.Now().Unix()),
		ProrationBehavior:  gostripe.String("create_prorations"),
	}
	// Empty string tells Stripe to null these nested objects.
	// Struct fields with nil pointers are omitted from the request, so we use AddExtra.
	params.AddExtra("billing_thresholds", "")
	params.AddExtra("pending_invoice_item_interval", "")

	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to clear trial billing params: %w", err)
	}
	return nil
}

// SetDefaultPaymentMethod sets the customer's default payment method for invoices.
func (c *Client) SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	_, err := customer.Update(customerID, &gostripe.CustomerParams{
		InvoiceSettings: &gostripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: gostripe.String(paymentMethodID),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to set default payment method: %w", err)
	}
	return nil
}

// CreateSetupCheckoutSession creates a Stripe Checkout Session in setup mode
// to collect a payment method for an existing customer. The plan is stored in
// session metadata so it can be retrieved on completion.
func (c *Client) CreateSetupCheckoutSession(ctx context.Context, customerID, planPriceID, planName, successURL, cancelURL string) (string, string, error) {
	params := &gostripe.CheckoutSessionParams{
		Mode:     gostripe.String(string(gostripe.CheckoutSessionModeSetup)),
		Customer: gostripe.String(customerID),
		Currency: gostripe.String(string(gostripe.CurrencyEUR)),
		SuccessURL: gostripe.String(successURL),
		CancelURL:  gostripe.String(cancelURL),
		CustomText: &gostripe.CheckoutSessionCustomTextParams{
			Submit: &gostripe.CheckoutSessionCustomTextSubmitParams{
				Message: gostripe.String(fmt.Sprintf("You're signing up for the %s plan. Your card will be charged at the end of each billing cycle.", planName)),
			},
		},
	}
	params.AddMetadata("plan_price_id", planPriceID)

	session, err := checkoutsession.New(params)
	if err != nil {
		return "", "", fmt.Errorf("failed to create setup checkout session: %w", err)
	}
	return session.URL, session.ID, nil
}

// RetrieveSetupCheckoutSession retrieves a setup-mode checkout session and
// returns the payment method ID from the completed SetupIntent.
func (c *Client) RetrieveSetupCheckoutSession(ctx context.Context, sessionID string) (paymentMethodID string, planPriceID string, err error) {
	params := &gostripe.CheckoutSessionParams{}
	params.AddExpand("setup_intent")
	params.AddExpand("setup_intent.payment_method")

	session, err := checkoutsession.Get(sessionID, params)
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve setup checkout session: %w", err)
	}

	if session.Status != gostripe.CheckoutSessionStatusComplete {
		return "", "", fmt.Errorf("checkout session not complete (status: %s)", session.Status)
	}

	if session.SetupIntent == nil || session.SetupIntent.PaymentMethod == nil {
		return "", "", fmt.Errorf("checkout session has no payment method")
	}

	return session.SetupIntent.PaymentMethod.ID, session.Metadata["plan_price_id"], nil
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

// WorkspaceBilling holds the Stripe customer and subscription IDs for a billable workspace.
type WorkspaceBilling struct {
	CustomerID     string
	SubscriptionID string
}

// BillableWorkspaces returns all Stripe customers with workspace metadata and their active subscription IDs.
func (c *Client) BillableWorkspaces(ctx context.Context) (map[string]WorkspaceBilling, error) {
	result := make(map[string]WorkspaceBilling)
	params := &gostripe.CustomerSearchParams{
		SearchParams: gostripe.SearchParams{
			Query: `metadata["workspace"]:"*"`,
		},
	}

	iter := customer.Search(params)
	for iter.Next() {
		cust := iter.Customer()
		ws := cust.Metadata["workspace"]
		if ws == "" {
			continue
		}
		// Find active subscription
		subID := ""
		subParams := &gostripe.SubscriptionListParams{
			Customer: gostripe.String(cust.ID),
			Status:   gostripe.String("active"),
		}
		subIter := subscription.List(subParams)
		if subIter.Next() {
			subID = subIter.Subscription().ID
		}
		if subID == "" {
			continue
		}
		result[ws] = WorkspaceBilling{
			CustomerID:     cust.ID,
			SubscriptionID: subID,
		}
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	return result, nil
}

// CustomerByWorkspace searches for an existing Stripe customer with matching workspace metadata.
// Returns the customer ID if found, empty string if not. Used for idempotent customer creation.
func (c *Client) CustomerByWorkspace(ctx context.Context, workspace string) (string, error) {
	params := &gostripe.CustomerSearchParams{
		SearchParams: gostripe.SearchParams{
			Query: fmt.Sprintf(`metadata["workspace"]:"%s"`, workspace),
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
