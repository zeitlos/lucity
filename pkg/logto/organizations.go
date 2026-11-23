package logto

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// ErrOrganizationNotFound is returned when no organization has the requested name.
var ErrOrganizationNotFound = errors.New("organization not found")

// ErrOrganizationTaken is returned when an organization, deleted or not, already
// has the requested name.
var ErrOrganizationTaken = errors.New("organization name taken")

type Organization struct {
	TenantID      string                 `json:"tenantId"`
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	CustomData    map[string]interface{} `json:"customData,omitempty"`
	IsMfaRequired bool                   `json:"isMfaRequired"`
	Color         struct {
		PrimaryColor      string `json:"primaryColor"`
		IsDarkModeEnabled bool   `json:"isDarkModeEnabled"`
		DarkPrimaryColor  string `json:"darkPrimaryColor"`
	} `json:"color"`
	Branding struct {
		LogoURL     string `json:"logoUrl"`
		DarkLogoURL string `json:"darkLogoUrl"`
		Favicon     string `json:"favicon"`
		DarkFavicon string `json:"darkFavicon"`
	} `json:"branding"`
	CustomCSS         string `json:"customCss"`
	OrganizationRoles []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"organizationRoles"`
	CreatedAt float64 `json:"createdAt"`
}

// OrganizationRole represents a role assigned to a user within an organization.
type OrganizationRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// OrganizationMember represents a user in an organization with their roles.
type OrganizationMember struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name,omitempty"`
	Email        string                  `json:"primaryEmail,omitempty"`
	Avatar       string                  `json:"avatar,omitempty"`
	OrgRoles     []OrganizationRole      `json:"organizationRoles,omitempty"`
	Identities   map[string]UserIdentity `json:"identities,omitempty"`
	CustomData   map[string]any          `json:"customData,omitempty"`
	CreatedAt    int64                   `json:"createdAt,omitempty"`
	LastSignInAt int64                   `json:"lastSignInAt,omitempty"`
}

// Deleted reports whether the organization is the tombstone of a deleted
// workspace. Its name stays taken for good.
func (o Organization) Deleted() bool {
	_, ok := o.CustomData["deleted"]
	return ok
}

// Organization returns a single organization by its Logto internal ID.
func (c *Client) Organization(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	if err := c.doJSON(ctx, "GET", "/api/organizations/"+id, nil, &org); err != nil {
		return nil, fmt.Errorf("failed to get organization %q: %w", id, err)
	}
	return &org, nil
}

// OrganizationByName returns the organization with exactly this name, deleted
// ones included. Should several share it, the oldest wins.
func (c *Client) OrganizationByName(ctx context.Context, name string) (*Organization, error) {
	matches, err := c.organizationsNamed(ctx, name)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("organization with name %q: %w", name, ErrOrganizationNotFound)
	}

	return &matches[0], nil
}

func (c *Client) organizationsNamed(ctx context.Context, name string) ([]Organization, error) {
	const pageSize = 100

	var matches []Organization
	for page := 1; ; page++ {
		path := fmt.Sprintf("/api/organizations?q=%s&page=%d&page_size=%d", url.QueryEscape(name), page, pageSize)

		var orgs []Organization
		if err := c.doJSON(ctx, "GET", path, nil, &orgs); err != nil {
			return nil, fmt.Errorf("failed to search organizations by name %q: %w", name, err)
		}

		for _, org := range orgs {
			if org.Name == name {
				matches = append(matches, org)
			}
		}

		if len(orgs) < pageSize {
			break
		}
	}

	slices.SortFunc(matches, func(a, b Organization) int {
		return cmp.Or(cmp.Compare(a.CreatedAt, b.CreatedAt), strings.Compare(a.ID, b.ID))
	})

	return matches, nil
}

// CreateOrganization creates a new organization.
// The name parameter is used as the org name (workspace ID). The displayName is stored
// in customData only if it differs from the name. Names are never reused: if another
// organization, deleted ones included, already has the name, the new one is removed
// again and ErrOrganizationTaken is returned.
func (c *Client) CreateOrganization(ctx context.Context, name, displayName string, customData map[string]any) (*Organization, error) {
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
		return nil, fmt.Errorf("failed to create organization %q: %w", name, err)
	}

	first, err := c.OrganizationByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm organization %q: %w", name, err)
	}

	if first.ID != org.ID {
		if err := c.DeleteOrganization(ctx, org.ID); err != nil {
			return nil, fmt.Errorf("failed to remove duplicate organization %q: %w", name, err)
		}
		return nil, fmt.Errorf("organization %q: %w", name, ErrOrganizationTaken)
	}

	return &org, nil
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
func (c *Client) UpdateOrganizationCustomData(ctx context.Context, orgID string, data map[string]interface{}) (*Organization, error) {
	body, _ := json.Marshal(map[string]interface{}{"customData": data})

	var org Organization
	if err := c.doJSON(ctx, "PATCH", "/api/organizations/"+orgID, bytes.NewReader(body), &org); err != nil {
		return nil, err
	}

	return &org, nil
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

// Organizations returns all organizations that are not deleted, paginating
// through all pages.
func (c *Client) Organizations(ctx context.Context) ([]Organization, error) {
	var all []Organization
	page := 1
	const pageSize = 100
	for {
		path := fmt.Sprintf("/api/organizations?page=%d&page_size=%d", page, pageSize)
		var orgs []Organization
		if err := c.doJSON(ctx, "GET", path, nil, &orgs); err != nil {
			return nil, fmt.Errorf("failed to list organizations (page %d): %w", page, err)
		}
		for _, org := range orgs {
			if !org.Deleted() {
				all = append(all, org)
			}
		}
		if len(orgs) < pageSize {
			break
		}
		page++
	}
	return all, nil
}

// MemberRoles returns the roles assigned to a user in an organization.
func (c *Client) MemberRoles(ctx context.Context, orgID, userID string) ([]OrganizationRole, error) {
	var roles []OrganizationRole
	if err := c.doJSON(ctx, "GET", "/api/organizations/"+orgID+"/users/"+userID+"/roles", nil, &roles); err != nil {
		return nil, fmt.Errorf("failed to get roles for user %q in org %q: %w", userID, orgID, err)
	}
	return roles, nil
}
