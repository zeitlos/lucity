package handler

import (
	"context"

	"github.com/zeitlos/lucity/pkg/auth"
)

// Me returns the current user's profile from the JWT claims.
func (c *Client) Me(ctx context.Context) (*User, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, nil
	}

	return &User{
		Name:       claims.Subject,
		Email:      claims.Email,
		AvatarURL:  claims.AvatarURL,
		Workspaces: claims.Workspaces,
	}, nil
}

// User represents an authenticated user's profile.
type User struct {
	Name       string
	Email      string
	AvatarURL  string
	Workspaces []auth.WorkspaceMembership
}
