package logto

import (
	"context"
	"fmt"
	"net/url"
)

// User represents a Logto user.
type User struct {
	ID           string `json:"id"`
	PrimaryEmail string `json:"primaryEmail,omitempty"`
	Username     string `json:"username,omitempty"`
	Name         string `json:"name,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
}

// UserWithOrgs represents a Logto user with their organization memberships.
type UserOrganization struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	CustomData  map[string]interface{} `json:"customData,omitempty"`
}

// User returns a single user by ID.
func (c *Client) User(ctx context.Context, id string) (*User, error) {
	var user User
	if err := c.doJSON(ctx, "GET", "/api/users/"+id, nil, &user); err != nil {
		return nil, fmt.Errorf("failed to get user %q: %w", id, err)
	}
	return &user, nil
}

// UserByEmail finds a user by email. Returns nil if not found.
func (c *Client) UserByEmail(ctx context.Context, email string) (*User, error) {
	var users []User
	path := "/api/users?search=" + url.QueryEscape(email)
	if err := c.doJSON(ctx, "GET", path, nil, &users); err != nil {
		return nil, fmt.Errorf("failed to search users by email: %w", err)
	}

	for _, u := range users {
		if u.PrimaryEmail == email {
			return &u, nil
		}
	}
	return nil, nil
}

// UserOrganizations returns all organizations a user belongs to.
func (c *Client) UserOrganizations(ctx context.Context, userID string) ([]UserOrganization, error) {
	var orgs []UserOrganization
	if err := c.doJSON(ctx, "GET", "/api/users/"+userID+"/organizations", nil, &orgs); err != nil {
		return nil, fmt.Errorf("failed to get organizations for user %q: %w", userID, err)
	}
	return orgs, nil
}
