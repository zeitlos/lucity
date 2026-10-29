// Package customdomains attaches custom domains as soon as their DNS
// verifies, keeps every domain's attachment and edge registration in sync, and
// releases the edge zones nothing routes to any more.
package customdomains

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/domains"
	"github.com/zeitlos/lucity/services/stagehand/workspaces"
)

type Worker struct {
	domains         *domains.Client
	directory       *workspaces.Directory
	pendingInterval time.Duration
	syncInterval    time.Duration
}

// New returns a worker that checks domains waiting for DNS every
// pendingInterval, and re-verifies and syncs all domains every syncInterval.
func New(domains *domains.Client, directory *workspaces.Directory, pendingInterval, syncInterval time.Duration) *Worker {
	return &Worker{
		domains:         domains,
		directory:       directory,
		pendingInterval: pendingInterval,
		syncInterval:    syncInterval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	var loops sync.WaitGroup

	loops.Go(func() { every(ctx, w.pendingInterval, w.attachPending) })
	loops.Go(func() { every(ctx, w.syncInterval, w.sync) })

	loops.Wait()
}

func (w *Worker) attachPending(ctx context.Context) {
	known, err := w.directory.Known(ctx)

	if err != nil {
		slog.Error("pending domain check skipped", "error", err)
		return
	}

	if err := w.domains.AttachPending(ctx, known); err != nil && ctx.Err() == nil {
		slog.Error("pending domain check failed", "error", err)
	}
}

func (w *Worker) sync(ctx context.Context) {
	if err := w.domains.ReleaseZones(ctx); err != nil && ctx.Err() == nil {
		slog.Error("zone release failed", "error", err)
	}

	environments, err := w.directory.Environments(ctx)

	if err != nil {
		slog.Error("domain sync skipped", "error", err)
		return
	}

	if err := w.domains.Reconcile(ctx, workspaces.IDs(environments)); err != nil && ctx.Err() == nil {
		slog.Error("domain sync failed", "error", err)
	}
}

func every(ctx context.Context, interval time.Duration, run func(context.Context)) {
	for {
		run(ctx)

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
