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

// SocialLogin returns the login/username from the user's social identity.
// Checks connectors in order: GitHub, GitLab, Bitbucket. Returns the first
// non-empty login found, or empty string if none match.
func (c *Client) SocialLogin(ctx context.Context, userID string) (string, error) {
	var result struct {
		Identities map[string]UserIdentity `json:"identities"`
	}
	if err := c.doJSON(ctx, "GET", "/api/users/"+userID, nil, &result); err != nil {
		return "", fmt.Errorf("failed to get user identities: %w", err)
	}

	// Each connector stores the login in a slightly different path.
	// GitHub: rawData.userInfo.login
	// GitLab: rawData.userInfo.username
	// Bitbucket: rawData.userInfo.username
	connectors := []struct {
		name  string
		field string
	}{
		{"github", "login"},
		{"gitlab", "username"},
		{"bitbucket", "username"},
	}

	for _, conn := range connectors {
		identity, ok := result.Identities[conn.name]
		if !ok {
			continue
		}

		rawData, ok := identity.Details["rawData"].(map[string]interface{})
		if !ok {
			continue
		}

		userInfo, ok := rawData["userInfo"].(map[string]interface{})
		if !ok {
			continue
		}

		if login, _ := userInfo[conn.field].(string); login != "" {
			return login, nil
		}
	}

	return "", nil
}

// logtoUsernameRe replaces characters not allowed in Logto usernames.
var logtoUsernameRe = regexp.MustCompile(`[^A-Za-z0-9_]`)

// SanitizeUsername converts a social login into a Logto-compatible username.
// Logto's regex is /^[A-Z_a-z]\w*$/ — no hyphens or dots allowed.
func SanitizeUsername(login string) string {
	s := logtoUsernameRe.ReplaceAllString(login, "_")
	s = strings.Trim(s, "_")
	if s == "" || (s[0] >= '0' && s[0] <= '9') {
		s = "_" + s
	}
	return s
}

// EnsureUsername sets the username on a Logto user, sanitizing it to match
// Logto's regex. If the username is taken, appends _0, _1, ... _9 until
// an available one is found. Returns the username that was set.
func (c *Client) EnsureUsername(ctx context.Context, userID, login string) (string, error) {
	base := SanitizeUsername(login)

	candidates := []string{base}
	for i := 0; i < 10; i++ {
		candidates = append(candidates, fmt.Sprintf("%s_%d", base, i))
	}

	for _, candidate := range candidates {
		body, _ := json.Marshal(map[string]string{"username": candidate})
		err := c.doJSON(ctx, "PATCH", "/api/users/"+userID, bytes.NewReader(body), &json.RawMessage{})
		if err == nil {
			return candidate, nil
		}
		// If the error is not a conflict, stop trying.
		if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "exists") && !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "409") {
			return "", fmt.Errorf("failed to set username %q: %w", candidate, err)
		}
	}

	return "", fmt.Errorf("all username candidates exhausted for base %q", base)
}

// UserOrganizations returns all organizations a user belongs to.
func (c *Client) UserOrganizations(ctx context.Context, userID string) ([]UserOrganization, error) {
	var orgs []UserOrganization
	if err := c.doJSON(ctx, "GET", "/api/users/"+userID+"/organizations", nil, &orgs); err != nil {
		return nil, fmt.Errorf("failed to get organizations for user %q: %w", userID, err)
	}
	return orgs, nil
}
