package gateway

import (
	"context"
	"log/slog"
)

func (c *Client) Sync(ctx context.Context, hostnames []string, removeOrphans bool) error {
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

		if err := c.ensureCertificate(ctx, host); err != nil {
			slog.Warn("gateway sync: ensure certificate failed", "host", host, "error", err)
			continue
		}

		secretName := ResourceNameFor(host) + "-tls"

		if state.https && state.httpsSecret != secretName {
			if err := c.removeListener(ctx, host, "HTTPS"); err != nil {
				slog.Warn("gateway sync: remove stale https listener failed", "host", host, "error", err)
				continue
			}

			state.https = false

			slog.Info("gateway sync: removed https listener with stale secret ref", "host", host)
		}

		if !state.https {
			exists, err := c.secretExists(ctx, secretName)

			if err != nil {
				slog.Warn("gateway sync: check tls secret failed", "host", host, "error", err)
				continue
			}

			if !exists {
				slog.Info("gateway sync: waiting for tls secret before adding https listener", "host", host, "secret", secretName)
				continue
			}

			if err := c.addListener(ctx, host, "HTTPS", secretName); err != nil {
				slog.Warn("gateway sync: add https listener failed", "host", host, "error", err)
				continue
			}

			slog.Info("gateway sync: added https listener", "host", host)
		}
	}

	if !removeOrphans {
		return nil
	}

	for host, state := range listeners {
		if _, ok := desired[host]; ok {
			continue
		}

		if state.http {
			if err := c.removeListener(ctx, host, "HTTP"); err != nil {
				slog.Warn("gateway sync: remove orphan http listener failed", "host", host, "error", err)
			} else {
				slog.Info("gateway sync: removed orphan http listener", "host", host)
			}
		}

		if state.https {
			if err := c.removeListener(ctx, host, "HTTPS"); err != nil {
				slog.Warn("gateway sync: remove orphan https listener failed", "host", host, "error", err)
			} else {
				slog.Info("gateway sync: removed orphan https listener", "host", host)
			}
		}

		if err := c.removeCertificate(ctx, host); err != nil {
			slog.Warn("gateway sync: remove orphan certificate failed", "host", host, "error", err)
		}
	}

	return nil
}
