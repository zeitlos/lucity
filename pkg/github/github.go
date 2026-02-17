package github

import (
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

// App holds the GitHub App configuration for OAuth and API access.
type App struct {
	appID         int64
	clientID      string
	clientSecret  string
	webhookSecret string
	oauthConfig   *oauth2.Config
}

// User represents a GitHub user profile.
type User struct {
	Login     string
	Name      string
	Email     string
	AvatarURL string
}

// NewApp creates a new GitHub App client.
func NewApp(appID int64, clientID, clientSecret, webhookSecret, callbackURL string) *App {
	return &App{
		appID:         appID,
		clientID:      clientID,
		clientSecret:  clientSecret,
		webhookSecret: webhookSecret,
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     githubOAuth.Endpoint,
			RedirectURL:  callbackURL,
			Scopes:       []string{"read:user", "user:email"},
		},
	}
}

// WebhookSecret returns the configured webhook secret for signature validation.
func (a *App) WebhookSecret() string {
	return a.webhookSecret
}
