package handler

import (
	"context"
	"fmt"
	"log/slog"

	gh "github.com/google/go-github/v68/github"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/tenant"
)

// GitHubRepository represents a repo accessible via the workspace's GitHub App installation.
type GitHubRepository struct {
	ID            string
	Name          string
	FullName      string
	HTMLURL       string
	DefaultBranch string
	Private       bool
}

// GitHubRepositories lists repos accessible to the workspace's GitHub App installation.
// Uses the workspace's installation ID to mint a short-lived installation token.
func (c *Client) GitHubRepositories(ctx context.Context) ([]GitHubRepository, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}

	ghToken, err := c.installationToken(ctx)
	if err != nil {
		return nil, err
	}

	client := gh.NewClient(nil).WithAuthToken(ghToken)

	var result []GitHubRepository
	opts := &gh.ListOptions{PerPage: 100}

	for {
		repos, resp, err := client.Apps.ListRepos(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, r := range repos.Repositories {
			result = append(result, GitHubRepository{
				ID:            fmt.Sprintf("%d", r.GetID()),
				Name:          r.GetName(),
				FullName:      r.GetFullName(),
				HTMLURL:       r.GetHTMLURL(),
				DefaultBranch: r.GetDefaultBranch(),
				Private:       r.GetPrivate(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return result, nil
}

// installationToken mints a GitHub App installation token for the current workspace.
// It fetches the workspace's installation ID from the deployer, then uses the
// GitHub App's private key to create a short-lived token.
func (c *Client) installationToken(ctx context.Context) (string, error) {
	if c.GitHubApp == nil {
		return "", fmt.Errorf("github app not configured")
	}

	ws, err := tenant.Require(ctx)
	if err != nil {
		return "", err
	}

	// Fetch workspace metadata to get the installation ID
	metaCtx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()

	outCtx := auth.OutgoingContext(metaCtx)
	outCtx = tenant.OutgoingContext(outCtx)

	resp, err := c.Deployer.WorkspaceMetadata(outCtx, &deployer.WorkspaceMetadataRequest{
		Workspace: ws,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get workspace metadata: %w", err)
	}

	if resp.GithubInstallationId == 0 {
		return "", fmt.Errorf("workspace %q has no GitHub App installation linked", ws)
	}

	token, err := c.GitHubApp.InstallationToken(ctx, resp.GithubInstallationId)
	if err != nil {
		return "", fmt.Errorf("failed to mint installation token: %w", err)
	}

	slog.Debug("minted installation token", "workspace", ws, "installation_id", resp.GithubInstallationId)
	return token, nil
}

// withInstallationToken mints an installation token for the current workspace
// and attaches it to the context for gRPC propagation via auth.OutgoingContext.
// Best-effort — returns the original context unchanged if the GitHub App is not
// configured or the workspace has no installation linked.
func (c *Client) withInstallationToken(ctx context.Context) context.Context {
	token, err := c.installationToken(ctx)
	if err != nil {
		slog.Debug("skipping installation token", "reason", err)
		return ctx
	}
	return auth.WithGitHubToken(ctx, token)
}
