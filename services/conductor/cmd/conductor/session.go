package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zeitlos/lucity/pkg/auth"
)

// TODO(stage-6b): delete this entire file (mintSessionToken, mintMachineToken,
// signSessionToken, sessionExpiry). HS256 session minting is replaced by
// IdP-issued OIDC bearer tokens. Its only callers are handleCallback/
// handleRefresh (oidc.go) and the CLI handoff (clihandoff.go), all also removed
// in stage-6b. Gate on the "legacy HS256 session token accepted" log in
// testauth.go going quiet in prod first.

const sessionExpiry = 7 * 24 * time.Hour // 7 days

// mintSessionToken creates an HMAC-SHA256 signed JWT containing the user's
// identity and workspace memberships. This is stored as the session cookie
// and verified by the auth middleware on every request.
func mintSessionToken(secret string, claims *auth.Claims) (string, error) {
	return signSessionToken(secret, claims, sessionExpiry)
}

func mintMachineToken(secret string, claims *auth.Claims, ttl time.Duration) (string, error) {
	return signSessionToken(secret, claims, ttl)
}

func signSessionToken(secret string, claims *auth.Claims, ttl time.Duration) (string, error) {
	workspaces := make([]map[string]string, len(claims.Workspaces))
	for i, ws := range claims.Workspaces {
		workspaces[i] = map[string]string{"id": ws.Workspace, "role": string(ws.Role)}
	}

	now := time.Now()
	mapClaims := jwt.MapClaims{
		"sub":        claims.Subject,
		"name":       claims.Name,
		"email":      claims.Email,
		"picture":    claims.AvatarURL,
		"workspaces": workspaces,
		"iat":        now.Unix(),
		"exp":        now.Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	return token.SignedString([]byte(secret))
}
