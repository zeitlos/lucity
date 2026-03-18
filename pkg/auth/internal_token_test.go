package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestKeypair(t *testing.T) ([]byte, []byte) {
	t.Helper()

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	privBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return privPEM, pubPEM
}

func TestMintAndValidateToken(t *testing.T) {
	privPEM, pubPEM := generateTestKeypair(t)

	issuer, err := NewIssuer(privPEM)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewInternalVerifier(pubPEM)
	if err != nil {
		t.Fatal(err)
	}

	claims := &Claims{
		Subject: "user-123",
		Email:   "user@example.com",
		Roles:   []Role{RoleUser, RoleAdmin},
	}

	tokenStr, err := issuer.MintToken(claims, "my-workspace")
	if err != nil {
		t.Fatal(err)
	}

	got, err := verifier.Validate(tokenStr)
	if err != nil {
		t.Fatal(err)
	}

	if got.Subject != "user-123" {
		t.Errorf("subject = %q, want %q", got.Subject, "user-123")
	}
	if got.Email != "user@example.com" {
		t.Errorf("email = %q, want %q", got.Email, "user@example.com")
	}
	if got.Workspace != "my-workspace" {
		t.Errorf("workspace = %q, want %q", got.Workspace, "my-workspace")
	}
	if got.IsSystem {
		t.Error("IsSystem = true, want false")
	}
	if len(got.Roles) != 2 || got.Roles[0] != RoleUser || got.Roles[1] != RoleAdmin {
		t.Errorf("roles = %v, want [USER, ADMIN]", got.Roles)
	}
}

func TestMintSystemToken(t *testing.T) {
	privPEM, pubPEM := generateTestKeypair(t)

	issuer, err := NewIssuer(privPEM)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewInternalVerifier(pubPEM)
	if err != nil {
		t.Fatal(err)
	}

	tokenStr, err := issuer.MintSystemToken("webhook", "my-workspace", "build")
	if err != nil {
		t.Fatal(err)
	}

	got, err := verifier.Validate(tokenStr)
	if err != nil {
		t.Fatal(err)
	}

	if got.Subject != "webhook" {
		t.Errorf("subject = %q, want %q", got.Subject, "webhook")
	}
	if got.Workspace != "my-workspace" {
		t.Errorf("workspace = %q, want %q", got.Workspace, "my-workspace")
	}
	if !got.IsSystem {
		t.Error("IsSystem = false, want true")
	}
	if got.Scope != "build" {
		t.Errorf("scope = %q, want %q", got.Scope, "build")
	}
}

func TestValidateWithWrongKey(t *testing.T) {
	privPEM, _ := generateTestKeypair(t)
	_, otherPubPEM := generateTestKeypair(t)

	issuer, err := NewIssuer(privPEM)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewInternalVerifier(otherPubPEM)
	if err != nil {
		t.Fatal(err)
	}

	tokenStr, err := issuer.MintToken(&Claims{Subject: "user"}, "ws")
	if err != nil {
		t.Fatal(err)
	}

	_, err = verifier.Validate(tokenStr)
	if err == nil {
		t.Error("expected error validating with wrong key, got nil")
	}
}

func TestValidateExpiredToken(t *testing.T) {
	privPEM, pubPEM := generateTestKeypair(t)

	issuer, err := NewIssuer(privPEM)
	if err != nil {
		t.Fatal(err)
	}
	issuer.expiry = -1 * time.Second // already expired

	verifier, err := NewInternalVerifier(pubPEM)
	if err != nil {
		t.Fatal(err)
	}

	tokenStr, err := issuer.MintToken(&Claims{Subject: "user"}, "ws")
	if err != nil {
		t.Fatal(err)
	}

	_, err = verifier.Validate(tokenStr)
	if err == nil {
		t.Error("expected error validating expired token, got nil")
	}
}

func TestValidateMalformedToken(t *testing.T) {
	_, pubPEM := generateTestKeypair(t)

	verifier, err := NewInternalVerifier(pubPEM)
	if err != nil {
		t.Fatal(err)
	}

	_, err = verifier.Validate("not.a.jwt")
	if err == nil {
		t.Error("expected error validating malformed token, got nil")
	}
}

func TestValidateWrongIssuer(t *testing.T) {
	privPEM, pubPEM := generateTestKeypair(t)

	privKey, _ := parseECPrivateKey(privPEM)

	// Mint a token with wrong issuer
	claims := internalJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wrong-issuer",
			Subject:   "user",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenStr, err := token.SignedString(privKey)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewInternalVerifier(pubPEM)
	if err != nil {
		t.Fatal(err)
	}

	_, err = verifier.Validate(tokenStr)
	if err == nil {
		t.Error("expected error for wrong issuer, got nil")
	}
}

func TestEmptyWorkspace(t *testing.T) {
	privPEM, pubPEM := generateTestKeypair(t)

	issuer, err := NewIssuer(privPEM)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewInternalVerifier(pubPEM)
	if err != nil {
		t.Fatal(err)
	}

	// Empty workspace is valid (some queries don't need workspace context)
	tokenStr, err := issuer.MintToken(&Claims{Subject: "user"}, "")
	if err != nil {
		t.Fatal(err)
	}

	got, err := verifier.Validate(tokenStr)
	if err != nil {
		t.Fatal(err)
	}

	if got.Workspace != "" {
		t.Errorf("workspace = %q, want empty", got.Workspace)
	}
}
