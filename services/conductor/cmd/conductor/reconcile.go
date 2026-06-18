package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
)

const domainReconcileInterval = 2 * time.Minute
const serviceReconcileInterval = 2 * time.Minute

func runDomainReconciler(ctx context.Context, c *conductor.Client) {
	reconcile(ctx, domainReconcileInterval, func() {
		if err := c.ReconcileDomains(ctx); err != nil {
			slog.Error("domain reconcile failed", "error", err)
		}
	})
}

func runServiceReconciler(ctx context.Context, c *conductor.Client) {
	reconcile(ctx, serviceReconcileInterval, func() {
		if err := c.ReconcileServices(ctx); err != nil {
			slog.Error("service reconcile failed", "error", err)
		}
	})
}

func reconcile(ctx context.Context, interval time.Duration, fun func()) {
	fun()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fun()
		}
	}

}
