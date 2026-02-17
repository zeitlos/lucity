package handler

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/auth"
	gh "github.com/zeitlos/lucity/pkg/github"
)

// GitHubRepository represents a repo accessible via the GitHub App installation.
type GitHubRepository struct {
	ID            string
	Name          string
	FullName      string
	HTMLURL       string
	DefaultBranch string
	Private       bool
}

// GitHubRepositories lists repos the GitHub App has access to for the current user's installation.
func (c *Client) GitHubRepositories(ctx context.Context) ([]GitHubRepository, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}
	if claims.InstallationID == 0 {
		return nil, fmt.Errorf("github app not installed")
	}

	repos, err := c.GitHubApp.Repositories(ctx, claims.InstallationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	result := make([]GitHubRepository, 0, len(repos))
	for _, r := range repos {
		result = append(result, convertGHRepo(r))
	}
	return result, nil
}

func convertGHRepo(r gh.Repository) GitHubRepository {
	return GitHubRepository{
		ID:            fmt.Sprintf("%d", r.ID),
		Name:          r.Name,
		FullName:      r.FullName,
		HTMLURL:       r.HTMLURL,
		DefaultBranch: r.DefaultBranch,
		Private:       r.Private,
	}
}
