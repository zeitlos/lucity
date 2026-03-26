package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	gh "github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/logto"
)

// GitHubInstallation represents a GitHub App installation on an account.
type GitHubInstallation struct {
	ID               string
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
	if c.GitHubApp == nil {
		return nil, fmt.Errorf("github app not configured")
	}

	ghToken, err := c.userGitHubToken(ctx)
	if err != nil {
		return nil, err
	}

	installations, err := c.GitHubApp.UserInstallations(ctx, &oauth2.Token{AccessToken: ghToken})
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
			ID:               fmt.Sprintf("%d", inst.ID),
			AccountLogin:     inst.AccountLogin,
			AccountAvatarURL: inst.AccountAvatar,
			AccountType:      accountType,
		})
	}

	return result, nil
}

// GitHubRepositories lists repos accessible from a specific installation.
// Uses the App's private key to mint an installation token for the given installation ID.
func (c *Client) GitHubRepositories(ctx context.Context, installationID string) ([]GitHubRepository, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}

	if c.GitHubApp == nil {
		return nil, fmt.Errorf("github app not configured")
	}

	instID, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid installation ID: %w", err)
	}

	ghToken, err := c.GitHubApp.InstallationToken(ctx, instID)
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

// userGitHubToken retrieves the user's GitHub OAuth token from Logto's Account API.
// The user must have signed in via GitHub (social sign-in) for a token to be available.
// If the Logto access token is expired, it transparently refreshes using the refresh token
// and updates the response cookies.
func (c *Client) userGitHubToken(ctx context.Context) (string, error) {
	logtoToken := auth.TokenFrom(ctx)
	if logtoToken == "" {
		return "", fmt.Errorf("no Logto access token in context")
	}

	token, err := c.Logto.GitHubToken(ctx, logtoToken)
	if err == nil {
		return token, nil
	}

	// If the token is expired, try refreshing it
	if !errors.Is(err, logto.ErrTokenExpired) {
		return "", fmt.Errorf("failed to get github token: %w", err)
	}

	if c.TokenRefresher == nil {
		return "", fmt.Errorf("failed to get github token (token expired, no refresher configured): %w", err)
	}

	refreshToken := auth.RefreshTokenFrom(ctx)
	if refreshToken == "" {
		return "", fmt.Errorf("failed to get github token (token expired, no refresh token): %w", err)
	}

	slog.Info("logto access token expired, refreshing")

	newAccessToken, refreshErr := c.TokenRefresher(ctx, refreshToken)
	if refreshErr != nil {
		return "", fmt.Errorf("failed to get github token (token expired, refresh failed: %v): %w", refreshErr, err)
	}

	// Retry with the refreshed token
	token, err = c.Logto.GitHubToken(ctx, newAccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to get github token after refresh: %w", err)
	}
	return token, nil
}

// installationTokenForService mints a GitHub App installation token for a specific
// installation ID. Used by commit enrichment where the installation ID comes from
// the service's metadata (not from a workspace).
func (c *Client) installationTokenForService(ctx context.Context, installationID int64) (string, error) {
	if c.GitHubApp == nil {
		return "", fmt.Errorf("github app not configured")
	}

	if installationID == 0 {
		return "", fmt.Errorf("service has no GitHub installation linked")
	}

	token, err := c.GitHubApp.InstallationToken(ctx, installationID)
	if err != nil {
		return "", fmt.Errorf("failed to mint installation token: %w", err)
	}

	return token, nil
}
