package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zeitlos/lucity/pkg/auth"
)

// TODO(stage-6b): delete this entire file. hmacValidateFunc is the legacy HS256
// session-token fallback. Remove it together with the WithFallback wiring in
// main.go, SESSION_SECRET/AUTH_TEST_SECRET, and session.go — but only after the
// "legacy HS256 session token accepted" warning below has stopped appearing in
// prod logs (i.e. all clients have migrated to OIDC bearer tokens).
//
// hmacValidateFunc validates HS256 session tokens minted by the conductor
// (mintSessionToken). This is the production session-validation path, not
// test-only; AUTH_TEST_SECRET reuses it for test tokens.
func hmacValidateFunc(secret string) auth.ValidateFunc {
	return func(ctx context.Context, tokenString string) (*auth.Claims, error) {
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
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

		var workspaces []auth.WorkspaceClaim
		if ws, ok := mapClaims["workspaces"].([]interface{}); ok {
			for _, item := range ws {
				if m, ok := item.(map[string]interface{}); ok {
					id, _ := m["id"].(string)
					role, _ := m["role"].(string)
					workspaces = append(workspaces, auth.WorkspaceClaim{ID: id, Role: role})
				}
			}
		}

		// Reached only when JWKS/OIDC verification already failed and this HS256
		// token verified — i.e. a legacy client that hasn't migrated to bearer
		// tokens. This warning is the signal for when it's safe to delete HS256
		// (see TODO(stage-6b) above): once it stops appearing, no clients rely
		// on the fallback anymore.
		slog.WarnContext(ctx, "legacy HS256 session token accepted (fallback) — client not migrated to OIDC bearer tokens",
			"subject", sub, "legacy_auth", "hs256")

		return auth.ClaimsFromJSON(sub, name, email, picture, workspaces), nil
	}
}
