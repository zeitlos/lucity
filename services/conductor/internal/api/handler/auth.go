package handler

import (
	"context"

	"github.com/zeitlos/lucity/pkg/auth"
)

type User struct {
	Name       string
	Email      string
	AvatarURL  string
	Workspaces []auth.WorkspaceMembership
}

func (c *Client) Me(ctx context.Context) (*User, error) {
	claims, err := auth.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	return &User{
		Name:       claims.Name,
		Email:      claims.Email,
		AvatarURL:  claims.AvatarURL,
		Workspaces: claims.Workspaces,
	}, nil
}
