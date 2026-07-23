package conductor

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/tenant"
)

const apiTokenKind = "api-token"

// APIToken is a workspace-scoped machine credential.
type APIToken struct {
	ID        string
	Name      string
	Role      auth.WorkspaceRole
	CreatedAt time.Time
	CreatedBy string
}

// CreatedAPIToken pairs a newly created token with its one-time secret string.
type CreatedAPIToken struct {
	APIToken APIToken
	Token    string
}

// CreateAPIToken provisions a machine identity in the active workspace, scoped to
// the given role, and returns the one-time token string.
func (c *Client) CreateAPIToken(ctx context.Context, name string, role auth.WorkspaceRole) (*CreatedAPIToken, error) {
	claims, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	workspace, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := c.orgID(ctx, workspace)
	if err != nil {
		return nil, err
	}
	roleID, err := c.apiTokenRoleID(ctx, role)
	if err != nil {
		return nil, err
	}

	app, err := c.logto.CreateApplication(ctx, name, "", map[string]any{
		"kind":      apiTokenKind,
		"workspace": workspace,
		"role":      string(role),
		"createdBy": claims.Subject,
	})
	if err != nil {
		return nil, err
	}

	if err := c.logto.AddApplicationToOrganization(ctx, orgID, app.ID); err != nil {
		_ = c.logto.DeleteApplication(ctx, app.ID)
		return nil, err
	}
	if err := c.logto.AssignApplicationOrganizationRoles(ctx, orgID, app.ID, []string{roleID}); err != nil {
		_ = c.logto.DeleteApplication(ctx, app.ID)
		return nil, err
	}
	secret, err := c.logto.CreateApplicationSecret(ctx, app.ID, "primary", time.Time{})
	if err != nil {
		_ = c.logto.DeleteApplication(ctx, app.ID)
		return nil, err
	}

	return &CreatedAPIToken{
		APIToken: APIToken{
			ID:        app.ID,
			Name:      app.Name,
			Role:      role,
			CreatedAt: time.UnixMilli(app.CreatedAt),
			CreatedBy: claims.Subject,
		},
		Token: encodeAPIToken(app.ID, secret.Value, orgID, workspace),
	}, nil
}

// APITokens lists the machine credentials in the active workspace.
func (c *Client) APITokens(ctx context.Context) ([]APIToken, error) {
	workspace, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := c.orgID(ctx, workspace)
	if err != nil {
		return nil, err
	}
	apps, err := c.logto.OrganizationApplications(ctx, orgID)
	if err != nil {
		return nil, err
	}

	tokens := make([]APIToken, 0)
	for _, app := range apps {
		if customDataString(app.CustomData, "kind") != apiTokenKind {
			continue
		}
		tokens = append(tokens, APIToken{
			ID:        app.ID,
			Name:      app.Name,
			Role:      apiTokenRole(app.CustomData),
			CreatedAt: time.UnixMilli(app.CreatedAt),
			CreatedBy: customDataString(app.CustomData, "createdBy"),
		})
	}
	return tokens, nil
}

// RevokeAPIToken deletes a machine credential. It only revokes tokens that belong
// to the active workspace, so a token ID from another tenant cannot be deleted.
func (c *Client) RevokeAPIToken(ctx context.Context, id string) error {
	workspace, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}
	orgID, err := c.orgID(ctx, workspace)
	if err != nil {
		return err
	}
	apps, err := c.logto.OrganizationApplications(ctx, orgID)
	if err != nil {
		return err
	}
	for _, app := range apps {
		if app.ID == id && customDataString(app.CustomData, "kind") == apiTokenKind {
			return c.logto.DeleteApplication(ctx, id)
		}
	}
	return fmt.Errorf("api token %q not found in workspace %q", id, workspace)
}

func (c *Client) apiTokenRoleID(ctx context.Context, role auth.WorkspaceRole) (string, error) {
	var name string
	switch role {
	case auth.WorkspaceRoleAdmin:
		name = "api-admin"
	case auth.WorkspaceRoleUser:
		name = "api-member"
	default:
		return "", fmt.Errorf("unsupported api token role %q", role)
	}
	roles, err := c.logto.OrganizationRoles(ctx)
	if err != nil {
		return "", err
	}
	for _, r := range roles {
		if r.Name == name {
			return r.ID, nil
		}
	}
	return "", fmt.Errorf("missing machine org role %q for api tokens", name)
}

func encodeAPIToken(clientID, secret, orgID, workspace string) string {
	return "lucity_" + base64.RawURLEncoding.EncodeToString([]byte(clientID+":"+secret+":"+orgID+":"+workspace))
}

func apiTokenRole(customData map[string]any) auth.WorkspaceRole {
	switch auth.WorkspaceRole(customDataString(customData, "role")) {
	case auth.WorkspaceRoleAdmin:
		return auth.WorkspaceRoleAdmin
	case auth.WorkspaceRoleDeployer:
		return auth.WorkspaceRoleDeployer
	default:
		return auth.WorkspaceRoleUser
	}
}

func customDataString(customData map[string]any, key string) string {
	if s, ok := customData[key].(string); ok {
		return s
	}
	return ""
}
