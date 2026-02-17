package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtClaims is the JWT claims structure stored in the token.
type jwtClaims struct {
	jwt.RegisteredClaims
	Email          string `json:"email,omitempty"`
	GitHubLogin    string `json:"github_login"`
	AvatarURL      string `json:"avatar_url"`
	Roles          []Role `json:"roles"`
	InstallationID int64  `json:"iid,omitempty"`
	GitHubToken    string `json:"ght,omitempty"`
}

// NewToken creates a signed JWT token from the given claims.
func NewToken(claims *Claims, secret string, expiry time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.Subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
		Email:          claims.Email,
		GitHubLogin:    claims.GitHubLogin,
		AvatarURL:      claims.AvatarURL,
		Roles:          claims.Roles,
		InstallationID: claims.InstallationID,
		GitHubToken:    claims.GitHubToken,
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

	return &Claims{
		Subject:        jc.Subject,
		Email:          jc.Email,
		GitHubLogin:    jc.GitHubLogin,
		AvatarURL:      jc.AvatarURL,
		Roles:          jc.Roles,
		InstallationID: jc.InstallationID,
		GitHubToken:    jc.GitHubToken,
	}, nil
}
