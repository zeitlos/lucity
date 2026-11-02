package cockpit

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/stripe/stripe-go/v87"
	"golang.org/x/sync/errgroup"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/logto"
)

const (
	cardGateAccountAge = 90 * 24 * time.Hour
	billingConcurrency = 4
	rateLimitWait      = 8 * time.Second
)

func (c *Client) cardGates(ctx context.Context, organizations []logto.Organization, subscriptions map[string]*stripe.Subscription, members map[string][]logto.OrganizationMember) (map[string]bool, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(billingConcurrency)

	gated := make([]bool, len(organizations))
	lifted := make([]bool, len(organizations))

	for i, organization := range organizations {
		personal, _ := organization.CustomData[personalKey].(bool)
		abuse, _ := organization.CustomData[abuseKey].(bool)
		suspended, _ := organization.CustomData[suspendedKey].(bool)
		customer, _ := organization.CustomData[customerKey].(string)

		if !personal || abuse || suspended || c.paid(subscriptions[customer]) || !youngAccount(members[organization.ID]) {
			continue
		}

		gated[i] = true

		if customer == "" {
			continue
		}

		group.Go(func() (err error) {
			lifted[i], err = c.usableCard(groupCtx, customer)
			return err
		})
	}

	if err := group.Wait(); err != nil {
		return nil, fmt.Errorf("check cards: %w", err)
	}

	result := map[string]bool{}

	for i, organization := range organizations {
		if gated[i] {
			result[organization.ID] = lifted[i]
		}
	}

	return result, nil
}

func (c *Client) usableCard(ctx context.Context, customer string) (bool, error) {
	for wait := time.Second; ; wait *= 2 {
		usable, err := c.billing.UsableCard(ctx, customer)

		var stripeErr *stripe.Error

		if !errors.As(err, &stripeErr) || stripeErr.Code != stripe.ErrorCodeRateLimit || wait > rateLimitWait {
			return usable, err
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(wait + rand.N(wait)):
		}
	}
}

func (c *Client) paid(subscription *stripe.Subscription) bool {
	return subscription != nil && !billing.Ended(subscription) && c.billing.Plan(subscription).Paid()
}

func youngAccount(members []logto.OrganizationMember) bool {
	if len(members) == 0 {
		return false
	}

	owner := members[0]

	for _, identity := range owner.Identities {
		created, ok := identity.CreatedAt()

		if !ok || owner.CreatedAt == 0 {
			continue
		}

		return time.UnixMilli(owner.CreatedAt).Sub(created) < cardGateAccountAge
	}

	return false
}
