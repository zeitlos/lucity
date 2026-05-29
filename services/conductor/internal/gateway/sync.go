package gateway

import (
	"context"
	"log/slog"
)

// Sync ensures the shared Gateway has an HTTP+HTTPS listener pair for every
// hostname in the input. The HTTPS listener references a Secret named
// `<resource>-tls`; cert-manager observes the Gateway annotation and creates
// the Certificate + Secret asynchronously. The HTTP listener exists so
// cert-manager's HTTP-01 solver HTTPRoute can attach.
//
// COEXISTENCE: additive only. Orphan listeners are NOT pruned — the legacy
// `deployerold.ReconcileCustomDomains` goroutine still runs and owns
// deletion. Both reconcilers use the same `custom-*` naming, so post-cutover
// (step 6 of the refactor) this method gains a prune step and the legacy
// goroutine is removed.
//
// Idempotent.
func (c *Client) Sync(ctx context.Context, hostnames []string) error {
	listeners, err := c.listListeners(ctx)

	if err != nil {
		return err
	}

	for _, host := range hostnames {
		state := listeners[host]

		if !state.http {
			if err := c.addListener(ctx, host, "HTTP", ""); err != nil {
				slog.Warn("gateway sync: add http listener failed", "host", host, "error", err)
				continue
			}

			slog.Info("gateway sync: added http listener", "host", host)
		}

		if !state.https {
			secretName := resourceNameFor(host) + "-tls"

			if err := c.addListener(ctx, host, "HTTPS", secretName); err != nil {
				slog.Warn("gateway sync: add https listener failed", "host", host, "error", err)
				continue
			}

			slog.Info("gateway sync: added https listener", "host", host)
		}
	}

	return nil
}
