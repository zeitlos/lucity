package main

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zeitlos/lucity/pkg/auth"
)

// hmacValidateFunc returns an auth.ValidateFunc that validates HS256 JWTs
// signed with the given secret. Used only for integration tests when
// AUTH_TEST_SECRET is set. Never use in production.
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
