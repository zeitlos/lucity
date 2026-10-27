// Package billing runs the Stripe side of the billing a user triggers:
// customers, the trial subscription, plan selection and checkout, the
// billing portal, and the numbers on the billing page.
package billing

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/stripe/stripe-go/v87"

	"github.com/zeitlos/lucity/pkg/billing"
)

const trialCreditCents = 500

// ErrSubscriptionEnded is returned when a workspace that already had a
// subscription would get a new trial.
var ErrSubscriptionEnded = errors.New("subscription ended")

type Status string

const (
	StatusActive     Status = "ACTIVE"
	StatusPastDue    Status = "PAST_DUE"
	StatusCanceled   Status = "CANCELED"
	StatusIncomplete Status = "INCOMPLETE"
	StatusTrialing   Status = "TRIALING"
)

// Prices are the Stripe price IDs of the two plans and of every metered
// resource a subscription carries.
type Prices struct {
	Hobby   string
	Pro     string
	Metered []string
}

type Client struct {
	stripe  *stripe.Client
	billing *billing.Client
	prices  Prices
}

func New(secretKey string, prices Prices) *Client {
	return &Client{
		stripe:  stripe.NewClient(secretKey),
		billing: billing.New(secretKey, prices.Hobby, prices.Pro),
		prices:  prices,
	}
}

// Subscription is what the billing page shows about a workspace's
// subscription.
type Subscription struct {
	Plan             billing.Plan
	Status           Status
	CurrentPeriodEnd time.Time
	CreditCents      int64
	CreditExpiry     *time.Time
	HasPaymentMethod bool
}

type Usage struct {
	ResourceCents       int64
	CreditCents         int64
	EstimatedTotalCents int64
}

func (c *Client) CreateCustomer(ctx context.Context, workspace, name, email string) (string, error) {
	customer, err := c.stripe.V1Customers.Create(ctx, &stripe.CustomerCreateParams{
		Name:     stripe.String(name),
		Email:    stripe.String(email),
		Metadata: map[string]string{"workspace": workspace},
	})

	if err != nil {
		return "", fmt.Errorf("create customer: %w", err)
	}

	return customer.ID, nil
}

// StartTrial gives a customer that never had a subscription the metered trial
// subscription, with creditDays of trial credit. A customer whose
// subscription ended gets ErrSubscriptionEnded, never a second trial.
func (c *Client) StartTrial(ctx context.Context, customerID, workspace string, creditDays int) error {
	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil {
		return err
	}

	if current != nil {
		if billing.Ended(current) {
			return ErrSubscriptionEnded
		}

		return nil
	}

	params := &stripe.SubscriptionCreateParams{
		Customer:        stripe.String(customerID),
		PaymentBehavior: stripe.String("allow_incomplete"),
		BillingMode:     &stripe.SubscriptionCreateBillingModeParams{Type: stripe.String(string(stripe.SubscriptionBillingModeTypeClassic))},
		Items:           c.items(""),
		Metadata:        map[string]string{"workspace": workspace},
	}

	if creditDays > 0 {
		params.BillingThresholds = &stripe.SubscriptionCreateBillingThresholdsParams{AmountGTE: stripe.Int64(trialCreditCents)}
		params.PendingInvoiceItemInterval = &stripe.SubscriptionCreatePendingInvoiceItemIntervalParams{
			Interval:      stripe.String("day"),
			IntervalCount: stripe.Int64(int64(creditDays)),
		}
	}

	if _, err := c.stripe.V1Subscriptions.Create(ctx, params); err != nil {
		return fmt.Errorf("create trial subscription: %w", err)
	}

	if creditDays > 0 {
		expiresAt := time.Now().Add(time.Duration(creditDays)*24*time.Hour + time.Hour).Unix()

		if err := c.grantTrialCredit(ctx, customerID, expiresAt); err != nil {
			slog.WarnContext(ctx, "billing: failed to grant trial credit", "customer", customerID, "error", err)
		}
	}

	return nil
}

