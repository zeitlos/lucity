package handler

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/auth"
	gh "github.com/google/go-github/v68/github"
)

// GitHubRepository represents a repo accessible to the authenticated user.
type GitHubRepository struct {
	ID            string
	Name          string
	FullName      string
	HTMLURL       string
	DefaultBranch string
	Private       bool
}

// GitHubRepositories lists repos accessible to the authenticated user.
func (c *Client) GitHubRepositories(ctx context.Context) ([]GitHubRepository, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}
	if claims.GitHubToken == "" {
		return nil, fmt.Errorf("no github token in session")
	}

	client := gh.NewClient(nil).WithAuthToken(claims.GitHubToken)

	var result []GitHubRepository
	opts := &gh.RepositoryListOptions{
		Sort:        "updated",
		ListOptions: gh.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := client.Repositories.List(ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, r := range repos {
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
