package conductor

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func (c *Client) Deployments(ctx context.Context, serviceID ServiceID) ([]Deployment, error) {
	return c.platform.Deployments(ctx, serviceID)
}

func (c *Client) Deployment(ctx context.Context, id DeploymentID) (*Deployment, error) {
	return c.platform.Deployment(ctx, id)
}

func (c *Client) DefaultCommand(ctx context.Context, imageRef string) (string, error) {
	ref, err := name.ParseReference(imageRef)

	if err != nil {
		return "", err
	}

	registry := ref.Context().RegistryStr()
	authConfig := authn.Anonymous

	if strings.EqualFold(registry, c.config.RegistryPullURL) {
		registry = c.config.RegistryPushURL

		pullRegistry, err := name.NewRegistry(c.config.RegistryPullURL)

		if err != nil {
			return "", err
		}

		authConfig, err = c.config.RegistryPullSecret.Resolve(pullRegistry)

		if err != nil {
			return "", err
		}
	}

	repo, err := name.NewRepository(registry + "/" + ref.Context().RepositoryStr())

	if err != nil {
		return "", err
	}

	switch r := ref.(type) {
	case name.Tag:
		ref = repo.Tag(r.TagStr())
	case name.Digest:
		ref = repo.Digest(r.DigestStr())
	default:
		return "", fmt.Errorf("unexpected reference type %T", ref)
	}

	img, err := remote.Image(ref,
		remote.WithContext(ctx),
		remote.WithAuth(authConfig),
	)

	if err != nil {
		return "", err
	}

	cfg, err := img.ConfigFile()

	if err != nil {
		return "", err
	}

	args := append(cfg.Config.Entrypoint, cfg.Config.Cmd...)

	return strings.Join(args, " "), nil
}
