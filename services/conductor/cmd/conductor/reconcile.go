package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
)

const domainReconcileInterval = 2 * time.Minute

func runDomainReconciler(ctx context.Context, c *conductor.Client) {
	ticker := time.NewTicker(domainReconcileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.ReconcileDomains(ctx); err != nil {
				slog.Error("domain reconcile failed", "error", err)
			}
		}
	}
}
