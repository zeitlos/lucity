package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zeitlos/lucity/pkg/auth"
)

const sessionExpiry = 7 * 24 * time.Hour // 7 days

// mintSessionToken creates an HMAC-SHA256 signed JWT containing the user's
// identity and workspace memberships. This is stored as the session cookie
// and verified by the auth middleware on every request.
func mintSessionToken(secret string, claims *auth.Claims) (string, error) {
	workspaces := make([]map[string]string, len(claims.Workspaces))
	for i, ws := range claims.Workspaces {
		workspaces[i] = map[string]string{"id": ws.Workspace, "role": string(ws.Role)}
	}

	mapClaims := jwt.MapClaims{
		"sub":        claims.Subject,
		"name":       claims.Name,
		"email":      claims.Email,
		"picture":    claims.AvatarURL,
		"workspaces": workspaces,
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(sessionExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	return token.SignedString([]byte(secret))
}
