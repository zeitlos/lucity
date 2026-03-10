package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	gh "github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
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

// GitHubConnected returns whether the current user has a stored GitHub OAuth token.
func (c *Client) GitHubConnected(ctx context.Context) (bool, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return false, fmt.Errorf("unauthenticated")
	}

	resp, err := c.Deployer.UserGitHubToken(ctx, &deployer.UserGitHubTokenRequest{
		UserId: claims.Subject,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check github token: %w", err)
	}

	return resp.Connected, nil
}

// GitHubSources returns all GitHub App installations accessible to the user.
// Requires a connected GitHub account (per-user OAuth token).
func (c *Client) GitHubSources(ctx context.Context) ([]GitHubInstallation, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, fmt.Errorf("unauthenticated")
	}

	if c.GitHubApp == nil {
		return nil, fmt.Errorf("github app not configured")
	}

	token, err := c.userGitHubToken(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}

	installations, err := c.GitHubApp.UserInstallations(ctx, token)
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

// StoreGitHubToken stores a GitHub OAuth token for the given user.
func (c *Client) StoreGitHubToken(ctx context.Context, userID string, token *oauth2.Token) error {
	var expiresAt int64
	if !token.Expiry.IsZero() {
		expiresAt = token.Expiry.Unix()
	}

	_, err := c.Deployer.StoreUserGitHubToken(ctx, &deployer.StoreUserGitHubTokenRequest{
		UserId:       userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    expiresAt,
	})
	return err
}

// userGitHubToken fetches the user's stored GitHub OAuth token and refreshes if expired.
func (c *Client) userGitHubToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	resp, err := c.Deployer.UserGitHubToken(ctx, &deployer.UserGitHubTokenRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get github token: %w", err)
	}

	if !resp.Connected {
		return nil, fmt.Errorf("github account not connected")
	}

	token := &oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
	if resp.ExpiresAt > 0 {
		token.Expiry = time.Unix(resp.ExpiresAt, 0)
	}

	// Auto-refresh if expired
	if !token.Expiry.IsZero() && token.Expiry.Before(time.Now()) && token.RefreshToken != "" {
		fresh, err := c.GitHubApp.RefreshToken(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh github token: %w", err)
		}

		// Store the refreshed token
		if err := c.StoreGitHubToken(ctx, userID, fresh); err != nil {
			slog.Warn("failed to store refreshed github token", "error", err)
		}

		return fresh, nil
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