// Subscription returns the subscription a customer is billed through, or nil
// when it never had one.
func (c *Client) Subscription(ctx context.Context, customerID string) (*Subscription, error) {
	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil || current == nil {
		return nil, err
	}

	hasPaymentMethod, err := c.hasPaymentMethod(ctx, customerID, current)

	if err != nil {
		return nil, err
	}

	credit, err := c.creditBalance(ctx, customerID)

	if err != nil {
		slog.WarnContext(ctx, "billing: failed to read credit balance", "customer", customerID, "error", err)
	}

	plan := c.billing.Plan(current)
	view := &Subscription{
		Plan:             plan,
		Status:           status(current, plan),
		CurrentPeriodEnd: periodEnd(current),
		CreditCents:      credit,
		HasPaymentMethod: hasPaymentMethod,
	}

	if !hasPaymentMethod {
		if expiry, err := c.creditExpiry(ctx, customerID); err != nil {
			slog.WarnContext(ctx, "billing: failed to read credit expiry", "customer", customerID, "error", err)
		} else if expiry.After(time.Now()) {
			view.CreditExpiry = &expiry
		}
	}

	return view, nil
}

// Usage returns the costs of the current billing period so far, or nil when
// the customer never had a subscription.
func (c *Client) Usage(ctx context.Context, customerID string) (*Usage, error) {
	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil || current == nil {
		return nil, err
	}

	var resource, plan int64

	preview, err := c.stripe.V1Invoices.CreatePreview(ctx, &stripe.InvoiceCreatePreviewParams{
		Customer:     stripe.String(customerID),
		Subscription: stripe.String(current.ID),
	})

	if err != nil {
		slog.DebugContext(ctx, "billing: no invoice preview", "customer", customerID, "error", err)
	} else if preview.Lines != nil {
		for _, line := range preview.Lines.Data {
			if line.Pricing == nil || line.Pricing.PriceDetails == nil || line.Pricing.PriceDetails.Price == nil {
				continue
			}

			switch line.Pricing.PriceDetails.Price.ID {
			case c.prices.Hobby, c.prices.Pro:
				plan = line.Amount
			default:
				resource += line.Amount
			}
		}
	}

	unpaid, err := c.unpaidTotal(ctx, current.ID)

	if err != nil {
		slog.WarnContext(ctx, "billing: failed to read unpaid invoices", "customer", customerID, "error", err)
	}

	resource += unpaid

	credit, err := c.creditBalance(ctx, customerID)

	if err != nil {
		slog.WarnContext(ctx, "billing: failed to read credit balance", "customer", customerID, "error", err)
	}

	applied := min(credit, resource)

	return &Usage{
		ResourceCents:       resource,
		CreditCents:         applied,
		EstimatedTotalCents: plan + resource - applied,
	}, nil
}

func (c *Client) PortalURL(ctx context.Context, customerID, returnURL string) (string, error) {
	session, err := c.stripe.V1BillingPortalSessions.Create(ctx, &stripe.BillingPortalSessionCreateParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	})

	if err != nil {
		return "", fmt.Errorf("create billing portal session: %w", err)
	}

	return session.URL, nil
}

// Paid reports whether the customer's subscription is live and carries a
// plan.
func (c *Client) Paid(ctx context.Context, customerID string) (bool, error) {
	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil || current == nil {
		return false, err
	}

	return c.billing.Paid(current), nil
}

// Standing reports whether the customer is in good standing, by the same rule
// stagehand suspends on. The second result is false when the customer never
// had a subscription, which stagehand leaves alone.
func (c *Client) Standing(ctx context.Context, customerID string) (billing.Standing, bool, error) {
	current, err := c.billing.Subscription(ctx, customerID)

	if err != nil || current == nil {
		return billing.Standing{}, false, err
	}

	customer, err := c.stripe.V1Customers.Retrieve(ctx, customerID, nil)

	if err != nil {
		return billing.Standing{}, false, fmt.Errorf("read customer: %w", err)
	}

	return c.billing.Standing(current, invoiceDefault(customer), time.Now()), true, nil
}

func (c *Client) items(planPrice string) []*stripe.SubscriptionCreateItemParams {
	var items []*stripe.SubscriptionCreateItemParams

	if planPrice != "" {
		items = append(items, &stripe.SubscriptionCreateItemParams{Price: stripe.String(planPrice)})
	}

	for _, price := range c.prices.Metered {
		items = append(items, &stripe.SubscriptionCreateItemParams{Price: stripe.String(price)})
	}

	return items
}

func (c *Client) hasPaymentMethod(ctx context.Context, customerID string, subscription *stripe.Subscription) (bool, error) {
	if subscription != nil && subscription.DefaultPaymentMethod != nil {
		return true, nil
	}

	customer, err := c.stripe.V1Customers.Retrieve(ctx, customerID, nil)

	if err != nil {
		return false, fmt.Errorf("read customer: %w", err)
	}

	return invoiceDefault(customer), nil
}

