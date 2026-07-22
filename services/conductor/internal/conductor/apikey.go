package conductor

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/tenant"
)

const apiKeyKind = "api-key"

// APIKey is a workspace-scoped machine credential.
type APIKey struct {
	ID        string
	Name      string
	Role      auth.WorkspaceRole
	CreatedAt time.Time
	CreatedBy string
}

// CreatedAPIKey pairs a newly created key with its one-time secret string.
type CreatedAPIKey struct {
	APIKey APIKey
	Key    string
}

// CreateAPIKey provisions a machine identity in the active workspace, scoped to
// the given role, and returns the one-time key string.
func (c *Client) CreateAPIKey(ctx context.Context, name string, role auth.WorkspaceRole) (*CreatedAPIKey, error) {
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
	roleID, err := c.apiKeyRoleID(ctx, role)
	if err != nil {
		return nil, err
	}

	app, err := c.logto.CreateApplication(ctx, name, "", map[string]any{
		"kind":      apiKeyKind,
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

	return &CreatedAPIKey{
		APIKey: APIKey{
			ID:        app.ID,
			Name:      app.Name,
			Role:      role,
			CreatedAt: time.UnixMilli(app.CreatedAt),
			CreatedBy: claims.Subject,
		},
		Key: encodeAPIKey(app.ID, secret.Value, orgID, workspace),
	}, nil
}

// APIKeys lists the machine credentials in the active workspace.
func (c *Client) APIKeys(ctx context.Context) ([]APIKey, error) {
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

	keys := make([]APIKey, 0)
	for _, app := range apps {
		if apiKeyString(app.CustomData, "kind") != apiKeyKind {
			continue
		}
		keys = append(keys, APIKey{
			ID:        app.ID,
			Name:      app.Name,
			Role:      apiKeyRole(app.CustomData),
			CreatedAt: time.UnixMilli(app.CreatedAt),
			CreatedBy: apiKeyString(app.CustomData, "createdBy"),
		})
	}
	return keys, nil
}

// RevokeAPIKey deletes a machine credential. It only revokes keys that belong to
// the active workspace, so a key ID from another tenant cannot be deleted.
func (c *Client) RevokeAPIKey(ctx context.Context, id string) error {
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
		if app.ID == id && apiKeyString(app.CustomData, "kind") == apiKeyKind {
			return c.logto.DeleteApplication(ctx, id)
		}
	}
	return fmt.Errorf("api key %q not found in workspace %q", id, workspace)
}

func (c *Client) apiKeyRoleID(ctx context.Context, role auth.WorkspaceRole) (string, error) {
	var name string
	switch role {
	case auth.WorkspaceRoleAdmin:
		name = "api-admin"
	case auth.WorkspaceRoleUser:
		name = "api-member"
	default:
		return "", fmt.Errorf("unsupported api key role %q", role)
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
	return "", fmt.Errorf("missing machine org role %q for api keys", name)
}

func encodeAPIKey(clientID, secret, orgID, workspace string) string {
	return "lucity_" + base64.RawURLEncoding.EncodeToString([]byte(clientID+":"+secret+":"+orgID+":"+workspace))
}

func apiKeyRole(customData map[string]any) auth.WorkspaceRole {
	switch auth.WorkspaceRole(apiKeyString(customData, "role")) {
	case auth.WorkspaceRoleAdmin:
		return auth.WorkspaceRoleAdmin
	case auth.WorkspaceRoleDeployer:
		return auth.WorkspaceRoleDeployer
	default:
		return auth.WorkspaceRoleUser
	}
}

func apiKeyString(customData map[string]any, key string) string {
	if s, ok := customData[key].(string); ok {
		return s
	}
	return ""
}
