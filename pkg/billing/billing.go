// Package billing interprets Stripe state for the platform: which plan a
// customer is on.
package billing

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v87"
)

type Plan string

const (
	PlanNone  Plan = ""
	PlanHobby Plan = "hobby"
	PlanPro   Plan = "pro"
)

func (p Plan) Paid() bool {
	return p != PlanNone
}

type Client struct {
	stripe       *stripe.Client
	hobbyPriceID string
	proPriceID   string
}

func New(secretKey, hobbyPriceID, proPriceID string) *Client {
	return &Client{
		stripe:       stripe.NewClient(secretKey),
		hobbyPriceID: hobbyPriceID,
		proPriceID:   proPriceID,
	}
}

// Plans returns the plan of every customer with a live subscription that
// carries a plan price, keyed by Stripe customer ID. A customer with several
// live subscriptions gets the highest plan among them.
func (c *Client) Plans(ctx context.Context) (map[string]Plan, error) {
	plans := map[string]Plan{}

	for _, plan := range []struct {
		priceID string
		plan    Plan
	}{
		{c.hobbyPriceID, PlanHobby},
		{c.proPriceID, PlanPro},
	} {
		params := &stripe.SubscriptionListParams{Price: stripe.String(plan.priceID)}
		params.Limit = stripe.Int64(100)

		for subscription, err := range c.stripe.V1Subscriptions.List(ctx, params).All(ctx) {
			if err != nil {
				return nil, fmt.Errorf("list subscriptions for price %s: %w", plan.priceID, err)
			}

			if !live(subscription.Status) || subscription.Customer == nil {
				continue
			}

			plans[subscription.Customer.ID] = plan.plan
		}
	}

	return plans, nil
}

func live(status stripe.SubscriptionStatus) bool {
	switch status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing, stripe.SubscriptionStatusPastDue:
		return true
	}

	return false
}
