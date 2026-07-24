package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/logto"
)

// newOrgResolver resolves a Logto organization_id into a workspace ID (the
// organization name) and the caller's role (read from the token's organization
// scopes). Organization names are immutable identifiers, so they are cached
// permanently.
func newOrgResolver(client *logto.Client) auth.OrgResolver {
	var mu sync.Mutex
	names := map[string]string{}

	return func(ctx context.Context, organizationID, scope string) (string, auth.WorkspaceRole, error) {
		mu.Lock()
		name, ok := names[organizationID]
		mu.Unlock()

		if !ok {
			org, err := client.Organization(ctx, organizationID)
			if err != nil {
				return "", "", err
			}
			name = org.Name
			mu.Lock()
			names[organizationID] = name
			mu.Unlock()
		}

		role := highestRole(strings.Fields(scope))
		if role == "" {
			return "", "", fmt.Errorf("organization token for %q carries no workspace role scope", organizationID)
		}

		return name, role, nil
	}
}

// TODO: drop this role hierarchy once the @hasRole GraphQL directive accepts
// multiple roles again. Collapsing a token's roles to a single "highest" one via
// the deployer < user < admin ranking is too limiting; a directive that matches
// against a set of roles lets us model non-hierarchic roles directly.
func highestRole(scopes []string) auth.WorkspaceRole {
	best := auth.WorkspaceRole("")
	for _, s := range scopes {
		var candidate auth.WorkspaceRole
		switch s {
		case "admin":
			candidate = auth.WorkspaceRoleAdmin
		case "member":
			candidate = auth.WorkspaceRoleUser
		case "deployer":
			candidate = auth.WorkspaceRoleDeployer
		default:
			continue
		}
		if best == "" || candidate.Satisfies(best) {
			best = candidate
		}
	}
	return best
}
