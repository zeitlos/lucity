package cockpit

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/stripe/stripe-go/v87"
	"golang.org/x/sync/errgroup"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/services/cockpit/internal/signals"
)

const identityConcurrency = 16

func (c *Client) load(ctx context.Context) ([]Workspace, error) {
	var (
		organizations []logto.Organization
		subscriptions map[string]*stripe.Subscription
		methods       map[string]bool
		environments  map[string][]Environment
		egress        map[string]float64
	)

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() (err error) {
		organizations, err = c.identity.Organizations(groupCtx)
		return err
	})
	group.Go(func() (err error) {
		subscriptions, err = c.billing.Subscriptions(groupCtx)
		return err
	})
	group.Go(func() (err error) {
		methods, err = c.billing.PaymentMethods(groupCtx)
		return err
	})
	group.Go(func() (err error) {
		environments, err = c.cluster(groupCtx)
		return err
	})
	group.Go(func() (err error) {
		egress, err = c.egress(groupCtx)
		return err
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	members, err := c.members(ctx, organizations)

	if err != nil {
		return nil, err
	}

	var githubIDs []int64
	var repositories []string

	for _, list := range members {
		for _, member := range list {
			if id, ok := githubID(member.Identities["github"].UserID); ok && identityAccount(member.Identities["github"]) == nil {
				githubIDs = append(githubIDs, id)
			}
		}
	}

	for _, list := range environments {
		for _, environment := range list {
			for _, service := range environment.Services {
				if service.Repository != "" {
					repositories = append(repositories, service.Repository)
				}
			}
		}
	}

	accounts := c.accountsByID(ctx, unique(githubIDs))
	parents := c.parentsByURL(ctx, unique(repositories))

	now := time.Now()
	workspaces := make([]Workspace, 0, len(organizations))
	facts := make([]signals.Workspace, 0, len(organizations))

	for _, organization := range organizations {
		workspace := Workspace{
			ID:           organization.Name,
			CreatedAt:    time.UnixMilli(int64(organization.CreatedAt)),
			Environments: environments[organization.Name],
		}

		workspace.Abuse, _ = organization.CustomData[abuseKey].(bool)
		workspace.AbuseNote, _ = organization.CustomData[abuseNoteKey].(string)
		workspace.Suspended, _ = organization.CustomData[suspendedKey].(bool)

		if customer, _ := organization.CustomData[customerKey].(string); customer != "" {
			if subscription := subscriptions[customer]; subscription != nil {
				standing := c.billing.Standing(subscription, methods[customer], now)

				workspace.Billing = &Billing{
					Plan:             c.billing.Plan(subscription),
					GoodStanding:     standing.Good,
					StandingReason:   standing.Reason,
					HasPaymentMethod: methods[customer],
				}

				if workspace.Billing.Plan == billing.PlanNone {
					workspace.Billing.TrialEndsAt = billing.TrialEnd(subscription)
				}
			}
		}

		fact := signals.Workspace{ID: workspace.ID, CreatedAt: workspace.CreatedAt}

		for _, member := range members[organization.ID] {
			result := Member{Name: member.Name, Email: member.Email}

			if len(member.OrgRoles) > 0 {
				result.Role = member.OrgRoles[0].Name
			}

			if account := identityAccount(member.Identities["github"]); account != nil {
				result.Account = account
			} else if id, ok := githubID(member.Identities["github"].UserID); ok {
				result.Account = accounts[id]
			}

			if result.Account != nil {
				fact.Accounts = append(fact.Accounts, result.Account.CreatedAt)
			}

			workspace.Members = append(workspace.Members, result)
		}

		for _, environment := range workspace.Environments {
			workspace.RunningPods += environment.RunningPods
			workspace.PendingPods += environment.PendingPods
			workspace.MemoryRequested += environment.MemoryRequested
			workspace.Egress7d += int64(egress[environment.ID.Namespace()])

			for _, service := range environment.Services {
				fact.Commits = append(fact.Commits, service.Commit)
				fact.Images = append(fact.Images, service.Image)
				fact.ForkParents = append(fact.ForkParents, parents[service.Repository])
			}
		}

		workspaces = append(workspaces, workspace)
		facts = append(facts, fact)
	}

	evaluated := signals.Evaluate(facts)

	for i := range workspaces {
		workspaces[i].Signals = evaluated[workspaces[i].ID]
	}

	sort.Slice(workspaces, func(i, j int) bool {
		return workspaces[i].CreatedAt.After(workspaces[j].CreatedAt)
	})

	return workspaces, nil
}

func unique[T comparable](values []T) []T {
	seen := make(map[T]bool, len(values))
	var result []T

	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}

	return result
}

func (c *Client) members(ctx context.Context, organizations []logto.Organization) (map[string][]logto.OrganizationMember, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(identityConcurrency)

	lists := make([][]logto.OrganizationMember, len(organizations))

	for i, organization := range organizations {
		group.Go(func() (err error) {
			lists[i], err = c.identity.OrganizationMembers(groupCtx, organization.ID)
			return err
		})
	}

	if err := group.Wait(); err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	result := make(map[string][]logto.OrganizationMember, len(organizations))

	for i, organization := range organizations {
		result[organization.ID] = lists[i]
	}

	return result, nil
}
