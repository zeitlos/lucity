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
func (a *App) InstallationClient(ctx context.Context, installationID int64) (*gh.Client, error) {
	transport, err := ghinstallation.New(http.DefaultTransport, a.appID, installationID, a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create installation transport: %w", err)
	}

	return gh.NewClient(&http.Client{Transport: transport}), nil
}

// Repositories lists all repositories accessible via the given installation.
func (a *App) Repositories(ctx context.Context, installationID int64) ([]Repository, error) {
	client, err := a.InstallationClient(ctx, installationID)
	if err != nil {
		return nil, err
	}

	var repos []Repository
	opts := &gh.ListOptions{PerPage: 100}

	for {
		result, resp, err := client.Apps.ListRepos(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list installation repos: %w", err)
		}

		for _, r := range result.Repositories {
			repos = append(repos, repoFromGitHub(r))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return repos, nil
}

// CreateRepository creates a new repository in the given org or user account.
// For organizations, it uses the installation token. For personal accounts,
// GitHub App installation tokens cannot call POST /user/repos, so the user's
// OAuth token is required.
func (a *App) CreateRepository(ctx context.Context, installationID int64, org, name, userToken string, private bool) (*Repository, error) {
	repoOpts := &gh.Repository{
		Name:     gh.Ptr(name),
		Private:  gh.Ptr(private),
		AutoInit: gh.Ptr(false),
	}

	// Try org endpoint with installation token first
	client, err := a.InstallationClient(ctx, installationID)
	if err != nil {
		return nil, err
	}

	repo, _, err := client.Repositories.Create(ctx, org, repoOpts)
	if err != nil && userToken != "" {
		// Org endpoint failed (likely a personal account) — use OAuth token
		oauthClient := gh.NewClient(nil).WithAuthToken(userToken)
		repo, _, err = oauthClient.Repositories.Create(ctx, "", repoOpts)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create repository %s/%s: %w", org, name, err)
	}

	r := repoFromGitHub(repo)
	return &r, nil
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
