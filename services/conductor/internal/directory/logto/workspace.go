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

func (p *Provider) Workspaces(ctx context.Context) ([]directory.Workspace, error) {
	orgs, err := p.api.Organizations(ctx)

	if err != nil {
		return nil, err
	}

	return toWorkspaces(orgs), nil
}

func (p *Provider) WorkspacesForUser(ctx context.Context, userID string) ([]directory.Workspace, error) {
	orgs, err := p.api.UserOrganizations(ctx, userID)

	if err != nil {
		return nil, err
	}

	return toWorkspaces(orgs), nil
}

func (p *Provider) Workspace(ctx context.Context, id string) (*directory.Workspace, error) {
	orgID, err := p.orgID(ctx, id)

	if err != nil {
		return nil, err
	}

	org, err := p.api.Organization(ctx, orgID)

	if err != nil {
		return nil, err
	}

	members, err := p.api.OrganizationMembers(ctx, orgID)

	if err != nil {
		return nil, err
	}

	workspace := toWorkspace(*org)
	workspace.Members = toWorkspaceMembers(members)

	return &workspace, nil
}

func (p *Provider) CreateWorkspace(ctx context.Context, id, name string) (*directory.Workspace, error) {
	if !workspaceIDPattern.MatchString(id) {
		return nil, fmt.Errorf("invalid workspace ID: must be 3-63 lowercase alphanumeric characters or hyphens")
	}

	if _, err := p.api.OrganizationByName(ctx, id); err == nil {
		return nil, fmt.Errorf("workspace ID %q is already taken", id)
	}

	org, err := p.api.CreateOrganization(ctx, id, name, nil)

	if err != nil {
		return nil, err
	}

	p.cacheOrgID(id, org.ID)

	claims, err := auth.FromContext(ctx)

	if err != nil {
		return nil, err
	}

	if err := p.api.AddOrganizationMember(ctx, org.ID, claims.Subject); err != nil {
		return nil, err
	}

	if err := p.api.AssignOrganizationRoles(ctx, org.ID, claims.Subject, []string{p.adminRoleID, p.memberRoleID}); err != nil {
		return nil, err
	}

	members, err := p.api.OrganizationMembers(ctx, org.ID)

	if err != nil {
		return nil, err
	}

	workspace := toWorkspace(*org)
	workspace.Members = toWorkspaceMembers(members)

	return &workspace, nil
}

func (p *Provider) UpdateWorkspace(ctx context.Context, id, name string) (*directory.Workspace, error) {
	orgID, err := p.orgID(ctx, id)

	if err != nil {
		return nil, err
	}

	org, err := p.api.Organization(ctx, orgID)

	if err != nil {
		return nil, err
	}

	customData := org.CustomData

	if id != name {
		customData["displayName"] = name
	} else {
		delete(customData, "displayName")
	}

	org, err = p.api.UpdateOrganizationCustomData(ctx, orgID, customData)

	if err != nil {
		return nil, err
	}

	members, err := p.api.OrganizationMembers(ctx, org.ID)

	if err != nil {
		return nil, err
	}

	workspace := toWorkspace(*org)
	workspace.Members = toWorkspaceMembers(members)

	return &workspace, nil
}

func (p *Provider) DeleteWorkspace(ctx context.Context, id string) error {
	orgID, err := p.orgID(ctx, id)

	if err != nil {
		return err
	}

	org, err := p.api.Organization(ctx, orgID)

	if err != nil {
		return err
	}

	if personal, _ := org.CustomData["personal"].(bool); personal {
		return errors.New("cannot delete personal workspace")
	}

	if err := p.api.DeleteOrganization(ctx, id); err != nil {
		return err
	}

	p.invalidateOrgID(id)

	return nil
}

func (p *Provider) InviteMember(ctx context.Context, workspaceID, email string, role directory.Role) (*directory.WorkspaceMember, error) {
	orgID, err := p.orgID(ctx, workspaceID)

	if err != nil {
		return nil, err
	}

	user, err := p.api.UserByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user with email %q not found in identity provider", email)
	}

	if err := p.api.AddOrganizationMember(ctx, orgID, user.ID); err != nil {
		return nil, err
	}

	roleIDs := []string{p.memberRoleID}

	if role == auth.WorkspaceRoleAdmin {
		roleIDs = append(roleIDs, p.adminRoleID)
	}

	if err := p.api.AssignOrganizationRoles(ctx, orgID, user.ID, roleIDs); err != nil {
		return nil, err
	}

	return &directory.WorkspaceMember{
		ID:    user.ID,
		Email: user.PrimaryEmail,
		Name:  user.Name,
		Role:  role,
	}, nil
}

func (p *Provider) UpdateMemberRole(ctx context.Context, workspaceID, userID string, role directory.Role) (*directory.WorkspaceMember, error) {
	orgID, err := p.orgID(ctx, workspaceID)

	if err != nil {
		return nil, err
	}

	if role == auth.WorkspaceRoleAdmin {
		// Assign admin role
		if err := p.api.AssignOrganizationRoles(ctx, orgID, userID, []string{p.adminRoleID}); err != nil {
			return nil, err
		}
	} else {
		// Remove admin role
		if err := p.api.RemoveOrganizationRole(ctx, orgID, userID, p.adminRoleID); err != nil {
			return nil, fmt.Errorf("failed to remove admin role: %w", err)
		}
	}

	user, err := p.api.User(ctx, userID)

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

func (p *Provider) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	orgID, err := p.orgID(ctx, workspaceID)

	if err != nil {
		return err
	}

	if err := p.api.RemoveOrganizationMember(ctx, orgID, userID); err != nil {
		return err
	}

	return nil
}

func (p *Provider) orgID(ctx context.Context, workspaceID string) (string, error) {
	p.orgIDCacheMu.RLock()
	orgID, found := p.orgIDCache[workspaceID]
	p.orgIDCacheMu.RUnlock()

	if found {
		return orgID, nil
	}

	org, err := p.api.OrganizationByName(ctx, workspaceID)

	if err != nil {
		return "", err
	}

	p.cacheOrgID(workspaceID, org.ID)
	return org.ID, nil
}

func (p *Provider) cacheOrgID(workspaceID, logtoOrgID string) {
	p.orgIDCacheMu.Lock()
	p.orgIDCache[workspaceID] = logtoOrgID
	p.orgIDCacheMu.Unlock()
}

func (p *Provider) invalidateOrgID(workspaceID string) {
	p.orgIDCacheMu.Lock()
	delete(p.orgIDCache, workspaceID)
	p.orgIDCacheMu.Unlock()
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
