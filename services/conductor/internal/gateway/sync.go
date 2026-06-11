package gateway

import (
	"context"
	"log/slog"
)

func (c *Client) Sync(ctx context.Context, hostnames []string) error {
	desired := make(map[string]struct{}, len(hostnames))

	for _, host := range hostnames {
		desired[host] = struct{}{}
	}

	listeners, err := c.listListeners(ctx)

	if err != nil {
		return err
	}

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

	for host, state := range listeners {
		if _, ok := desired[host]; ok {
			continue
		}

		name := ResourceNameFor(host)

		if state.http {
			slog.WarnContext(ctx, "dry run: would remove orphan https listener", "host", host, "name", name+"-http")
			// if err := c.removeListener(ctx, name+"-http"); err != nil {
			// 	slog.Warn("gateway sync: remove orphan http listener failed", "host", host, "error", err)
			// } else {
			// 	slog.Info("gateway sync: removed orphan http listener", "host", host)
			// }
		}

		if state.https {
			slog.WarnContext(ctx, "dry run: would remove orphan https listener", "host", host, "name", name+"-https")
			// if err := c.removeListener(ctx, name+"-https"); err != nil {
			// 	slog.Warn("gateway sync: remove orphan https listener failed", "host", host, "error", err)
			// } else {
			// 	slog.Info("gateway sync: removed orphan https listener", "host", host)
			// }
		}
	}

	return nil
}
