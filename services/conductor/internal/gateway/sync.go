package gateway

import (
	"context"
	"log/slog"
)

// Sync brings the shared Gateway's listener set into agreement with the
// desired hostname list: every hostname gets an HTTP+HTTPS listener pair,
// and any custom-* listener whose hostname isn't in the input gets removed.
// The HTTPS listener references a Secret named `<resource>-tls`;
// cert-manager observes the Gateway annotation and creates the Certificate
// and Secret asynchronously. The HTTP listener exists so cert-manager's
// HTTP-01 solver HTTPRoute can attach.
//
// Idempotent — safe to call from a reconcile loop.
func (c *Client) Sync(ctx context.Context, hostnames []string) error {
	desired := make(map[string]struct{}, len(hostnames))

	for _, host := range hostnames {
		desired[host] = struct{}{}
	}

	listeners, err := c.listListeners(ctx)

	if err != nil {
		return err
	}

	// Add missing listeners for desired hostnames.
	for host := range desired {
		state := listeners[host]

		if !state.http {
			if err := c.addListener(ctx, host, "HTTP", ""); err != nil {
				slog.Warn("gateway sync: add http listener failed", "host", host, "error", err)
				continue
			}

			slog.Info("gateway sync: added http listener", "host", host)
		}

		if !state.https {
			secretName := ResourceNameFor(host) + "-tls"

			if err := c.addListener(ctx, host, "HTTPS", secretName); err != nil {
				slog.Warn("gateway sync: add https listener failed", "host", host, "error", err)
				continue
			}

			slog.Info("gateway sync: added https listener", "host", host)
		}
	}

	// Remove orphan listeners for hostnames no longer desired.
	for host, state := range listeners {
		if _, ok := desired[host]; ok {
			continue
		}

		name := ResourceNameFor(host)

		if state.http {
			if err := c.removeListener(ctx, name+"-http"); err != nil {
				slog.Warn("gateway sync: remove orphan http listener failed", "host", host, "error", err)
			} else {
				slog.Info("gateway sync: removed orphan http listener", "host", host)
			}
		}

		if state.https {
			if err := c.removeListener(ctx, name+"-https"); err != nil {
				slog.Warn("gateway sync: remove orphan https listener failed", "host", host, "error", err)
			} else {
				slog.Info("gateway sync: removed orphan https listener", "host", host)
			}
		}
	}

	return nil
}
