// Package releases converges every tenant environment's namespace
// scaffolding and Helm release: the embedded chart version and the
// placement of the workspace's plan.
package releases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/environment"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/services/stagehand/workspaces"
)

const customerKey = "stripeCustomerId"

type Worker struct {
	environment  environment.Interface
	environments deployer.EnvironmentClient
	directory    *workspaces.Directory
	billing      *billing.Client
	defaults     values.Placement
	paid         values.Placement
	interval     time.Duration
	dryRun       bool
}

// New returns a worker that converges environments every interval. With a
// nil billing client every environment gets the default placement.
func New(environment environment.Interface, environments deployer.EnvironmentClient, directory *workspaces.Directory, billing *billing.Client, defaults, paid values.Placement, interval time.Duration, dryRun bool) *Worker {
	return &Worker{
		environment:  environment,
		environments: environments,
		directory:    directory,
		billing:      billing,
		defaults:     defaults,
		paid:         paid,
		interval:     interval,
		dryRun:       dryRun,
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

	paid, err := w.paidWorkspaces(ctx, organizations)
	frozen := err != nil

	if frozen {
		slog.Error("placement frozen for this pass", "error", err)
	}

	environments, err := w.directory.Environments(ctx)

	if err != nil {
		return err
	}

	applied, planned := 0, 0

	for _, environment := range environments {
		if err := ctx.Err(); err != nil {
			return err
		}

		id := environment.ID
		workspacePaid := paid[id.Workspace]

		if err := w.environment.Ensure(ctx, id, environment.ResourceTier); err != nil {
			slog.Warn("release: failed to ensure scaffolding", "environment", id, "error", err)
		}

		current, err := w.environments.Placement(ctx, id)

		if err != nil {
			slog.Warn("release: failed to read release", "environment", id, "error", err)
			continue
		}

		if current == nil {
			continue
		}

		desired := w.defaults

		if workspacePaid {
			desired = w.paid
		}

		if frozen {
			desired = *current
		}

		outdated, err := w.environments.Outdated(ctx, id)

		if err != nil {
			slog.Warn("release: failed to read chart version", "environment", id, "error", err)
			continue
		}

		move := !current.Equal(desired)

		if !outdated && !move {
			continue
		}

		log := slog.With("environment", id, "paid", workspacePaid, "outdated", outdated, "move", move)

		if w.dryRun {
			log.Info("release: would apply")
			planned++
			continue
		}

		revision, err := w.environments.Place(ctx, id, desired)

		if err != nil {
			log.Warn("release: failed to apply", "error", err)
			continue
		}

		log.Info("release: applied", "revision", revision)
		applied++
	}

	if applied > 0 || planned > 0 {
		slog.Info("release pass done", "applied", applied, "planned", planned)
	}

	return nil
}

func (w *Worker) paidWorkspaces(ctx context.Context, organizations []logto.Organization) (map[string]bool, error) {
	paid := map[string]bool{}

	if w.billing == nil {
		return paid, nil
	}

	plans, err := w.billing.Plans(ctx)

	if err != nil {
		return nil, fmt.Errorf("read plans: %w", err)
	}

	for _, organization := range organizations {
		customer, _ := organization.CustomData[customerKey].(string)

		if customer != "" && plans[customer].Paid() {
			paid[organization.Name] = true
		}
	}

	if len(paid) == 0 {
		return nil, errors.New("no workspace has a paid plan, check the plan price IDs")
	}

	return paid, nil
}
