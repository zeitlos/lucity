package main

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zeitlos/lucity/pkg/auth"
)

// hmacValidateFunc validates HS256 session tokens minted by the conductor
// (mintSessionToken). This is the production session-validation path, not
// test-only; AUTH_TEST_SECRET reuses it for test tokens.
func hmacValidateFunc(secret string) auth.ValidateFunc {
	return func(_ context.Context, tokenString string) (*auth.Claims, error) {
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

		return auth.ClaimsFromJSON(sub, name, email, picture, workspaces), nil
	}
}
