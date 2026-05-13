package logto

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/services/conductor/internal/directory"
)

var workspaceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

func (c *Client) Workspaces(ctx context.Context) ([]directory.Workspace, error) {
	orgs, err := c.api.Organizations(ctx)

	if err != nil {
		return nil, err
	}

	return toWorkspaces(orgs), nil
}

func (c *Client) WorkspacesForUser(ctx context.Context, userID string) ([]directory.Workspace, error) {
	orgs, err := c.api.UserOrganizations(ctx, userID)

	if err != nil {
		return nil, err
	}

	return toWorkspaces(orgs), nil
}

func (c *Client) Workspace(ctx context.Context, id string) (*directory.WorkspaceDetails, error) {
	orgID, err := c.orgID(ctx, id)

	if err != nil {
		return nil, err
	}

	org, err := c.api.Organization(ctx, orgID)

	if err != nil {
		return nil, err
	}

	members, err := c.api.OrganizationMembers(ctx, orgID)

	if err != nil {
		return nil, err
	}

	workspace := toWorkspaceDetails(*org, members)

	return &workspace, nil
}

func (c *Client) CreateWorkspace(ctx context.Context, id, name string, metadata map[string]any) (*directory.WorkspaceDetails, error) {
	if !workspaceIDPattern.MatchString(id) {
		return nil, fmt.Errorf("invalid workspace ID: must be 3-63 lowercase alphanumeric characters or hyphens")
	}

	if _, err := c.api.OrganizationByName(ctx, id); err == nil {
		return nil, fmt.Errorf("workspace ID %q is already taken", id)
	}

	org, err := c.api.CreateOrganization(ctx, id, name, metadata)

	if err != nil {
		return nil, err
	}

	c.cacheOrgID(id, org.ID)

	claims, err := auth.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	if err := c.api.AddOrganizationMember(ctx, org.ID, claims.Subject); err != nil {
		return nil, err
	}

	if err := c.api.AssignOrganizationRoles(ctx, org.ID, claims.Subject, []string{c.adminRoleID, c.memberRoleID}); err != nil {
		return nil, err
	}

	members, err := c.api.OrganizationMembers(ctx, org.ID)

	if err != nil {
		return nil, err
	}

	workspace := toWorkspaceDetails(*org, members)

	return &workspace, nil
}

func (c *Client) UpdateWorkspace(ctx context.Context, id, name string) (*directory.WorkspaceDetails, error) {
	orgID, err := c.orgID(ctx, id)

	if err != nil {
		return nil, err
	}

	org, err := c.api.Organization(ctx, orgID)

	if err != nil {
		return nil, err
	}

	customData := org.CustomData

	if id != name {
		customData["displayName"] = name
	} else {
		delete(customData, "displayName")
	}

	org, err = c.api.UpdateOrganizationCustomData(ctx, orgID, customData)

	if err != nil {
		return nil, err
	}

	members, err := c.api.OrganizationMembers(ctx, org.ID)

	if err != nil {
		return nil, err
	}

	workspace := toWorkspaceDetails(*org, members)

	return &workspace, nil
}

func (c *Client) DeleteWorkspace(ctx context.Context, id string) error {
	orgID, err := c.orgID(ctx, id)

	if err != nil {
		return err
	}

	org, err := c.api.Organization(ctx, orgID)

	if err != nil {
		return err
	}

	if personal, _ := org.CustomData["personal"].(bool); personal {
		return errors.New("cannot delete personal workspace")
	}

	if err := c.api.DeleteOrganization(ctx, id); err != nil {
		return err
	}

	c.invalidateOrgID(id)

	return nil
}

func (c *Client) InviteMember(ctx context.Context, workspaceID, email string, role directory.Role) (*directory.WorkspaceMember, error) {
	orgID, err := c.orgID(ctx, workspaceID)

	if err != nil {
		return nil, err
	}

	user, err := c.api.UserByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user with email %q not found in identity provider", email)
	}

	if err := c.api.AddOrganizationMember(ctx, orgID, user.ID); err != nil {
		return nil, err
	}

	roleIDs := []string{c.memberRoleID}

	if role == auth.WorkspaceRoleAdmin {
		roleIDs = append(roleIDs, c.adminRoleID)
	}

	if err := c.api.AssignOrganizationRoles(ctx, orgID, user.ID, roleIDs); err != nil {
		return nil, err
	}

	return &directory.WorkspaceMember{
		ID:    user.ID,
		Email: user.PrimaryEmail,
		Name:  user.Name,
		Role:  role,
	}, nil
}

