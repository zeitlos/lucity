// Package customdomains attaches custom domains as soon as their DNS
// verifies, keeps every domain's attachment and edge registration in sync, and
// releases the edge zones nothing routes to any more.
package customdomains

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/domains"
	"github.com/zeitlos/lucity/pkg/metrics"
	"github.com/zeitlos/lucity/services/stagehand/workspaces"
)

type Worker struct {
	domains         *domains.Client
	directory       *workspaces.Directory
	loops           *metrics.Loops
	pendingInterval time.Duration
	syncInterval    time.Duration
}

// New returns a worker that checks domains waiting for DNS every
// pendingInterval, and re-verifies and syncs all domains every syncInterval.
func New(domains *domains.Client, directory *workspaces.Directory, loops *metrics.Loops, pendingInterval, syncInterval time.Duration) *Worker {
	return &Worker{
		domains:         domains,
		directory:       directory,
		loops:           loops,
		pendingInterval: pendingInterval,
		syncInterval:    syncInterval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	var loops sync.WaitGroup

	loops.Go(func() { w.every(ctx, "domains_pending", w.pendingInterval, w.attachPending) })
	loops.Go(func() { w.every(ctx, "domains_sync", w.syncInterval, w.sync) })

	loops.Wait()
}

func (w *Worker) attachPending(ctx context.Context) error {
	known, err := w.directory.Known(ctx)

	if err != nil {
		slog.Error("pending domain check skipped", "error", err)
		return err
	}

	err = w.domains.AttachPending(ctx, known)

	if err != nil && ctx.Err() == nil {
		slog.Error("pending domain check failed", "error", err)
	}

	return err
}

func (w *Worker) sync(ctx context.Context) error {
	zonesErr := w.domains.ReleaseZones(ctx)

	if zonesErr != nil && ctx.Err() == nil {
		slog.Error("zone release failed", "error", zonesErr)
	}

	environments, err := w.directory.Environments(ctx)

	if err != nil {
		slog.Error("domain sync skipped", "error", err)
		return err
	}

	err = w.domains.Reconcile(ctx, workspaces.IDs(environments))

	if err != nil && ctx.Err() == nil {
		slog.Error("domain sync failed", "error", err)
	}

	return errors.Join(zonesErr, err)
}

func (w *Worker) every(ctx context.Context, loop string, interval time.Duration, run func(context.Context) error) {
	for {
		start := time.Now()
		w.loops.Observe(loop, start, run(ctx))

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
