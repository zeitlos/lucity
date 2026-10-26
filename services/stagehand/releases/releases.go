// Package releases converges every tenant environment's namespace
// scaffolding and Helm release: the embedded chart version, the placement of
// the workspace's plan, and whether billing standing or an abuse flag
// suspends the workspace.
package releases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/environment"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/services/stagehand/workspaces"
)

const (
	customerKey  = "stripeCustomerId"
	suspendedKey = "suspended"
	abuseKey     = "abuse"

	maxNewSuspensions = 50
)

type Worker struct {
	environment  environment.Interface
	environments deployer.EnvironmentClient
	directory    *workspaces.Directory
	billing      *billing.Client
	defaults     values.Placement
	paid         values.Placement
	interval     time.Duration
	dryRun       bool
	enforce      bool

	shadow map[string]string
}

type verdict struct {
	paid      bool
	judged    bool
	suspended bool
	reason    string
}

// New returns a worker that converges environments every interval. With a
// nil billing client every environment gets the default placement and only
// the abuse flag suspends. Unless enforce is set, suspension is only logged.
func New(environment environment.Interface, environments deployer.EnvironmentClient, directory *workspaces.Directory, billing *billing.Client, defaults, paid values.Placement, interval time.Duration, dryRun, enforce bool) *Worker {
	return &Worker{
		environment:  environment,
		environments: environments,
		directory:    directory,
		billing:      billing,
		defaults:     defaults,
		paid:         paid,
		interval:     interval,
		dryRun:       dryRun,
		enforce:      enforce,
	}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		if err := w.reconcile(ctx); err != nil && ctx.Err() == nil {
			slog.Error("release pass skipped", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(w.interval):
		}
	}
}

func (w *Worker) reconcile(ctx context.Context) error {
	organizations, err := w.directory.Organizations(ctx)

	if err != nil {
		return err
	}

	verdicts, placementFrozen := w.verdicts(ctx, organizations)
	suspend, resume := changes(organizations, verdicts)

	if !w.enforce {
		w.logShadow(organizations, verdicts, suspend, resume)
	} else {
		w.mark(ctx, suspend, verdicts, true)
	}

	environments, err := w.directory.Environments(ctx)

	if err != nil {
		return err
	}

	desired := func(id platform.EnvironmentID, current deployer.EnvironmentState) deployer.EnvironmentState {
		verdict := verdicts[id.Workspace]
		state := current

		if !placementFrozen {
			state.Placement = w.defaults

			if verdict.paid {
				state.Placement = w.paid
			}
		}

		if w.enforce && verdict.judged {
			state.Suspended = verdict.suspended
		}

		return state
	}

	applied, planned := 0, 0
	done := map[platform.EnvironmentID]bool{}

	for _, environment := range environments {
		if err := ctx.Err(); err != nil {
			return err
		}

		verdict := verdicts[environment.ID.Workspace]

		if !w.enforce || !verdict.judged {
			continue
		}

		current, err := w.environments.State(ctx, environment.ID)

		if err != nil || current == nil || current.Suspended == verdict.suspended {
			continue
		}

		outdated, err := w.environments.Outdated(ctx, environment.ID)

		if err != nil {
			continue
		}

		done[environment.ID] = true

		if w.apply(ctx, environment.ID, verdict, *current, desired(environment.ID, *current), outdated) {
			applied++
		} else if w.dryRun {
			planned++
		}
	}

	for _, environment := range environments {
		if err := ctx.Err(); err != nil {
			return err
		}

		id := environment.ID

		if err := w.environment.Ensure(ctx, id, environment.ResourceTier); err != nil {
			slog.Warn("release: failed to ensure scaffolding", "environment", id, "error", err)
		}

		if done[id] {
			continue
		}

		current, err := w.environments.State(ctx, id)

		if err != nil {
			slog.Warn("release: failed to read release", "environment", id, "error", err)
			continue
		}

		if current == nil {
			continue
		}

		outdated, err := w.environments.Outdated(ctx, id)

		if err != nil {
			slog.Warn("release: failed to read chart version", "environment", id, "error", err)
			continue
		}

		state := desired(id, *current)

		if !outdated && state.Placement.Equal(current.Placement) && state.Suspended == current.Suspended {
			continue
		}

		if w.apply(ctx, id, verdicts[id.Workspace], *current, state, outdated) {
			applied++
		} else if w.dryRun {
			planned++
		}
	}

	if w.enforce {
		w.mark(ctx, resume, verdicts, false)
	}

	if applied > 0 || planned > 0 || len(suspend) > 0 || len(resume) > 0 {
		slog.Info("release pass done", "applied", applied, "planned", planned, "suspended", len(suspend), "resumed", len(resume), "enforced", w.enforce)
	}

	return nil
}

func (w *Worker) apply(ctx context.Context, id platform.EnvironmentID, verdict verdict, current, desired deployer.EnvironmentState, outdated bool) bool {
	log := slog.With("environment", id, "paid", verdict.paid, "outdated", outdated, "move", !current.Placement.Equal(desired.Placement), "suspended", desired.Suspended)

	if w.dryRun {
		log.Info("release: would apply")
		return false
	}

	revision, err := w.environments.SetState(ctx, id, desired)

	if err != nil {
		log.Warn("release: failed to apply", "error", err)
		return false
	}

	log.Info("release: applied", "revision", revision)

	return true
}

