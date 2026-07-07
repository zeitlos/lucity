package github

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v68/github"
)

func (a *App) Branches(ctx context.Context, installationID int64, repository string) ([]string, error) {
	owner, repo, ok := strings.Cut(repository, "/")

	if !ok || owner == "" || repo == "" {
		return nil, fmt.Errorf("repository must be in owner/repo format, got %q", repository)
	}

	token, err := a.InstallationToken(ctx, installationID)

	if err != nil {
		return nil, err
	}

	client := gh.NewClient(nil).WithAuthToken(token)

	opts := &gh.BranchListOptions{ListOptions: gh.ListOptions{PerPage: 100}}

	var branches []string

	for {
		page, resp, err := client.Repositories.ListBranches(ctx, owner, repo, opts)

		if err != nil {
			return nil, fmt.Errorf("list branches %s: %w", repository, err)
		}

		for _, branch := range page {
			branches = append(branches, branch.GetName())
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return branches, nil
}
