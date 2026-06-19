package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/services/conductor/internal/source"
)

type Client struct {
	app *github.App
}

func New(app *github.App) *Client {
	return &Client{app: app}
}

func (c *Client) Commit(ctx context.Context, repoURL, ref string) (source.Commit, error) {
	if ref == "" {
		ref = "HEAD" // GitHub treats this as the head commit of the default branch
	}

	repository, err := parseRepoURL(repoURL)

	if err != nil {
		return source.Commit{}, err
	}

	installationID, err := c.app.FindInstallation(ctx, repository)

	if err != nil {
		return source.Commit{}, err
	}

	commit, err := c.app.Commit(ctx, installationID, repository, ref)

	if err != nil {
		return source.Commit{}, err
	}

	return source.Commit{SHA: commit.SHA, Message: commit.Message}, nil
}

func (c *Client) Token(ctx context.Context, repoURL string) (string, error) {
	repository, err := parseRepoURL(repoURL)

	if err != nil {
		return "", err
	}

	installationID, err := c.app.FindInstallation(ctx, repository)

	if err != nil {
		return "", err
	}

	return c.app.InstallationTokenForRepo(ctx, installationID, repository)
}

// parseRepoURL extracts the "owner/repo" path from a GitHub HTTPS URL,
// stripping the optional ".git" suffix and any trailing slashes.
func parseRepoURL(repoURL string) (string, error) {
	parsed, err := url.Parse(repoURL)

	if err != nil {
		return "", fmt.Errorf("parse repo url %q: %w", repoURL, err)
	}

	path := strings.Trim(parsed.Path, "/")
	path = strings.TrimSuffix(path, ".git")

	parts := strings.SplitN(path, "/", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid github repo url %q: expected owner/repo path", repoURL)
	}

	return path, nil
}

var _ source.Interface = (*Client)(nil)
