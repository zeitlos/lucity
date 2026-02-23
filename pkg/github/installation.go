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

// InstallationClient creates an authenticated GitHub client for the given installation.
// Uses the App's private key to generate a short-lived installation access token.
// This is used for automated operations where no user OAuth token is available.
func (a *App) InstallationClient(ctx context.Context, installationID int64) (*gh.Client, error) {
	transport, err := ghinstallation.New(http.DefaultTransport, a.appID, installationID, a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create installation transport: %w", err)
	}

	return gh.NewClient(&http.Client{Transport: transport}), nil
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

func repoFromGitHub(r *gh.Repository) Repository {
	repo := Repository{
		ID:            r.GetID(),
		Name:          r.GetName(),
		FullName:      r.GetFullName(),
		CloneURL:      r.GetCloneURL(),
		HTMLURL:       r.GetHTMLURL(),
		DefaultBranch: r.GetDefaultBranch(),
		Private:       r.GetPrivate(),
	}
	if r.Owner != nil {
		repo.Owner = r.Owner.GetLogin()
	}
	return repo
}
