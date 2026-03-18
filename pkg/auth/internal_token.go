package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	internalIssuer = "lucity-gateway"
	tokenExpiry    = 30 * time.Second
)

// InternalClaims represents the claims extracted from an internal JWT.
type InternalClaims struct {
	Claims
	Workspace string // active workspace for this request
	IsSystem  bool   // true for webhook/system-initiated calls
	Scope     string // optional scope for system calls (e.g. "build", "deploy")
}

// internalJWTClaims is the JWT claims structure for internal tokens.
type internalJWTClaims struct {
	jwt.RegisteredClaims
	Email     string `json:"email,omitempty"`
	Roles     string `json:"roles,omitempty"`
	Workspace string `json:"ws,omitempty"`
	IsSystem  bool   `json:"sys,omitempty"`
	Scope     string `json:"scope,omitempty"`
}

// Issuer mints ES256 internal JWTs. Only the gateway and webhook hold the private key.
type Issuer struct {
	key    *ecdsa.PrivateKey
	expiry time.Duration
}

// NewIssuer creates an Issuer from a PEM-encoded EC private key.
func NewIssuer(pemKey []byte) (*Issuer, error) {
	key, err := parseECPrivateKey(pemKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse internal JWT private key: %w", err)
	}
	return &Issuer{key: key, expiry: tokenExpiry}, nil
}

// NewIssuerFromFile reads a PEM file and creates an Issuer.
func NewIssuerFromFile(path string) (*Issuer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read internal JWT private key file: %w", err)
	}
	return NewIssuer(data)
}

// MintToken creates a short-lived ES256 JWT for a user-initiated gRPC call.
func (iss *Issuer) MintToken(claims *Claims, workspace string) (string, error) {
	now := time.Now()
	jwtClaims := internalJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    internalIssuer,
			Subject:   claims.Subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(iss.expiry)),
		},
		Email:     claims.Email,
		Roles:     rolesToString(claims.Roles),
		Workspace: workspace,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwtClaims)
	return token.SignedString(iss.key)
}

// MintSystemToken creates a short-lived ES256 JWT for system-initiated calls (e.g. webhook).
func (iss *Issuer) MintSystemToken(subject, workspace, scope string) (string, error) {
	now := time.Now()
	jwtClaims := internalJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    internalIssuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(iss.expiry)),
		},
		Roles:     string(RoleUser),
		Workspace: workspace,
		IsSystem:  true,
		Scope:     scope,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwtClaims)
	return token.SignedString(iss.key)
}

// InternalVerifier validates ES256 internal JWTs. All backend services hold the public key.
type InternalVerifier struct {
	key *ecdsa.PublicKey
}

// NewInternalVerifier creates a verifier from a PEM-encoded EC public key.
func NewInternalVerifier(pemKey []byte) (*InternalVerifier, error) {
	key, err := parseECPublicKey(pemKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse internal JWT public key: %w", err)
	}
	return &InternalVerifier{key: key}, nil
}

// NewInternalVerifierFromFile reads a PEM file and creates an InternalVerifier.
func NewInternalVerifierFromFile(path string) (*InternalVerifier, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read internal JWT public key file: %w", err)
	}
	return NewInternalVerifier(data)
}

// Validate verifies an internal JWT signature and extracts claims.
func (v *InternalVerifier) Validate(tokenString string) (*InternalClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &internalJWTClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid internal token: %w", err)
	}

	jwtClaims, ok := token.Claims.(*internalJWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid internal token claims")
	}

	if jwtClaims.Issuer != internalIssuer {
		return nil, fmt.Errorf("invalid issuer: %s", jwtClaims.Issuer)
	}

	return &InternalClaims{
		Claims: Claims{
			Subject: jwtClaims.Subject,
			Email:   jwtClaims.Email,
			Roles:   parseRoles(jwtClaims.Roles),
		},
		Workspace: jwtClaims.Workspace,
		IsSystem:  jwtClaims.IsSystem,
		Scope:     jwtClaims.Scope,
	}, nil
}

// Context helpers for the Issuer

type issuerContextKey struct{}

// WithIssuer attaches an Issuer to the context.
func WithIssuer(ctx context.Context, iss *Issuer) context.Context {
	return context.WithValue(ctx, issuerContextKey{}, iss)
}

// IssuerFrom extracts the Issuer from the context.
func IssuerFrom(ctx context.Context) *Issuer {
	iss, _ := ctx.Value(issuerContextKey{}).(*Issuer)
	return iss
}

// Context helpers for active workspace (avoids circular import with pkg/tenant)

type activeWorkspaceContextKey struct{}

// WithActiveWorkspace attaches the active workspace to the auth context.
// Used by the gateway to propagate workspace into JWT minting.
func WithActiveWorkspace(ctx context.Context, ws string) context.Context {
	return context.WithValue(ctx, activeWorkspaceContextKey{}, ws)
}

// ActiveWorkspaceFrom extracts the active workspace from the auth context.
func ActiveWorkspaceFrom(ctx context.Context) string {
	ws, _ := ctx.Value(activeWorkspaceContextKey{}).(string)
	return ws
}

// Helper functions

func rolesToString(roles []Role) string {
	strs := make([]string, len(roles))
	for i, r := range roles {
		strs[i] = string(r)
	}
	return strings.Join(strs, ",")
}

func parseRoles(rolesStr string) []Role {
	if rolesStr == "" {
		return nil
	}
	parts := strings.Split(rolesStr, ",")
	roles := make([]Role, len(parts))
	for i, r := range parts {
		roles[i] = Role(r)
	}
	return roles
}

func parseECPrivateKey(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	// Try PKCS8 first (openssl genpkey output), then EC key format (openssl ecparam output)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		ecKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key is not ECDSA")
		}
		return ecKey, nil
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

func parseECPublicKey(pemData []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ecKey, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not ECDSA")
	}
	return ecKey, nil
}
