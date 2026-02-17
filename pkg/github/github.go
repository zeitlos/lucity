package github

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

// App holds the GitHub App configuration for OAuth and API access.
type App struct {
	appID         int64
	clientID      string
	clientSecret  string
	webhookSecret string
	privateKey    []byte
	oauthConfig   *oauth2.Config
}

// User represents a GitHub user profile.
type User struct {
	Login     string
	Name      string
	Email     string
	AvatarURL string
}

// Repository represents a GitHub repository.
type Repository struct {
	ID            int64
	Name          string
	FullName      string // "org/repo"
	CloneURL      string
	HTMLURL       string
	DefaultBranch string
	Private       bool
	Owner         string // org or user login
}

// NewApp creates a new GitHub App client.
// privateKeyPath is the path to the GitHub App's PEM private key file.
// If empty, the app will work for OAuth but installation token features will be unavailable.
func NewApp(appID int64, clientID, clientSecret, webhookSecret, callbackURL, privateKeyPath string) (*App, error) {
	var key []byte
	if privateKeyPath != "" {
		var err error
		key, err = os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
	}

	return &App{
		appID:         appID,
		clientID:      clientID,
		clientSecret:  clientSecret,
		webhookSecret: webhookSecret,
		privateKey:    key,
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     githubOAuth.Endpoint,
			RedirectURL:  callbackURL,
			Scopes:       []string{},
		},
	}, nil
}

// WebhookSecret returns the configured webhook secret for signature validation.
func (a *App) WebhookSecret() string {
	return a.webhookSecret
}
