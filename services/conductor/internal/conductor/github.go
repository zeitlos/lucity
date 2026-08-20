package conductor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	gh "github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"

	"github.com/zeitlos/lucity/pkg/auth"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
)

// GitHubInstallation represents a GitHub App installation on an account.
type GitHubInstallation struct {
	AccountLogin     string
	AccountAvatarURL string
	AccountType      string // "ORGANIZATION" or "USER"
}

// GitHubRepository represents a repo accessible via a GitHub App installation.
type GitHubRepository struct {
	ID            string
	Name          string
	FullName      string
	HTMLURL       string
	DefaultBranch string
	Private       bool
}

// GitHubConnected returns whether the current user has a GitHub identity linked
// in Logto (via social sign-in with the GitHub App connector).
func (c *Client) GitHubConnected(ctx context.Context) (bool, error) {
	_, err := c.userGitHubToken(ctx)
	if err != nil {
		slog.Debug("github not connected", "error", err)
		return false, nil
	}
	return true, nil
}

// GitHubSources returns all GitHub App installations accessible to the user.
// Requires a connected GitHub account (GitHub identity linked in Logto via social sign-in).
func (c *Client) GitHubSources(ctx context.Context) ([]GitHubInstallation, error) {
	ghToken, err := c.userGitHubToken(ctx)

	if err != nil {
		return nil, err
	}

	installations, err := c.gitHubApp.UserInstallations(ctx, &oauth2.Token{AccessToken: ghToken})
	if err != nil {
		return nil, fmt.Errorf("failed to list user installations: %w", err)
	}

	result := make([]GitHubInstallation, 0, len(installations))
	for _, inst := range installations {
		accountType := "USER"
		if inst.AccountType == "Organization" {
			accountType = "ORGANIZATION"
		}
		result = append(result, GitHubInstallation{
			AccountLogin:     inst.AccountLogin,
			AccountAvatarURL: inst.AccountAvatar,
			AccountType:      accountType,
		})
	}

	return result, nil
}

func (c *Client) GitHubRepositories(ctx context.Context, account string) ([]GitHubRepository, error) {
	instID, err := c.installationForOwner(ctx, account)

	if err != nil {
		return nil, err
	}

	ghToken, err := c.gitHubApp.InstallationToken(ctx, instID)
	if err != nil {
		return nil, fmt.Errorf("failed to mint installation token: %w", err)
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

func (c *Client) userInstallations(ctx context.Context) ([]ghpkg.Installation, error) {
	ghToken, err := c.userGitHubToken(ctx)

	if err != nil {
		return nil, err
	}

	return c.gitHubApp.UserInstallations(ctx, &oauth2.Token{AccessToken: ghToken})
}

func (c *Client) installationForOwner(ctx context.Context, owner string) (int64, error) {
	installations, err := c.userInstallations(ctx)

	if err != nil {
		return 0, err
	}

	for _, inst := range installations {
		if strings.EqualFold(inst.AccountLogin, owner) {
			return inst.ID, nil
		}
	}

	return 0, fmt.Errorf("no accessible GitHub App installation for %q", owner)
}

func (c *Client) installationForRepo(ctx context.Context, repository string) (int64, error) {
	owner, _, ok := strings.Cut(repository, "/")

	if !ok || owner == "" {
		return 0, fmt.Errorf("repository must be in owner/repo format, got %q", repository)
	}

	return c.installationForOwner(ctx, owner)
}

// userGitHubToken retrieves the user's GitHub OAuth token from Logto's Account API.
// The user must have signed in via GitHub (social sign-in) for a token to be available.
func (c *Client) userGitHubToken(ctx context.Context) (string, error) {
	logtoToken := auth.TokenFrom(ctx)
	if logtoToken == "" {
		return "", fmt.Errorf("no Logto access token in context")
	}

	token, err := c.logto.GitHubToken(ctx, logtoToken)
	if err != nil {
		return "", fmt.Errorf("failed to get github token: %w", err)
	}
	return token, nil
}
