package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// ValidateFunc is a token validation function that returns claims for a given
// JWT string. Used to extend the Verifier with additional validation methods
// (e.g., HS256 test tokens in development).
type ValidateFunc func(ctx context.Context, tokenString string) (*Claims, error)

// Verifier validates OIDC-issued JWTs using discovery and JWKS.
type Verifier struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	fallback ValidateFunc // optional fallback when JWKS validation fails
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
// when JWKS validation fails. Useful for accepting test tokens in development.
func (v *Verifier) WithFallback(fn ValidateFunc) *Verifier {
	return &Verifier{
		provider: v.provider,
		verifier: v.verifier,
		fallback: fn,
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
				Sub        string                `json:"sub"`
				Name       string                `json:"name,omitempty"`
				Email      string                `json:"email,omitempty"`
				Picture    string                `json:"picture,omitempty"`
				Workspaces []workspaceClaimEntry `json:"workspaces,omitempty"`
			}
			if err := idToken.Claims(&rawClaims); err != nil {
				return nil, fmt.Errorf("failed to extract claims: %w", err)
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

// DecodeTokenClaims decodes the payload of a JWT without signature verification.
// Used when we already trust the token (e.g., just received from the token endpoint).
func DecodeTokenClaims(tokenString string) (*Claims, error) {
	parts := strings.SplitN(tokenString, ".", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var rawClaims struct {
		Sub        string                `json:"sub"`
		Name       string                `json:"name,omitempty"`
		Email      string                `json:"email,omitempty"`
		Picture    string                `json:"picture,omitempty"`
		Workspaces []workspaceClaimEntry `json:"workspaces,omitempty"`
	}
	if err := json.Unmarshal(payload, &rawClaims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT payload: %w", err)
	}

	return claimsFromRaw(rawClaims.Sub, rawClaims.Name, rawClaims.Email, rawClaims.Picture, rawClaims.Workspaces), nil
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
		if ws.Role == "admin" {
			role = WorkspaceRoleAdmin
		}
		workspaces = append(workspaces, WorkspaceMembership{
			Workspace: ws.ID,
			Role:      role,
		})
	}

	roles := []Role{RoleUser}
	for _, m := range workspaces {
		if m.Role == WorkspaceRoleAdmin {
			roles = append(roles, RoleAdmin)
			break
		}
	}

	return &Claims{
		Subject:    sub,
		Name:       name,
		Email:      email,
		AvatarURL:  picture,
		Roles:      roles,
		Workspaces: workspaces,
	}
}
