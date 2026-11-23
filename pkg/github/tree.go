package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v92/github"
)

var ErrTreeTruncated = errors.New("tree has too many entries to list")

// TreeSize sums the sizes of every file in the tree of a commit.
func (a *App) TreeSize(ctx context.Context, installationID int64, repository, sha string) (int64, error) {
	owner, repo, ok := strings.Cut(repository, "/")

	if !ok || owner == "" || repo == "" {
		return 0, fmt.Errorf("repository must be in owner/repo format, got %q", repository)
	}

	token, err := a.InstallationToken(ctx, installationID)

	if err != nil {
		return 0, err
	}

	client, err := gh.NewClient(gh.WithAuthToken(token))

	if err != nil {
		return 0, err
	}

	tree, _, err := client.Git.GetTree(ctx, owner, repo, sha, true)

	if err != nil {
		return 0, fmt.Errorf("get tree %s@%s: %w", repository, sha, err)
	}

	if tree.GetTruncated() {
		return 0, ErrTreeTruncated
	}

	var size int64

	for _, entry := range tree.Entries {
		size += int64(entry.GetSize())
	}

	return size, nil
}
