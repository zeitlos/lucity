package conductor

import (
	"context"
	"fmt"
	"time"

	"github.com/zeitlos/lucity/pkg/logto"
)

const ciAppKind = "ci"

// IssueCIDeployToken brokers a short-lived, deployer-scoped organization access
// token for a repository's CI, backed by a per-repo machine identity. Conductor
// does not sign the token — the identity provider issues it; conductor only
// validates the external assertion and drives the grant.
func (c *Client) IssueCIDeployToken(ctx context.Context, repo, workspace, audience string) (string, time.Time, error) {
	orgID, err := c.orgID(ctx, workspace)
	if err != nil {
		return "", time.Time{}, err
	}
	roleID, err := c.orgRoleIDByName(ctx, "deployer")
	if err != nil {
		return "", time.Time{}, err
	}
	app, err := c.ensureCIApplication(ctx, orgID, repo, workspace, roleID)
	if err != nil {
		return "", time.Time{}, err
	}

	secret, err := c.logto.CreateApplicationSecret(ctx, app.ID, "ci", time.Now().Add(2*time.Minute))
	if err != nil {
		return "", time.Time{}, err
	}

	token, expiresIn, err := c.logto.OrganizationAccessToken(ctx, app.ID, secret.Value, audience, orgID, []string{"deployer"})
	if err != nil {
		return "", time.Time{}, err
	}
	return token, time.Now().Add(time.Duration(expiresIn) * time.Second), nil
}

func (c *Client) ensureCIApplication(ctx context.Context, orgID, repo, workspace, roleID string) (*logto.Application, error) {
	apps, err := c.logto.OrganizationApplications(ctx, orgID)
	if err != nil {
		return nil, err
	}
	for i := range apps {
		if customDataString(apps[i].CustomData, "kind") == ciAppKind && customDataString(apps[i].CustomData, "repo") == repo {
			return &apps[i], nil
		}
	}

	app, err := c.logto.CreateApplication(ctx, "ci:"+repo, "", map[string]any{
		"kind":      ciAppKind,
		"repo":      repo,
		"workspace": workspace,
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
	return app, nil
}

// CIDeployRepo returns the repository a CI machine identity deploys, given its
// client id (the subject of a CI-issued token) within a workspace.
func (c *Client) CIDeployRepo(ctx context.Context, workspace, clientID string) (string, bool) {
	orgID, err := c.orgID(ctx, workspace)
	if err != nil {
		return "", false
	}
	apps, err := c.logto.OrganizationApplications(ctx, orgID)
	if err != nil {
		return "", false
	}
	for i := range apps {
		if apps[i].ID == clientID && customDataString(apps[i].CustomData, "kind") == ciAppKind {
			return customDataString(apps[i].CustomData, "repo"), true
		}
	}
	return "", false
}

func (c *Client) orgRoleIDByName(ctx context.Context, name string) (string, error) {
	roles, err := c.logto.OrganizationRoles(ctx)
	if err != nil {
		return "", err
	}
	for _, r := range roles {
		if r.Name == name {
			return r.ID, nil
		}
	}
	return "", fmt.Errorf("missing org role %q", name)
}
