// Package billing interprets Stripe state for the platform: which subscription
// a customer is billed through, and which plan it carries.
package billing

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

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

// Subscriptions returns the subscription of every customer that ever had
// one, keyed by Stripe customer ID, picked the same way as Subscription.
func (c *Client) Subscriptions(ctx context.Context) (map[string]*stripe.Subscription, error) {
	params := &stripe.SubscriptionListParams{Status: stripe.String("all")}
	params.Limit = stripe.Int64(100)

	byCustomer := map[string][]*stripe.Subscription{}

	for subscription, err := range c.stripe.V1Subscriptions.List(ctx, params).All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list subscriptions: %w", err)
		}

		if subscription.Customer == nil {
			continue
		}

		byCustomer[subscription.Customer.ID] = append(byCustomer[subscription.Customer.ID], subscription)
	}

	picked := make(map[string]*stripe.Subscription, len(byCustomer))

	for customerID, subscriptions := range byCustomer {
		picked[customerID] = c.pick(ctx, customerID, subscriptions)
	}

	return picked, nil
}

// PaymentMethods reports, keyed by Stripe customer ID, whether a customer has
// a default payment method for its invoices.
func (c *Client) PaymentMethods(ctx context.Context) (map[string]bool, error) {
	params := &stripe.CustomerListParams{}
	params.Limit = stripe.Int64(100)

	methods := map[string]bool{}

	for customer, err := range c.stripe.V1Customers.List(ctx, params).All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list customers: %w", err)
		}

		methods[customer.ID] = customer.InvoiceSettings != nil && customer.InvoiceSettings.DefaultPaymentMethod != nil
	}

	return methods, nil
}

// Paid reports whether a subscription is live and carries a plan price.
func (c *Client) Paid(subscription *stripe.Subscription) bool {
	return live(subscription.Status) && c.Plan(subscription).Paid()
}

type Standing struct {
	Good   bool
	Reason string
}

// Standing reports whether a workspace billed through the subscription is in
// good standing. A past_due subscription still is, as a grace period, and so
// is a trial until it expires without a payment method.
func (c *Client) Standing(subscription *stripe.Subscription, hasPaymentMethod bool, now time.Time) Standing {
	switch {
	case Ended(subscription):
		return Standing{Reason: "ended"}
	case subscription.Status == stripe.SubscriptionStatusUnpaid:
		return Standing{Reason: "unpaid"}
	case subscription.Status == stripe.SubscriptionStatusPaused:
		return Standing{Reason: "paused"}
	}

	if c.Plan(subscription).Paid() || hasPaymentMethod || subscription.DefaultPaymentMethod != nil {
		return Standing{Good: true}
	}

	if end := trialEnd(subscription); !end.IsZero() && !now.Before(end) {
		return Standing{Reason: "trial ended"}
	}

	return Standing{Good: true}
}

func trialEnd(subscription *stripe.Subscription) time.Time {
	interval := subscription.PendingInvoiceItemInterval

	if interval == nil || interval.Interval != stripe.SubscriptionPendingInvoiceItemIntervalIntervalDay {
		return time.Time{}
	}

	return time.Unix(subscription.StartDate, 0).AddDate(0, 0, int(interval.IntervalCount))
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
