package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

// ValidateFunc is a token validation function that returns claims for a given
// JWT string. Used to extend the Verifier with additional validation methods
// (e.g., HS256 test tokens in development).
type ValidateFunc func(ctx context.Context, tokenString string) (*Claims, error)

// Verifier validates OIDC-issued JWTs using discovery and JWKS.
type Verifier struct {
	provider    *oidc.Provider
	verifier    *oidc.IDTokenVerifier
	fallback    ValidateFunc // optional fallback when JWKS validation fails
	orgResolver OrgResolver
}

// NewVerifier creates a JWT verifier by performing OIDC discovery against the issuer.
// The audience should match the API resource identifier registered in the OIDC provider.
func NewVerifier(ctx context.Context, issuerURL, audience string) (*Verifier, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC provider at %s: %w", issuerURL, err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: audience,
	})

	return &Verifier{
		provider: provider,
		verifier: verifier,
	}, nil
}

// WithFallback returns a copy of the verifier that tries the given ValidateFunc
// when JWKS validation fails. This is the primary path: conductor-minted session
// tokens are HS256 and always land here.
func (v *Verifier) WithFallback(fn ValidateFunc) *Verifier {
	return &Verifier{
		provider:    v.provider,
		verifier:    v.verifier,
		fallback:    fn,
		orgResolver: v.orgResolver,
	}
}

// OrgResolver maps the organization_id and organization scopes of a verified
// Logto organization access token to a workspace ID and role.
type OrgResolver func(ctx context.Context, organizationID, scope string) (workspace string, role WorkspaceRole, err error)

// WithOrgResolver returns a copy of the verifier that resolves organization
// tokens into workspace memberships via the given resolver.
func (v *Verifier) WithOrgResolver(fn OrgResolver) *Verifier {
	return &Verifier{
		provider:    v.provider,
		verifier:    v.verifier,
		fallback:    v.fallback,
		orgResolver: fn,
	}
}

// workspaceClaimEntry is the shape of workspace entries in the custom JWT claims.
type workspaceClaimEntry struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// ValidateToken validates a JWT and extracts claims.
// Tries JWKS validation first, falls back to the optional fallback function.
func (v *Verifier) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	// Try JWKS validation first (production path)
	if v.verifier != nil {
		idToken, err := v.verifier.Verify(ctx, tokenString)
		if err == nil {
			var rawClaims struct {
				Sub            string                `json:"sub"`
				Name           string                `json:"name,omitempty"`
				Email          string                `json:"email,omitempty"`
				Picture        string                `json:"picture,omitempty"`
				OrganizationID string                `json:"organization_id,omitempty"`
				Scope          string                `json:"scope,omitempty"`
				Workspaces     []workspaceClaimEntry `json:"workspaces,omitempty"`
			}
			if err := idToken.Claims(&rawClaims); err != nil {
				return nil, fmt.Errorf("failed to extract claims: %w", err)
			}
			if rawClaims.OrganizationID != "" && v.orgResolver != nil {
				workspace, role, resolveErr := v.orgResolver(ctx, rawClaims.OrganizationID, rawClaims.Scope)
				if resolveErr != nil {
					return nil, fmt.Errorf("failed to resolve organization %q: %w", rawClaims.OrganizationID, resolveErr)
				}
				return &Claims{
					Subject:    rawClaims.Sub,
					Name:       rawClaims.Name,
					Email:      rawClaims.Email,
					AvatarURL:  rawClaims.Picture,
					Workspaces: []WorkspaceMembership{{Workspace: workspace, Role: role}},
				}, nil
			}
			return claimsFromRaw(rawClaims.Sub, rawClaims.Name, rawClaims.Email, rawClaims.Picture, rawClaims.Workspaces), nil
		}
		// If no fallback configured, return the JWKS error
		if v.fallback == nil {
			return nil, fmt.Errorf("failed to verify token: %w", err)
		}
	}

	// Try fallback validation
	if v.fallback != nil {
		return v.fallback(ctx, tokenString)
	}

	return nil, fmt.Errorf("no verification method available")
}

// ClaimsFromJSON builds Claims from a raw JSON map. Useful for custom token
// validation functions that parse tokens independently.
func ClaimsFromJSON(sub, name, email, picture string, workspaces []WorkspaceClaim) *Claims {
	entries := make([]workspaceClaimEntry, len(workspaces))
	for i, ws := range workspaces {
		entries[i] = workspaceClaimEntry{ID: ws.ID, Role: ws.Role}
	}
	return claimsFromRaw(sub, name, email, picture, entries)
}

// WorkspaceClaim is the external representation of a workspace claim entry.
type WorkspaceClaim struct {
	ID   string
	Role string // "admin" or "user"
}

func claimsFromRaw(sub, name, email, picture string, wsClaims []workspaceClaimEntry) *Claims {
	workspaces := make([]WorkspaceMembership, 0, len(wsClaims))

	for _, ws := range wsClaims {
		role := WorkspaceRoleUser

		switch ws.Role {
		case "admin":
			role = WorkspaceRoleAdmin
		case "deployer":
			role = WorkspaceRoleDeployer
		}

		workspaces = append(workspaces, WorkspaceMembership{
			Workspace: ws.ID,
			Role:      role,
		})
	}

	return &Claims{
		Subject:    sub,
		Name:       name,
		Email:      email,
		AvatarURL:  picture,
		Workspaces: workspaces,
	}
}
