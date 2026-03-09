package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/zeitlos/lucity/pkg/auth"
)

const (
	stateCookieName    = "lucity_oauth_state"
	verifierCookieName = "lucity_pkce_verifier"
	tokenCookieName    = "lucity_token"
	tokenExpiry        = 7 * 24 * time.Hour
)

// OIDCProvider wraps the OIDC discovery provider, ID token verifier, and OAuth2 config.
type OIDCProvider struct {
	provider    *oidc.Provider
	verifier    *oidc.IDTokenVerifier
	oauthConfig oauth2.Config
}

// NewOIDCProvider performs OIDC discovery against the issuer and returns a configured provider.
// Uses PKCE (S256) — no client secret needed. The client must be configured as "public" in the IDP.
func NewOIDCProvider(ctx context.Context, issuerURL, clientID, callbackURL string) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC provider at %s: %w", issuerURL, err)
	}

	oauthConfig := oauth2.Config{
		ClientID:    clientID,
		Endpoint:    provider.Endpoint(),
		RedirectURL: callbackURL,
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCProvider{
		provider:    provider,
		verifier:    verifier,
		oauthConfig: oauthConfig,
	}, nil
}

// registerAuthRoutes adds OIDC auth endpoints to the mux.
func registerAuthRoutes(mux *http.ServeMux, provider *OIDCProvider, jwtSecret, dashboardURL string) {
	mux.HandleFunc("/auth/login", handleLogin(provider))
	mux.HandleFunc("/auth/callback", handleCallback(provider, jwtSecret, dashboardURL))
	mux.HandleFunc("/auth/me", handleMe())
	mux.HandleFunc("/auth/logout", handleLogout(dashboardURL))
}

// handleLogin redirects to the OIDC provider's authorization page with PKCE.
func handleLogin(provider *OIDCProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := generateState()
		verifier := generateCodeVerifier()
		challenge := codeChallenge(verifier)

		http.SetCookie(w, &http.Cookie{
			Name:     stateCookieName,
			Value:    state,
			Path:     "/",
			MaxAge:   600, // 10 minutes
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     verifierCookieName,
			Value:    verifier,
			Path:     "/",
			MaxAge:   600,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		url := provider.oauthConfig.AuthCodeURL(state,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// handleCallback exchanges the auth code for tokens, verifies the ID token,
// extracts claims, and creates a Lucity session.
func handleCallback(provider *OIDCProvider, jwtSecret, dashboardURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify state
		stateCookie, err := r.Cookie(stateCookieName)
		if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		// Clear state cookie
		http.SetCookie(w, &http.Cookie{
			Name:   stateCookieName,
			Path:   "/",
			MaxAge: -1,
		})

		// Retrieve PKCE verifier
		verifierCookie, err := r.Cookie(verifierCookieName)
		if err != nil {
			http.Error(w, "missing PKCE verifier", http.StatusBadRequest)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:   verifierCookieName,
			Path:   "/",
			MaxAge: -1,
		})

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		// Exchange code for OAuth2 token (with PKCE verifier)
		oauth2Token, err := provider.oauthConfig.Exchange(r.Context(), code,
			oauth2.SetAuthURLParam("code_verifier", verifierCookie.Value),
		)
		if err != nil {
			// Try to extract the OAuth2 error details
			if rErr, ok := err.(*oauth2.RetrieveError); ok {
				slog.Error("failed to exchange code", "error", err, "status", rErr.Response.StatusCode, "body", string(rErr.Body))
			} else {
				slog.Error("failed to exchange code", "error", err)
			}
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		// Extract and verify the ID token
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			slog.Error("no id_token in OAuth2 token response")
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		idToken, err := provider.verifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			slog.Error("failed to verify id token", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		// Extract claims from the ID token
		var oidcClaims struct {
			Email   string   `json:"email"`
			Name    string   `json:"name"`
			Picture string   `json:"picture"`
			Groups  []string `json:"groups"`
		}
		if err := idToken.Claims(&oidcClaims); err != nil {
			slog.Error("failed to extract claims", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		// Parse workspace memberships from Rauthy groups
		workspaces := auth.ParseRauthyGroups(oidcClaims.Groups)
		if len(workspaces) == 0 {
			slog.Warn("user has no workspace memberships", "email", oidcClaims.Email)
			http.Redirect(w, r, dashboardURL+"/login?error=no_workspace", http.StatusTemporaryRedirect)
			return
		}

		// Determine roles — admin if member of any workspace as admin
		roles := []auth.Role{auth.RoleUser}
		for _, m := range workspaces {
			if m.Role == auth.WorkspaceRoleAdmin {
				roles = append(roles, auth.RoleAdmin)
				break
			}
		}

		// Create Lucity JWT
		claims := &auth.Claims{
			Subject:    oidcClaims.Name,
			Email:      oidcClaims.Email,
			AvatarURL:  oidcClaims.Picture,
			Roles:      roles,
			Workspaces: workspaces,
		}

		jwt, err := auth.NewToken(claims, jwtSecret, tokenExpiry)
		if err != nil {
			slog.Error("failed to create token", "error", err)
			http.Error(w, "failed to create session", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     tokenCookieName,
			Value:    jwt,
			Path:     "/",
			MaxAge:   int(tokenExpiry.Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		slog.Info("user authenticated", "email", oidcClaims.Email, "workspaces", len(workspaces))
		http.Redirect(w, r, dashboardURL, http.StatusTemporaryRedirect)
	}
}

// handleMe returns the current user's profile from the JWT claims in context.
func handleMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		type workspaceEntry struct {
			Workspace string             `json:"workspace"`
			Role      auth.WorkspaceRole `json:"role"`
		}

		workspaces := make([]workspaceEntry, len(claims.Workspaces))
		for i, m := range claims.Workspaces {
			workspaces[i] = workspaceEntry{
				Workspace: m.Workspace,
				Role:      m.Role,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":       claims.Subject,
			"email":      claims.Email,
			"avatarUrl":  claims.AvatarURL,
			"workspaces": workspaces,
		})
	}
}

// handleLogout clears the session cookie.
func handleLogout(dashboardURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   tokenCookieName,
			Path:   "/",
			MaxAge: -1,
		})

		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}

		http.Redirect(w, r, dashboardURL+"/login", http.StatusTemporaryRedirect)
	}
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateCodeVerifier creates a random PKCE code verifier (43–128 chars, base64url).
func generateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// codeChallenge computes the S256 PKCE code challenge from a verifier.
func codeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
