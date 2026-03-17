package logto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Organization represents a Logto organization.
type Organization struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	CustomData  map[string]interface{} `json:"customData,omitempty"`
	CreatedAt   int64                  `json:"createdAt,omitempty"`
}

// OrganizationRole represents a role assigned to a user within an organization.
type OrganizationRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// OrganizationMember represents a user in an organization with their roles.
type OrganizationMember struct {
	ID        string             `json:"id"`
	Name      string             `json:"name,omitempty"`
	Email     string             `json:"primaryEmail,omitempty"`
	Avatar    string             `json:"avatar,omitempty"`
	OrgRoles  []OrganizationRole `json:"organizationRoles,omitempty"`
}

// Organization returns a single organization by its Logto internal ID.
func (c *Client) Organization(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	if err := c.doJSON(ctx, "GET", "/api/organizations/"+id, nil, &org); err != nil {
		return nil, fmt.Errorf("failed to get organization %q: %w", id, err)
	}
	return &org, nil
}

// OrganizationByName searches for an organization by exact name match.
// The Logto search API does fuzzy matching, so we filter client-side for an exact match.
func (c *Client) OrganizationByName(ctx context.Context, name string) (*Organization, error) {
	path := "/api/organizations?q=" + url.QueryEscape(name) + "&page=1&page_size=20"
	var orgs []Organization
	if err := c.doJSON(ctx, "GET", path, nil, &orgs); err != nil {
		return nil, fmt.Errorf("failed to search organizations by name %q: %w", name, err)
	}
	for _, org := range orgs {
		if org.Name == name {
			return &org, nil
		}
	}
	return nil, fmt.Errorf("organization with name %q not found", name)
}

// CreateOrganization creates a new organization.
// The name parameter is used as the org name (workspace ID). The displayName is stored
// in customData only if it differs from the name.
// Idempotent: if an org with the same name exists, returns it.
func (c *Client) CreateOrganization(ctx context.Context, name, displayName string, customData map[string]interface{}) (*Organization, error) {
	payload := map[string]interface{}{
		"name": name,
	}
	if customData == nil {
		customData = make(map[string]interface{})
	}
	if displayName != "" && displayName != name {
		customData["displayName"] = displayName
	}
	if len(customData) > 0 {
		payload["customData"] = customData
	}

	body, _ := json.Marshal(payload)
	var org Organization
	if err := c.doJSON(ctx, "POST", "/api/organizations", bytes.NewReader(body), &org); err != nil {
		// Handle conflict (already exists)
		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "already exists") {
			existing, findErr := c.OrganizationByName(ctx, name)
			if findErr != nil {
				return nil, fmt.Errorf("org %q already exists but failed to look up: %w", name, findErr)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("failed to create organization %q: %w", name, err)
	}
	return &org, nil
}

// UpdateOrganization updates an organization's name and/or description.
func (c *Client) UpdateOrganization(ctx context.Context, id string, name string) error {
	payload, _ := json.Marshal(map[string]string{"name": name})
	return c.doNoContent(ctx, "PATCH", "/api/organizations/"+id, bytes.NewReader(payload))
}

// DeleteOrganization deletes an organization.
// Idempotent: returns nil if already deleted.
func (c *Client) DeleteOrganization(ctx context.Context, id string) error {
	err := c.doNoContent(ctx, "DELETE", "/api/organizations/"+id, nil)
	if err != nil && strings.Contains(err.Error(), "404") {
		return nil
	}
	return err
}

// UpdateOrganizationCustomData replaces the custom data for an organization.
func (c *Client) UpdateOrganizationCustomData(ctx context.Context, orgID string, data map[string]interface{}) error {
	body, _ := json.Marshal(map[string]interface{}{"customData": data})
	return c.doNoContent(ctx, "PATCH", "/api/organizations/"+orgID, bytes.NewReader(body))
}

// OrganizationMembers returns all members of an organization.
func (c *Client) OrganizationMembers(ctx context.Context, orgID string) ([]OrganizationMember, error) {
	var members []OrganizationMember
	if err := c.doJSON(ctx, "GET", "/api/organizations/"+orgID+"/users", nil, &members); err != nil {
		return nil, fmt.Errorf("failed to list members of org %q: %w", orgID, err)
	}
	return members, nil
}

// AddOrganizationMember adds a user to an organization.
func (c *Client) AddOrganizationMember(ctx context.Context, orgID, userID string) error {
	payload, _ := json.Marshal(map[string][]string{"userIds": {userID}})
	return c.doNoContent(ctx, "POST", "/api/organizations/"+orgID+"/users", bytes.NewReader(payload))
}

// RemoveOrganizationMember removes a user from an organization.
func (c *Client) RemoveOrganizationMember(ctx context.Context, orgID, userID string) error {
	return c.doNoContent(ctx, "DELETE", "/api/organizations/"+orgID+"/users/"+userID, nil)
}

// AssignOrganizationRoles assigns roles to a user in an organization.
func (c *Client) AssignOrganizationRoles(ctx context.Context, orgID, userID string, roleIDs []string) error {
	payload, _ := json.Marshal(map[string][]string{"organizationRoleIds": roleIDs})
	return c.doNoContent(ctx, "POST", "/api/organizations/"+orgID+"/users/"+userID+"/roles", bytes.NewReader(payload))
}

// RemoveOrganizationRole removes a role from a user in an organization.
func (c *Client) RemoveOrganizationRole(ctx context.Context, orgID, userID, roleID string) error {
	return c.doNoContent(ctx, "DELETE", "/api/organizations/"+orgID+"/users/"+userID+"/roles/"+roleID, nil)
}

// OrganizationRoles returns all organization roles defined in Logto.
func (c *Client) OrganizationRoles(ctx context.Context) ([]OrganizationRole, error) {
	var roles []OrganizationRole
	if err := c.doJSON(ctx, "GET", "/api/organization-roles", nil, &roles); err != nil {
		return nil, fmt.Errorf("failed to list organization roles: %w", err)
	}
	return roles, nil
}

// MemberRoles returns the roles assigned to a user in an organization.
func (c *Client) MemberRoles(ctx context.Context, orgID, userID string) ([]OrganizationRole, error) {
	var roles []OrganizationRole
	if err := c.doJSON(ctx, "GET", "/api/organizations/"+orgID+"/users/"+userID+"/roles", nil, &roles); err != nil {
		return nil, fmt.Errorf("failed to get roles for user %q in org %q: %w", userID, orgID, err)
	}
	return roles, nil
}
