package stripe

import (
	"context"
	"fmt"
	"time"

	gostripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/billing/meterevent"
	portalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/invoice"
	"github.com/stripe/stripe-go/v82/invoiceitem"
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

// CreateSubscription creates a subscription with a plan + 6 resource line items.
// Metered items (eco) start with no quantity. Licensed items (production) start at 0.
// trialDays > 0 starts the subscription with a free trial period.
func (c *Client) CreateSubscription(ctx context.Context, customerID, workspace string, planPriceID string, trialDays int) (string, error) {
	params := &gostripe.SubscriptionParams{
		Customer:        gostripe.String(customerID),
		PaymentBehavior: gostripe.String("default_incomplete"),
		Items: []*gostripe.SubscriptionItemsParams{
			{Price: gostripe.String(planPriceID)},
			{Price: gostripe.String(c.Prices.EcoCPUPriceID)},
			{Price: gostripe.String(c.Prices.EcoMemPriceID)},
			{Price: gostripe.String(c.Prices.EcoDiskPriceID)},
			{Price: gostripe.String(c.Prices.ProdCPUPriceID), Quantity: gostripe.Int64(0)},
			{Price: gostripe.String(c.Prices.ProdMemPriceID), Quantity: gostripe.Int64(0)},
			{Price: gostripe.String(c.Prices.ProdDiskPriceID), Quantity: gostripe.Int64(0)},
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

// UpcomingInvoice retrieves a preview of the upcoming invoice for a customer (usage summary).
func (c *Client) UpcomingInvoice(ctx context.Context, customerID string) (*gostripe.Invoice, error) {
	params := &gostripe.InvoiceCreatePreviewParams{
		Customer: gostripe.String(customerID),
	}

	inv, err := invoice.CreatePreview(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming invoice: %w", err)
	}
	return inv, nil
}

// AddInvoiceCredit adds a negative invoice item (credit) to a customer's next invoice.
func (c *Client) AddInvoiceCredit(ctx context.Context, customerID string, amountCents int64, description string) error {
	params := &gostripe.InvoiceItemParams{
		Customer:    gostripe.String(customerID),
		Amount:      gostripe.Int64(-amountCents),
		Currency:    gostripe.String(string(gostripe.CurrencyEUR)),
		Description: gostripe.String(description),
	}

	_, err := invoiceitem.New(params)
	if err != nil {
		return fmt.Errorf("failed to add invoice credit: %w", err)
	}
	return nil
}

// ReportMeterEvent reports a billing meter event for usage-based billing.
// eventName corresponds to a Billing Meter's event_name in Stripe.
func (c *Client) ReportMeterEvent(ctx context.Context, eventName, customerID string, value int64, timestamp int64) error {
	params := &gostripe.BillingMeterEventParams{
		EventName: gostripe.String(eventName),
		Payload: map[string]string{
			"stripe_customer_id": customerID,
			"value":              fmt.Sprintf("%d", value),
		},
		Timestamp: gostripe.Int64(timestamp),
	}

	_, err := meterevent.New(params)
	if err != nil {
		return fmt.Errorf("failed to report meter event %q: %w", eventName, err)
	}
	return nil
}

// UpdateItemQuantity updates the quantity on a licensed subscription item.
func (c *Client) UpdateItemQuantity(ctx context.Context, subscriptionItemID string, quantity int64) error {
	_, err := subscriptionitem.Update(subscriptionItemID, &gostripe.SubscriptionItemParams{
		Quantity:          gostripe.Int64(quantity),
		ProrationBehavior: gostripe.String("create_prorations"),
	})
	if err != nil {
		return fmt.Errorf("failed to update item quantity: %w", err)
	}
	return nil
}

// FindItemByPrice returns the subscription item ID for a given price ID, or empty string if not found.
func FindItemByPrice(sub *gostripe.Subscription, priceID string) string {
	for _, item := range sub.Items.Data {
		if item.Price.ID == priceID {
			return item.ID
		}
	}
	return ""
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
