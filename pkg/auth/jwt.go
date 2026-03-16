package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
)

// Verifier validates OIDC-issued JWTs using discovery and JWKS.
// Optionally accepts HS256 test tokens when a testSecret is configured.
type Verifier struct {
	provider   *oidc.Provider
	verifier   *oidc.IDTokenVerifier
	testSecret string // HS256 secret for dev/test tokens (empty = disabled)
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

// NewTestVerifier creates a verifier that only accepts HS256 test tokens.
// Use this for integration tests when no OIDC provider is available.
func NewTestVerifier(testSecret string) *Verifier {
	return &Verifier{testSecret: testSecret}
}

// WithTestSecret returns a copy of the verifier that also accepts HS256 test tokens.
// Intended for local development — never set AUTH_TEST_SECRET in production.
func (v *Verifier) WithTestSecret(secret string) *Verifier {
	return &Verifier{
		provider:   v.provider,
		verifier:   v.verifier,
		testSecret: secret,
	}
}

// workspaceClaimEntry is the shape of workspace entries in the custom JWT claims.
type workspaceClaimEntry struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// ValidateToken validates a JWT and extracts claims.
// Tries JWKS validation first, falls back to HS256 test secret if configured.
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
		// If no test secret configured, return the JWKS error
		if v.testSecret == "" {
			return nil, fmt.Errorf("failed to verify token: %w", err)
		}
	}

	// Fall back to HS256 test token validation
	if v.testSecret != "" {
		return v.validateTestToken(tokenString)
	}

	return nil, fmt.Errorf("no verification method available")
}

// validateTestToken validates an HS256 JWT signed with the test secret.
func (v *Verifier) validateTestToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(v.testSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to verify test token: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type")
	}

	sub, _ := mapClaims["sub"].(string)
	name, _ := mapClaims["name"].(string)
	email, _ := mapClaims["email"].(string)
	picture, _ := mapClaims["picture"].(string)

	var wsClaims []workspaceClaimEntry
	if ws, ok := mapClaims["workspaces"].([]interface{}); ok {
		for _, item := range ws {
			if m, ok := item.(map[string]interface{}); ok {
				id, _ := m["id"].(string)
				role, _ := m["role"].(string)
				wsClaims = append(wsClaims, workspaceClaimEntry{ID: id, Role: role})
			}
		}
	}

	return claimsFromRaw(sub, name, email, picture, wsClaims), nil
}

// NewTestToken creates an HS256 JWT for integration tests.
// The token embeds the given claims and expires after the given duration.
func NewTestToken(claims *Claims, secret string, expiry time.Duration) (string, error) {
	wsClaims := make([]workspaceClaimEntry, len(claims.Workspaces))
	for i, m := range claims.Workspaces {
		role := "user"
		if m.Role == WorkspaceRoleAdmin {
			role = "admin"
		}
		wsClaims[i] = workspaceClaimEntry{ID: m.Workspace, Role: role}
	}

	mapClaims := jwt.MapClaims{
		"sub":        claims.Subject,
		"name":       claims.Name,
		"email":      claims.Email,
		"picture":    claims.AvatarURL,
		"workspaces": wsClaims,
		"exp":        time.Now().Add(expiry).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	return token.SignedString([]byte(secret))
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
