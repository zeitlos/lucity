package conductor

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
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

func (c *Client) CommitMessage(ctx context.Context, installationID, repo, hash string) (string, error) {
	if !isCommitSHA(hash) {
		return "", nil
	}

	id, err := strconv.Atoi(installationID)

	if err != nil {
		return "", fmt.Errorf("inalid installation id: %v", err)
	}

	repoURL, err := url.Parse(repo)

	if err != nil {
		return "", err
	}

	ownerRepo := repoURL.Path
	ownerRepo = strings.TrimPrefix(ownerRepo, "/")
	ownerRepo = strings.TrimSuffix(ownerRepo, path.Ext(ownerRepo))

	return c.GitHubApp.CommitMessage(ctx, int64(id), ownerRepo, hash)
}

func isCommitSHA(s string) bool {
	if len(s) < 7 || len(s) > 64 {
		return false
	}

	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}

	return true
}

func (c *Client) DefaultCommand(ctx context.Context, imageRef string) (string, error) {
	ref, err := name.ParseReference(imageRef)

	if err != nil {
		return "", err
	}

	registry := ref.Context().RegistryStr()
	authConfig := authn.Anonymous

	if strings.EqualFold(registry, c.Config.RegistryPullURL) {
		registry = c.Config.RegistryPushURL

		pullRegistry, err := name.NewRegistry(c.Config.RegistryPullURL)

		if err != nil {
			return "", err
		}

		authConfig, err = c.Config.RegistryPullSecret.Resolve(pullRegistry)

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