func (c *Client) grantTrialCredit(ctx context.Context, customerID string, expiresAt int64) error {
	_, err := c.stripe.V1BillingCreditGrants.Create(ctx, &stripe.BillingCreditGrantCreateParams{
		Customer: stripe.String(customerID),
		Name:     stripe.String("Trial credit"),
		Category: stripe.String(string(stripe.BillingCreditGrantCategoryPromotional)),
		Amount: &stripe.BillingCreditGrantCreateAmountParams{
			Type: stripe.String(string(stripe.BillingCreditGrantAmountTypeMonetary)),
			Monetary: &stripe.BillingCreditGrantCreateAmountMonetaryParams{
				Currency: stripe.String(string(stripe.CurrencyEUR)),
				Value:    stripe.Int64(trialCreditCents),
			},
		},
		ApplicabilityConfig: &stripe.BillingCreditGrantCreateApplicabilityConfigParams{
			Scope: &stripe.BillingCreditGrantCreateApplicabilityConfigScopeParams{PriceType: stripe.String("metered")},
		},
		ExpiresAt: stripe.Int64(expiresAt),
	})

	return err
}

func (c *Client) creditBalance(ctx context.Context, customerID string) (int64, error) {
	summary, err := c.stripe.V1BillingCreditBalanceSummary.Retrieve(ctx, &stripe.BillingCreditBalanceSummaryRetrieveParams{
		Customer: stripe.String(customerID),
		Filter: &stripe.BillingCreditBalanceSummaryRetrieveFilterParams{
			Type: stripe.String("applicability_scope"),
			ApplicabilityScope: &stripe.BillingCreditBalanceSummaryRetrieveFilterApplicabilityScopeParams{
				PriceType: stripe.String("metered"),
			},
		},
	})

	if err != nil {
		return 0, err
	}

	for _, balance := range summary.Balances {
		if balance.AvailableBalance != nil && balance.AvailableBalance.Monetary != nil {
			return balance.AvailableBalance.Monetary.Value, nil
		}
	}

	return 0, nil
}

func (c *Client) creditExpiry(ctx context.Context, customerID string) (time.Time, error) {
	var earliest int64
	now := time.Now().Unix()

	for grant, err := range c.stripe.V1BillingCreditGrants.List(ctx, &stripe.BillingCreditGrantListParams{Customer: stripe.String(customerID)}).All(ctx) {
		if err != nil {
			return time.Time{}, err
		}

		if grant.VoidedAt > 0 || grant.ExpiresAt <= now {
			continue
		}

		if earliest == 0 || grant.ExpiresAt < earliest {
			earliest = grant.ExpiresAt
		}
	}

	if earliest == 0 {
		return time.Time{}, nil
	}

	return time.Unix(earliest, 0), nil
}

func (c *Client) unpaidTotal(ctx context.Context, subscriptionID string) (int64, error) {
	var total int64

	for _, invoiceStatus := range []string{"draft", "open"} {
		params := &stripe.InvoiceListParams{
			Subscription: stripe.String(subscriptionID),
			Status:       stripe.String(invoiceStatus),
		}

		for invoice, err := range c.stripe.V1Invoices.List(ctx, params).All(ctx) {
			if err != nil {
				return 0, err
			}

			total += invoice.AmountDue
		}
	}

	return total, nil
}

func invoiceDefault(customer *stripe.Customer) bool {
	return customer.InvoiceSettings != nil && customer.InvoiceSettings.DefaultPaymentMethod != nil
}

func status(subscription *stripe.Subscription, plan billing.Plan) Status {
	switch subscription.Status {
	case stripe.SubscriptionStatusActive:
		if !plan.Paid() {
			return StatusTrialing
		}

		return StatusActive
	case stripe.SubscriptionStatusTrialing:
		return StatusTrialing
	case stripe.SubscriptionStatusPastDue, stripe.SubscriptionStatusUnpaid:
		return StatusPastDue
	case stripe.SubscriptionStatusIncomplete:
		return StatusIncomplete
	}

	return StatusCanceled
}

func periodEnd(subscription *stripe.Subscription) time.Time {
	if subscription.Items != nil {
		for _, item := range subscription.Items.Data {
			if item.CurrentPeriodEnd > 0 {
				return time.Unix(item.CurrentPeriodEnd, 0)
			}
		}
	}

	return time.Time{}
}
