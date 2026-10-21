// Package placement keeps every environment's workloads on the node pool that
// matches its workspace's plan.
package placement

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
	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/platform"
)

const customerKey = "stripeCustomerId"

type Worker struct {
	k8s          kubernetes.Interface
	environments deployer.EnvironmentClient
	directory    *logto.Client
	billing      *billing.Client
	defaults     values.Placement
	paid         values.Placement
	interval     time.Duration
	maxMoves     int
	dryRun       bool
}

// New returns a worker that moves environments between placements. With a nil
// billing client every environment gets the default placement.
func New(k8s kubernetes.Interface, environments deployer.EnvironmentClient, directory *logto.Client, billing *billing.Client, defaults, paid values.Placement, interval time.Duration, maxMoves int, dryRun bool) *Worker {
	return &Worker{
		k8s:          k8s,
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
			slog.Error("placement pass skipped", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(w.interval):
		}
	}
}

type environment struct {
	id   platform.EnvironmentID
	paid bool
}

func (w *Worker) reconcile(ctx context.Context) error {
	paid, err := w.paidWorkspaces(ctx)

	if err != nil {
		return err
	}

	environments, err := w.environmentList(ctx, paid)

	if err != nil {
		return err
	}

	moved, deferred, planned := 0, 0, 0

	for _, env := range environments {
		if err := ctx.Err(); err != nil {
			return err
		}

		desired := w.defaults

		if env.paid {
			desired = w.paid
		}

		current, err := w.environments.Placement(ctx, env.id)

		if err != nil {
			slog.Warn("placement: failed to read release", "environment", env.id, "error", err)
			continue
		}

		if current == nil || current.Equal(desired) {
			continue
		}

		running, err := w.hasRunningPods(ctx, env.id.Namespace())

		if err != nil {
			slog.Warn("placement: failed to list pods", "environment", env.id, "error", err)
			continue
		}

		if running && moved >= w.maxMoves {
			deferred++
			continue
		}

		log := slog.With("environment", env.id, "paid", env.paid, "running", running)

		if w.dryRun {
			log.Info("placement: would move environment")
			planned++
			continue
		}

		revision, err := w.environments.Place(ctx, env.id, desired)

		if err != nil {
			log.Warn("placement: failed to move environment", "error", err)
			continue
		}

		log.Info("placement: moved environment", "revision", revision)

		if running {
			moved++
		}
	}

	if moved > 0 || deferred > 0 || planned > 0 {
		slog.Info("placement pass done", "moved", moved, "deferred", deferred, "planned", planned)
	}

	return nil
}

func (w *Worker) paidWorkspaces(ctx context.Context) (map[string]bool, error) {
	paid := map[string]bool{}

	if w.billing == nil {
		return paid, nil
	}

	organizations, err := w.directory.Organizations(ctx)

	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
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

// environmentList returns every tenant environment, those of paid workspaces
// first, so they move before the free pool's backlog.
func (w *Worker) environmentList(ctx context.Context, paid map[string]bool) ([]environment, error) {
	namespaces, err := w.k8s.CoreV1().Namespaces().List(ctx, meta.ListOptions{
		LabelSelector: labels.Workspace + "," + labels.Project + "," + labels.Environment,
	})

	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	environments := make([]environment, 0, len(namespaces.Items))

	for _, namespace := range namespaces.Items {
		id := platform.EnvironmentID{
			Workspace: namespace.Labels[labels.Workspace],
			Project:   namespace.Labels[labels.Project],
			Name:      namespace.Labels[labels.Environment],
		}

		if id.Namespace() != namespace.Name {
			slog.Warn("placement: namespace labels don't match its name", "namespace", namespace.Name, "environment", id)
			continue
		}

		environments = append(environments, environment{id: id, paid: paid[id.Workspace]})
	}

	sort.SliceStable(environments, func(i, j int) bool {
		return environments[i].paid && !environments[j].paid
	})

	return environments, nil
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
