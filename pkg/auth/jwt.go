package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtWorkspaceMembership is the JWT-serializable form of WorkspaceMembership.
type jwtWorkspaceMembership struct {
	Workspace string        `json:"ws"`
	Role      WorkspaceRole `json:"role"`
}

// jwtClaims is the JWT claims structure stored in the token.
type jwtClaims struct {
	jwt.RegisteredClaims
	Email      string                   `json:"email,omitempty"`
	AvatarURL  string                   `json:"avatar_url"`
	Roles      []Role                   `json:"roles"`
	Workspaces []jwtWorkspaceMembership `json:"workspaces,omitempty"`
}

// NewToken creates a signed JWT token from the given claims.
func NewToken(claims *Claims, secret string, expiry time.Duration) (string, error) {
	now := time.Now()

	workspaces := make([]jwtWorkspaceMembership, len(claims.Workspaces))
	for i, m := range claims.Workspaces {
		workspaces[i] = jwtWorkspaceMembership{
			Workspace: m.Workspace,
			Role:      m.Role,
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.Subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
		Email:      claims.Email,
		AvatarURL:  claims.AvatarURL,
		Roles:      claims.Roles,
		Workspaces: workspaces,
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

// ParseToken validates and parses a JWT token string, returning the claims.
func ParseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	jc, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	workspaces := make([]WorkspaceMembership, len(jc.Workspaces))
	for i, m := range jc.Workspaces {
		workspaces[i] = WorkspaceMembership{
			Workspace: m.Workspace,
			Role:      m.Role,
		}
	}

	return &Claims{
		Subject:    jc.Subject,
		Email:      jc.Email,
		AvatarURL:  jc.AvatarURL,
		Roles:      jc.Roles,
		Workspaces: workspaces,
	}, nil
}
