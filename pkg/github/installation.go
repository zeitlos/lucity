package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	gh "github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// InstallationID discovers the GitHub App installation ID for the authenticated user.
// Uses the user's OAuth token to query their installations and returns the one
// matching this App's ID.
func (a *App) InstallationID(ctx context.Context, userToken *oauth2.Token) (int64, error) {
	client := gh.NewClient(a.oauthConfig.Client(ctx, userToken))

	installations, _, err := client.Apps.ListUserInstallations(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to list user installations: %w", err)
	}

	for _, inst := range installations {
		if inst.GetAppID() == a.appID {
			return inst.GetID(), nil
		}
	}

	return 0, fmt.Errorf("github app not installed for this user")
}

// InstallationToken creates a short-lived installation access token string.
// This is used by the webhook service to authenticate gRPC calls when no user session exists.
func (a *App) InstallationToken(ctx context.Context, installationID int64) (string, error) {
	transport, err := ghinstallation.New(http.DefaultTransport, a.appID, installationID, a.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to create installation transport: %w", err)
	}

	token, err := transport.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get installation token: %w", err)
	}

	return token, nil
}

// Installation represents a GitHub App installation on an account.
type Installation struct {
	ID            int64
	AccountLogin  string
	AccountAvatar string
	AccountType   string // "Organization" or "User"
}

// appClient creates a go-github client authenticated as the GitHub App (JWT).
func (a *App) appClient() (*gh.Client, error) {
	if len(a.privateKey) == 0 {
		return nil, fmt.Errorf("github app private key not configured")
	}

	transport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, a.appID, a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create app transport: %w", err)
	}

	return gh.NewClient(&http.Client{Transport: transport}), nil
}

// Installations lists all installations of this GitHub App.
// Uses the App's private key to authenticate as the app itself (JWT).
func (a *App) Installations(ctx context.Context) ([]Installation, error) {
	client, err := a.appClient()
	if err != nil {
		return nil, err
	}

	var result []Installation
	opts := &gh.ListOptions{PerPage: 100}

	for {
		installations, resp, err := client.Apps.ListInstallations(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list installations: %w", err)
		}

		for _, inst := range installations {
			result = append(result, Installation{
				ID:            inst.GetID(),
				AccountLogin:  inst.GetAccount().GetLogin(),
				AccountAvatar: inst.GetAccount().GetAvatarURL(),
				AccountType:   inst.GetTargetType(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return result, nil
}

// Installation fetches a single GitHub App installation by ID.
func (a *App) Installation(ctx context.Context, installationID int64) (*Installation, error) {
	client, err := a.appClient()
	if err != nil {
		return nil, err
	}

	inst, _, err := client.Apps.GetInstallation(ctx, installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installation: %w", err)
	}

	return &Installation{
		ID:            inst.GetID(),
		AccountLogin:  inst.GetAccount().GetLogin(),
		AccountAvatar: inst.GetAccount().GetAvatarURL(),
		AccountType:   inst.GetTargetType(),
	}, nil
}

