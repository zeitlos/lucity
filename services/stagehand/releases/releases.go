// Package releases converges every tenant environment's namespace
// scaffolding and Helm release: the embedded chart version and the
// placement of the workspace's plan.
package releases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/environment"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/services/stagehand/workspaces"
)

const customerKey = "stripeCustomerId"

type Worker struct {
	k8s          kubernetes.Interface
	environment  environment.Interface
	environments deployer.EnvironmentClient
	directory    *workspaces.Directory
	billing      *billing.Client
	defaults     values.Placement
	paid         values.Placement
	interval     time.Duration
	maxMoves     int
	dryRun       bool
}

// New returns a worker that converges environments every interval. With a
// nil billing client every environment gets the default placement.
func New(k8s kubernetes.Interface, environment environment.Interface, environments deployer.EnvironmentClient, directory *workspaces.Directory, billing *billing.Client, defaults, paid values.Placement, interval time.Duration, maxMoves int, dryRun bool) *Worker {
	return &Worker{
		k8s:          k8s,
		environment:  environment,
		environments: environments,
		directory:    directory,
		billing:      billing,
		defaults:     defaults,
		paid:         paid,
		interval:     interval,
		maxMoves:     maxMoves,
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

type tenant struct {
	environment platform.Environment
	paid        bool
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

	tenants, err := w.tenants(ctx, paid)

	if err != nil {
		return err
	}

	applied, deferred, planned := 0, 0, 0

	for _, tenant := range tenants {
		if err := ctx.Err(); err != nil {
			return err
		}

		id := tenant.environment.ID

		if err := w.environment.Ensure(ctx, id, tenant.environment.ResourceTier); err != nil {
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

		if tenant.paid {
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

		running, err := w.hasRunningPods(ctx, id.Namespace())

		if err != nil {
			slog.Warn("release: failed to list pods", "environment", id, "error", err)
			continue
		}

		if running && applied >= w.maxMoves {
			deferred++
			continue
		}

		log := slog.With("environment", id, "paid", tenant.paid, "running", running, "outdated", outdated, "move", move)

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

		if running {
			applied++
		}
	}

	if applied > 0 || deferred > 0 || planned > 0 {
		slog.Info("release pass done", "applied", applied, "deferred", deferred, "planned", planned)
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

// tenants returns every environment, those of paid workspaces first, so they
// move before the free pool's backlog.
func (w *Worker) tenants(ctx context.Context, paid map[string]bool) ([]tenant, error) {
	environments, err := w.directory.Environments(ctx)

	if err != nil {
		return nil, err
	}

	tenants := make([]tenant, 0, len(environments))

	for _, environment := range environments {
		tenants = append(tenants, tenant{environment: environment, paid: paid[environment.ID.Workspace]})
	}

	sort.SliceStable(tenants, func(i, j int) bool {
		return tenants[i].paid && !tenants[j].paid
	})

	return tenants, nil
}

func (w *Worker) hasRunningPods(ctx context.Context, namespace string) (bool, error) {
	pods, err := w.k8s.CoreV1().Pods(namespace).List(ctx, meta.ListOptions{
		FieldSelector: "status.phase=" + string(core.PodRunning),
		Limit:         1,
	})

	if err != nil {
		return false, err
	}

	return len(pods.Items) > 0, nil
}
