// Package admission starts queued build, scan and deploy Jobs as soon as the
// concurrency limits allow.
package admission

import (
	"context"
	"log/slog"
	"time"

	"github.com/zeitlos/lucity/pkg/jobs"
	"github.com/zeitlos/lucity/pkg/pipeline"
)

const fallbackInterval = 30 * time.Second

type Worker struct {
	pipeline pipeline.Interface
	trigger  chan struct{}
}

// New returns a worker that runs admission whenever a Job or a builder pod
// changes, and at least every 30 seconds.
func New(cache *jobs.Cache, builders *jobs.Builders, pipeline pipeline.Interface) (*Worker, error) {
	w := &Worker{
		pipeline: pipeline,
		trigger:  make(chan struct{}, 1),
	}

	if err := cache.OnChange(w.notify); err != nil {
		return nil, err
	}

	if err := builders.OnChange(w.notify); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Worker) notify() {
	select {
	case w.trigger <- struct{}{}:
	default:
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(fallbackInterval)
	defer ticker.Stop()

	for {
		if err := w.pipeline.Reconcile(ctx); err != nil && ctx.Err() == nil {
			slog.Error("release admission failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-w.trigger:
		case <-ticker.C:
		}
	}
}
