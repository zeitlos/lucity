package graphql

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

func (r *Resolver) requireServiceDeployBinding(ctx context.Context, serviceID platform.ServiceID) error {
	claims, err := auth.FromContext(ctx)

	if err != nil {
		return err
	}

	workspace, err := tenant.FromContext(ctx)

	if err != nil {
		return err
	}

	if claims.WorkspaceRoleIn(workspace) != auth.WorkspaceRoleDeployer {
		return nil
	}

	repository, ok := conductor.GitHubActionsRepo(claims.Subject)

	if !ok {
		return errors.New("not found")
	}

	service, err := r.Conductor.Service(ctx, serviceID)

	if err != nil {
		return err
	}

	if !service.CIDeploy || !sameGitHubRepo(service.SourceURL, repository) {
		return errors.New("not found")
	}

	return nil
}

func sameGitHubRepo(sourceURL, repository string) bool {
	want := strings.ToLower(strings.TrimSuffix(repository, ".git"))

	return want != "" && normalizeGitHubRepo(sourceURL) == want
}

func normalizeGitHubRepo(repoURL string) string {
	parsed, err := url.Parse(repoURL)

	if err != nil {
		return ""
	}

	path := strings.Trim(parsed.Path, "/")
	path = strings.TrimSuffix(path, ".git")

	return strings.ToLower(path)
}
