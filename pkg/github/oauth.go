package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// OAuthURL returns the GitHub OAuth authorization URL with the given state parameter.
func (a *App) OAuthURL(state string) string {
	return a.oauthConfig.AuthCodeURL(state)
}

// ExchangeCode exchanges an authorization code for an OAuth2 token.
func (a *App) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := a.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// RefreshToken returns a fresh token, refreshing it if expired.
// GitHub App OAuth tokens have a configurable expiry. This uses the standard
// oauth2 token source to handle refresh transparently.
func (a *App) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	src := a.oauthConfig.TokenSource(ctx, token)
	fresh, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return fresh, nil
}

// GetUser fetches the authenticated user's profile from GitHub.
func (a *App) GetUser(ctx context.Context, token *oauth2.Token) (*User, error) {
	client := gh.NewClient(a.oauthConfig.Client(ctx, token))

	ghUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user := &User{
		Login:     ghUser.GetLogin(),
		AvatarURL: ghUser.GetAvatarURL(),
	}

	if ghUser.Name != nil {
		user.Name = *ghUser.Name
	}
	if ghUser.Email != nil {
		user.Email = *ghUser.Email
	}

	// If email is not public, fetch from the emails API
	if user.Email == "" {
		emails, _, err := client.Users.ListEmails(ctx, nil)
		if err == nil {
			for _, e := range emails {
				if e.GetPrimary() && e.GetVerified() {
					user.Email = e.GetEmail()
					break
				}
			}
		}
	}

	return user, nil
}
