package github

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v68/github"
)

// CommitMessage fetches the commit message for the given SHA in the given repository.
// repository must be in "owner/repo" form. installationID selects the GitHub App
// installation whose token is used to authenticate the request.
func (a *App) CommitMessage(ctx context.Context, installationID int64, repository, sha string) (string, error) {
	owner, repo, ok := strings.Cut(repository, "/")
	if !ok || owner == "" || repo == "" {
		return "", fmt.Errorf("repository must be in owner/repo format, got %q", repository)
	}

	token, err := a.InstallationToken(ctx, installationID)
	if err != nil {
		return "", err
	}

	client := gh.NewClient(nil).WithAuthToken(token)

	commit, _, err := client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return "", fmt.Errorf("get commit %s/%s@%s: %w", owner, repo, sha, err)
	}

	return commit.GetCommit().GetMessage(), nil
}

// CommitSHA resolves a ref (branch, tag, or SHA) to a commit SHA on the given
// repository. repository must be in "owner/repo" form. installationID selects
// the GitHub App installation whose token is used to authenticate the request.
func (a *App) CommitSHA(ctx context.Context, installationID int64, repository, ref string) (string, error) {
	owner, repo, ok := strings.Cut(repository, "/")

	if !ok || owner == "" || repo == "" {
		return "", fmt.Errorf("repository must be in owner/repo format, got %q", repository)
	}

	token, err := a.InstallationToken(ctx, installationID)

	if err != nil {
		return "", err
	}

	client := gh.NewClient(nil).WithAuthToken(token)

	commit, _, err := client.Repositories.GetCommit(ctx, owner, repo, ref, nil)

	if err != nil {
		return "", fmt.Errorf("get commit %s@%s: %w", repository, ref, err)
	}

	return commit.GetSHA(), nil
}
