package stripe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	gostripe "github.com/stripe/stripe-go/v87"
	"github.com/stripe/stripe-go/v87/billing/creditgrant"
	"github.com/stripe/stripe-go/v87/billing/meterevent"
)

// PriceConfig holds the Stripe Price IDs of the two plans.
type PriceConfig struct {
	HobbyPriceID string
	ProPriceID   string
}

// MeterConfig holds Billing Meter event names for usage-based billing.
type MeterConfig struct {
	EcoCPUEventName  string
	EcoMemEventName  string
	EcoDiskEventName string
	StorageEventName string
	EgressEventName  string
}

// Client wraps the Stripe API for metering and plan credits.
type Client struct {
	Prices PriceConfig
	Meters MeterConfig
}

// NewClient creates a Stripe client and sets the global API key.
func NewClient(secretKey string, prices PriceConfig, meters MeterConfig) *Client {
	gostripe.Key = secretKey
	return &Client{Prices: prices, Meters: meters}
}

const planCreditExpiryGrace = 24 * time.Hour

// ReportMeterEvent reports a billing meter event for usage-based billing.
// eventName corresponds to a Billing Meter's event_name in Stripe.
// identifier enables Stripe's 24-hour deduplication window — same identifier = rejected as duplicate.
func (c *Client) ReportMeterEvent(ctx context.Context, eventName, customerID string, value float64, timestamp int64, identifier string) error {
	params := &gostripe.BillingMeterEventParams{
		EventName:  gostripe.String(eventName),
		Identifier: gostripe.String(identifier),
		Payload: map[string]string{
			"stripe_customer_id": customerID,
			"value":              strconv.FormatFloat(value, 'f', -1, 64),
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
		ExpiresAt: gostripe.Int64(periodEnd + int64(planCreditExpiryGrace.Seconds())),
		Metadata: map[string]string{
			"billing_period_start": fmt.Sprintf("%d", periodStart),
		},
	}
	// Idempotency key prevents duplicate grants when multiple workers race
	// (e.g. during rollouts or local dev with two air instances).
	params.IdempotencyKey = gostripe.String(fmt.Sprintf("credit-grant:%s:%d", customerID, periodStart))

	_, err = creditgrant.New(params)
	if err != nil {
		return fmt.Errorf("failed to create credit grant: %w", err)
	}
	return nil
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
