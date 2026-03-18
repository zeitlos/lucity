package logto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
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

// UserIdentity represents a social identity linked to a Logto user.
type UserIdentity struct {
	UserID  string                 `json:"userId"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// UserGitHubLogin returns the GitHub username from the user's social identities.
// Returns empty string if the user has no GitHub identity or no login.
func (c *Client) UserGitHubLogin(ctx context.Context, userID string) (string, error) {
	var result struct {
		Identities map[string]UserIdentity `json:"identities"`
	}
	if err := c.doJSON(ctx, "GET", "/api/users/"+userID, nil, &result); err != nil {
		return "", fmt.Errorf("failed to get user identities: %w", err)
	}

	gh, ok := result.Identities["github"]
	if !ok {
		return "", nil
	}

	rawData, ok := gh.Details["rawData"].(map[string]interface{})
	if !ok {
		return "", nil
	}

	// The GitHub connector wraps the API response in a "userInfo" key
	userInfo, ok := rawData["userInfo"].(map[string]interface{})
	if !ok {
		return "", nil
	}

	login, _ := userInfo["login"].(string)
	return login, nil
}

// logtoUsernameRe matches Logto's username validation: starts with a letter or
// underscore, followed by word characters only (letters, digits, underscores).
var logtoUsernameRe = regexp.MustCompile(`[^A-Za-z0-9_]`)

// sanitizeLogtoUsername converts a GitHub login into a Logto-compatible username.
// Logto's regex is /^[A-Z_a-z]\w*$/ — no hyphens or dots allowed.
func sanitizeLogtoUsername(login string) string {
	s := logtoUsernameRe.ReplaceAllString(login, "_")
	s = strings.Trim(s, "_")
	if s == "" || (s[0] >= '0' && s[0] <= '9') {
		s = "_" + s
	}
	return s
}

// UpdateUsername sets the username on a Logto user.
// Sanitizes the input to match Logto's username regex (/^[A-Z_a-z]\w*$/).
func (c *Client) UpdateUsername(ctx context.Context, userID, username string) error {
	safe := sanitizeLogtoUsername(username)
	body, _ := json.Marshal(map[string]string{"username": safe})
	return c.doJSON(ctx, "PATCH", "/api/users/"+userID, bytes.NewReader(body), &json.RawMessage{})
}

// UserOrganizations returns all organizations a user belongs to.
func (c *Client) UserOrganizations(ctx context.Context, userID string) ([]UserOrganization, error) {
	var orgs []UserOrganization
	if err := c.doJSON(ctx, "GET", "/api/users/"+userID+"/organizations", nil, &orgs); err != nil {
		return nil, fmt.Errorf("failed to get organizations for user %q: %w", userID, err)
	}
	return orgs, nil
}
