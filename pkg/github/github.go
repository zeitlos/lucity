package github

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

type App struct {
	appID       int64
	privateKey  []byte
	oauthConfig *oauth2.Config
}

// NewApp creates a new GitHub App client.
// privateKeyPath is the path to the GitHub App's PEM private key file.
func NewApp(appID int64, clientID, clientSecret, privateKeyPath string) (*App, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	return &App{
		appID:      appID,
		privateKey: key,
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     githubOAuth.Endpoint,
		},
	}, nil
}
