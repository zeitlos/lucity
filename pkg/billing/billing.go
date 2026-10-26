// Package billing interprets Stripe state for the platform: which subscription
// a customer is billed through, and which plan it carries.
package billing

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"

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

func (p Plan) rank() int {
	switch p {
	case PlanPro:
		return 2
	case PlanHobby:
		return 1
	}

	return 0
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

// Subscription returns the subscription a customer is billed through: the
// best current one, or the newest ended one when every subscription has
// ended. It returns nil only for a customer that never had a subscription.
func (c *Client) Subscription(ctx context.Context, customerID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
		Status:   stripe.String("all"),
	}
	params.Limit = stripe.Int64(100)

	var subscriptions []*stripe.Subscription

	for subscription, err := range c.stripe.V1Subscriptions.List(ctx, params).All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list subscriptions of customer %s: %w", customerID, err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	return c.pick(ctx, customerID, subscriptions), nil
}

// CurrentSubscriptions returns the subscription of every customer that has a
// current one, keyed by Stripe customer ID, picked the same way as
// Subscription.
func (c *Client) CurrentSubscriptions(ctx context.Context) (map[string]*stripe.Subscription, error) {
	params := &stripe.SubscriptionListParams{}
	params.Limit = stripe.Int64(100)

	byCustomer := map[string][]*stripe.Subscription{}

	for subscription, err := range c.stripe.V1Subscriptions.List(ctx, params).All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list subscriptions: %w", err)
		}

		if subscription.Customer == nil || Ended(subscription) {
			continue
		}

		byCustomer[subscription.Customer.ID] = append(byCustomer[subscription.Customer.ID], subscription)
	}

	current := make(map[string]*stripe.Subscription, len(byCustomer))

	for customerID, subscriptions := range byCustomer {
		current[customerID] = c.pick(ctx, customerID, subscriptions)
	}

	return current, nil
}

// Plans returns the plan of every customer whose subscription is live and
// carries a plan price, keyed by Stripe customer ID.
func (c *Client) Plans(ctx context.Context) (map[string]Plan, error) {
	subscriptions, err := c.CurrentSubscriptions(ctx)

	if err != nil {
		return nil, err
	}

	plans := map[string]Plan{}

	for customerID, subscription := range subscriptions {
		if !live(subscription.Status) {
			continue
		}

		if plan := c.Plan(subscription); plan.Paid() {
			plans[customerID] = plan
		}
	}

	return plans, nil
}

func (c *Client) Plan(subscription *stripe.Subscription) Plan {
	plan := PlanNone

	if subscription.Items == nil {
		return plan
	}

	for _, item := range subscription.Items.Data {
		if item.Price == nil {
			continue
		}

		switch item.Price.ID {
		case c.proPriceID:
			return PlanPro
		case c.hobbyPriceID:
			plan = PlanHobby
		}
	}

	return plan
}

// Ended reports whether a subscription can no longer bill.
func Ended(subscription *stripe.Subscription) bool {
	switch subscription.Status {
	case stripe.SubscriptionStatusCanceled, stripe.SubscriptionStatusIncompleteExpired:
		return true
	}

	return false
}

func (c *Client) pick(ctx context.Context, customerID string, subscriptions []*stripe.Subscription) *stripe.Subscription {
	if len(subscriptions) == 0 {
		return nil
	}

	picked := slices.MaxFunc(subscriptions, c.compare)

	var current []string

	for _, subscription := range subscriptions {
		if !Ended(subscription) {
			current = append(current, subscription.ID)
		}
	}

	if len(current) > 1 {
		slog.WarnContext(ctx, "billing: customer has several current subscriptions", "customer", customerID, "subscriptions", current, "picked", picked.ID)
	}

	return picked
}

func (c *Client) compare(a, b *stripe.Subscription) int {
	return cmp.Or(
		cmp.Compare(score(!Ended(a)), score(!Ended(b))),
		cmp.Compare(score(live(a.Status)), score(live(b.Status))),
		cmp.Compare(c.Plan(a).rank(), c.Plan(b).rank()),
		cmp.Compare(a.Created, b.Created),
	)
}

func score(ok bool) int {
	if ok {
		return 1
	}

	return 0
}

func live(status stripe.SubscriptionStatus) bool {
	switch status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing, stripe.SubscriptionStatusPastDue:
		return true
	}

	return false
}
