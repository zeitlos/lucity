package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/logto"
)

const orgResolverTTL = time.Minute

// newOrgResolver resolves a Logto organization_id into a workspace ID (the
// organization name) and the caller's role (read from the token's organization
// scopes). Tokens of a deleted workspace are rejected. A live organization is
// looked up again after orgResolverTTL, so a deletion takes effect within it.
func newOrgResolver(client *logto.Client) auth.OrgResolver {
	type entry struct {
		name      string
		deleted   bool
		checkedAt time.Time
	}

	var mu sync.Mutex
	entries := map[string]entry{}

	return func(ctx context.Context, organizationID, scope string) (string, auth.WorkspaceRole, error) {
		mu.Lock()
		cached, ok := entries[organizationID]
		mu.Unlock()

		if !ok || (!cached.deleted && time.Since(cached.checkedAt) > orgResolverTTL) {
			org, err := client.Organization(ctx, organizationID)
			if err != nil && !ok {
				return "", "", fmt.Errorf("%w: %w", auth.ErrOrganizationLookup, err)
			}
			if err == nil {
				cached = entry{name: org.Name, deleted: org.Deleted(), checkedAt: time.Now()}
				mu.Lock()
				entries[organizationID] = cached
				mu.Unlock()
			}
		}

		if cached.deleted {
			return "", "", fmt.Errorf("organization token for deleted workspace %q", cached.name)
		}

		role := highestRole(strings.Fields(scope))
		if role == "" {
			return "", "", fmt.Errorf("organization token for %q carries no workspace role scope", organizationID)
		}

		return cached.name, role, nil
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