func (c *Client) UpdateMemberRole(ctx context.Context, workspaceID, userID string, role directory.Role) (*directory.WorkspaceMember, error) {
	orgID, err := c.orgID(ctx, workspaceID)

	if err != nil {
		return nil, err
	}

	if role == auth.WorkspaceRoleAdmin {
		// Assign admin role
		if err := c.api.AssignOrganizationRoles(ctx, orgID, userID, []string{c.adminRoleID}); err != nil {
			return nil, err
		}
	} else {
		// Remove admin role
		if err := c.api.RemoveOrganizationRole(ctx, orgID, userID, c.adminRoleID); err != nil {
			return nil, fmt.Errorf("failed to remove admin role: %w", err)
		}
	}

	user, err := c.api.User(ctx, userID)

	if err != nil {
		return nil, err
	}

	return &directory.WorkspaceMember{
		ID:    user.ID,
		Email: user.PrimaryEmail,
		Name:  user.Name,
		Role:  role,
	}, nil
}

func (c *Client) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	orgID, err := c.orgID(ctx, workspaceID)

	if err != nil {
		return err
	}

	if err := c.api.RemoveOrganizationMember(ctx, orgID, userID); err != nil {
		return err
	}

	return nil
}

func (c *Client) orgID(ctx context.Context, workspaceID string) (string, error) {
	c.orgIDCacheMu.RLock()
	orgID, found := c.orgIDCache[workspaceID]
	c.orgIDCacheMu.RUnlock()

	if found {
		return orgID, nil
	}

	org, err := c.api.OrganizationByName(ctx, workspaceID)

	if err != nil {
		return "", err
	}

	c.cacheOrgID(workspaceID, org.ID)
	return org.ID, nil
}

func (c *Client) cacheOrgID(workspaceID, logtoOrgID string) {
	c.orgIDCacheMu.Lock()
	c.orgIDCache[workspaceID] = logtoOrgID
	c.orgIDCacheMu.Unlock()
}

func (c *Client) invalidateOrgID(workspaceID string) {
	c.orgIDCacheMu.Lock()
	delete(c.orgIDCache, workspaceID)
	c.orgIDCacheMu.Unlock()
}

func toWorkspace(org logto.Organization) directory.Workspace {
	workspace := directory.Workspace{
		ID:   org.Name,
		Name: org.Name,
	}

	if org.CustomData != nil {
		workspace.Personal, _ = org.CustomData["personal"].(bool)
		workspace.Suspended, _ = org.CustomData["suspended"].(bool)

		if displayName, ok := org.CustomData["displayName"].(string); ok && displayName != "" {
			workspace.Name = displayName
		}
	}

	return workspace
}

func toWorkspaces(orgs []logto.Organization) []directory.Workspace {
	var workspaces []directory.Workspace

	for _, org := range orgs {
		workspaces = append(workspaces, toWorkspace(org))
	}

	return workspaces
}

func toWorkspaceDetails(org logto.Organization, members []logto.OrganizationMember) directory.WorkspaceDetails {
	workspace := directory.WorkspaceDetails{
		Workspace: toWorkspace(org),
		Members:   toWorkspaceMembers(members),
	}

	return workspace
}

func toWorkspaceMember(member logto.OrganizationMember) directory.WorkspaceMember {
	result := directory.WorkspaceMember{
		ID:    member.ID,
		Email: member.Email,
		Name:  member.Name,
		Role:  auth.WorkspaceRoleUser,
	}

	for _, r := range member.OrgRoles {
		if r.Name == "admin" {
			result.Role = auth.WorkspaceRoleAdmin
			break
		}
	}

	return result
}

func toWorkspaceMembers(members []logto.OrganizationMember) []directory.WorkspaceMember {
	var result []directory.WorkspaceMember

	for _, member := range members {
		result = append(result, toWorkspaceMember(member))
	}

	return result
}
