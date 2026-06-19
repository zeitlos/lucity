package github

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v68/github"
)

type Commit struct {
	SHA     string
	Message string
}

func (a *App) Commit(ctx context.Context, installationID int64, repository, ref string) (Commit, error) {
	owner, repo, ok := strings.Cut(repository, "/")

	if !ok || owner == "" || repo == "" {
		return Commit{}, fmt.Errorf("repository must be in owner/repo format, got %q", repository)
	}

	token, err := a.InstallationToken(ctx, installationID)

	if err != nil {
		return Commit{}, err
	}

	client := gh.NewClient(nil).WithAuthToken(token)

	commit, _, err := client.Repositories.GetCommit(ctx, owner, repo, ref, nil)

	if err != nil {
		return Commit{}, fmt.Errorf("get commit %s@%s: %w", repository, ref, err)
	}

	return Commit{
		SHA:     commit.GetSHA(),
		Message: commit.GetCommit().GetMessage(),
	}, nil
}