func (w *Worker) verdicts(ctx context.Context, organizations []logto.Organization) (map[string]verdict, bool) {
	verdicts := make(map[string]verdict, len(organizations))

	for _, organization := range organizations {
		if abuse, _ := organization.CustomData[abuseKey].(bool); abuse {
			verdicts[organization.Name] = verdict{judged: true, suspended: true, reason: "abuse"}
		}
	}

	if w.billing == nil {
		return verdicts, false
	}

	subscriptions, err := w.billing.Subscriptions(ctx)

	if err != nil {
		slog.Error("placement and suspension frozen for this pass", "error", fmt.Errorf("read subscriptions: %w", err))
		return verdicts, true
	}

	paid := 0

	for _, organization := range organizations {
		customer, _ := organization.CustomData[customerKey].(string)

		if subscription := subscriptions[customer]; customer != "" && subscription != nil && w.billing.Paid(subscription) {
			verdict := verdicts[organization.Name]
			verdict.paid = true
			verdicts[organization.Name] = verdict
			paid++
		}
	}

	if paid == 0 {
		slog.Error("placement and suspension frozen for this pass", "error", errors.New("no workspace has a paid plan, check the plan price IDs"))
		return abuseOnly(verdicts), true
	}

	methods, err := w.billing.PaymentMethods(ctx)

	if err != nil {
		slog.Error("suspension frozen for this pass", "error", fmt.Errorf("read payment methods: %w", err))
		return verdicts, false
	}

	now := time.Now()
	newlySuspended := 0

	for _, organization := range organizations {
		verdict := verdicts[organization.Name]

		if verdict.reason == "abuse" {
			continue
		}

		customer, _ := organization.CustomData[customerKey].(string)
		subscription := subscriptions[customer]

		switch {
		case customer == "":
			verdict.reason = "no customer"
		case subscription == nil:
			verdict.reason = "no subscription"
		default:
			standing := w.billing.Standing(subscription, methods[customer], now)
			verdict.judged = true
			verdict.suspended = !standing.Good
			verdict.reason = standing.Reason

			if flagged, _ := organization.CustomData[suspendedKey].(bool); verdict.suspended && !flagged {
				newlySuspended++
			}
		}

		verdicts[organization.Name] = verdict
	}

	if newlySuspended > maxNewSuspensions {
		slog.Error("suspension frozen for this pass", "error", fmt.Errorf("%d workspaces would be newly suspended, more than %d", newlySuspended, maxNewSuspensions))
		return abuseOnly(verdicts), false
	}

	return verdicts, false
}

func abuseOnly(verdicts map[string]verdict) map[string]verdict {
	for name, current := range verdicts {
		if current.reason != "abuse" {
			verdicts[name] = verdict{paid: current.paid}
		}
	}

	return verdicts
}

func changes(organizations []logto.Organization, verdicts map[string]verdict) (suspend, resume []logto.Organization) {
	for _, organization := range organizations {
		verdict := verdicts[organization.Name]

		if !verdict.judged {
			continue
		}

		flagged, _ := organization.CustomData[suspendedKey].(bool)

		switch {
		case verdict.suspended && !flagged:
			suspend = append(suspend, organization)
		case !verdict.suspended && flagged:
			resume = append(resume, organization)
		}
	}

	return suspend, resume
}

func (w *Worker) mark(ctx context.Context, organizations []logto.Organization, verdicts map[string]verdict, suspended bool) {
	for _, organization := range organizations {
		if w.dryRun {
			slog.Info("suspension: would mark", "workspace", organization.Name, "suspended", suspended, "reason", verdicts[organization.Name].reason)
			continue
		}

		if err := w.directory.MarkSuspended(ctx, organization, suspended); err != nil {
			slog.Warn("suspension: failed to mark workspace", "workspace", organization.Name, "suspended", suspended, "error", err)
			continue
		}

		slog.Info("suspension: workspace marked", "workspace", organization.Name, "suspended", suspended, "reason", verdicts[organization.Name].reason)
	}
}

func (w *Worker) logShadow(organizations []logto.Organization, verdicts map[string]verdict, suspend, resume []logto.Organization) {
	lines := map[string]string{}

	for _, organization := range suspend {
		lines[organization.Name] = "would suspend: " + verdicts[organization.Name].reason
	}

	for _, organization := range resume {
		lines[organization.Name] = "would resume"
	}

	unjudged := 0

	for _, organization := range organizations {
		if verdict := verdicts[organization.Name]; !verdict.judged && verdict.reason != "" {
			lines[organization.Name] = "not judged: " + verdict.reason
			unjudged++
		}
	}

	if !maps.Equal(lines, w.shadow) {
		for _, name := range slices.Sorted(maps.Keys(lines)) {
			slog.Info("suspension shadow", "workspace", name, "verdict", lines[name])
		}

		w.shadow = lines
	}

	slog.Info("suspension shadow pass", "would_suspend", len(suspend), "would_resume", len(resume), "not_judged", unjudged)
}
