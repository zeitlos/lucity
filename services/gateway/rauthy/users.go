package rauthy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// User represents a Rauthy user.
type User struct {
	ID              string   `json:"id"`
	Email           string   `json:"email"`
	GivenName       string   `json:"given_name"`
	FamilyName      string   `json:"family_name"`
	Roles           []string `json:"roles"`           // role names — required for PUT round-trip
	Groups          []string `json:"groups"`           // group IDs
	Enabled         bool     `json:"enabled"`
	EmailVerified   bool     `json:"email_verified"`
	WebauthnEnabled bool     `json:"webauthn_enabled"`
}

// Name returns the user's display name, falling back to email.
func (u *User) Name() string {
	if u.GivenName != "" {
		if u.FamilyName != "" {
			return u.GivenName + " " + u.FamilyName
		}
		return u.GivenName
	}
	return u.Email
}

// Users returns all users from Rauthy.
func (c *Client) Users(ctx context.Context) ([]User, error) {
	var users []User
	if err := c.doJSON(ctx, "GET", "/users", nil, &users); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// User returns a single user by ID.
func (c *Client) User(ctx context.Context, id string) (*User, error) {
	var user User
	if err := c.doJSON(ctx, "GET", "/users/"+id, nil, &user); err != nil {
		return nil, fmt.Errorf("failed to get user %q: %w", id, err)
	}
	return &user, nil
}

// UpdateUserGroups sets the groups for a user by updating the full user object.
// This requires a PUT with the complete user payload.
func (c *Client) UpdateUserGroups(ctx context.Context, userID string, groupIDs []string) error {
	// First fetch the current user to get all fields
	user, err := c.User(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user for group update: %w", err)
	}

	user.Groups = groupIDs

	payload, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	if err := c.doNoContent(ctx, "PUT", "/users/"+userID, bytes.NewReader(payload)); err != nil {
		return fmt.Errorf("failed to update user groups: %w", err)
	}
	return nil
}

// UsersByGroupID returns all users that belong to a specific group (by group ID).
func (c *Client) UsersByGroupID(ctx context.Context, groupID string) ([]User, error) {
	all, err := c.Users(ctx)
	if err != nil {
		return nil, err
	}

	var members []User
	for _, u := range all {
		for _, gid := range u.Groups {
			if gid == groupID {
				members = append(members, u)
				break
			}
		}
	}
	return members, nil
}

// UserByEmail finds a user by their email address. Returns nil if not found.
func (c *Client) UserByEmail(ctx context.Context, email string) (*User, error) {
	all, err := c.Users(ctx)
	if err != nil {
		return nil, err
	}

	for _, u := range all {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, nil
}
