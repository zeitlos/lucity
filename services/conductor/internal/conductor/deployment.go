package conductor

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func (c *Client) Deployments(ctx context.Context, serviceID ServiceID) ([]Deployment, error) {
	return c.platform.Deployments(ctx, serviceID)
}

func (c *Client) Deployment(ctx context.Context, id DeploymentID) (*Deployment, error) {
	return c.platform.Deployment(ctx, id)
}

func (c *Client) DefaultCommand(ctx context.Context, imageRef string) (string, error) {
	cfg, err := c.registry.ImageConfig(ctx, imageRef)

	if err != nil {
		return "", err
	}

	args := make([]string, 0, len(cfg.Entrypoint)+len(cfg.Cmd))
	args = append(args, cfg.Entrypoint...)
	args = append(args, cfg.Cmd...)

	return strings.Join(args, " "), nil
}

func (c *Client) imageExposedPort(ctx context.Context, imageRef string) int {
	cfg, err := c.registry.ImageConfig(ctx, imageRef)

	if err != nil {
		slog.Warn("could not inspect image for exposed port", "image", imageRef, "error", err)
		return 0
	}

	return exposedPort(cfg)
}

func exposedPort(cfg *v1.Config) int {
	best := 0

	for spec := range cfg.ExposedPorts {
		proto := "tcp"

		if i := strings.IndexByte(spec, '/'); i >= 0 {
			proto = strings.ToLower(spec[i+1:])
			spec = spec[:i]
		}

		if proto != "tcp" {
			continue
		}

		port, err := strconv.Atoi(spec)

		if err != nil || port < 1 || port > 65535 {
			continue
		}

		if best == 0 || port < best {
			best = port
		}
	}

	return best
}
