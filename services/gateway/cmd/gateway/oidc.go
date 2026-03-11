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
	"github.com/zeitlos/lucity/services/gateway/handler"
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

const (
	githubStateCookieName = "lucity_github_state"
)

// registerAuthRoutes adds OIDC auth endpoints to the mux.
func registerAuthRoutes(mux *http.ServeMux, provider *OIDCProvider, api *handler.Client, jwtSecret, dashboardURL, githubAppSlug string) {
	mux.HandleFunc("/auth/login", handleLogin(provider))
	mux.HandleFunc("/auth/callback", handleCallback(provider, api, jwtSecret, dashboardURL))
	mux.HandleFunc("/auth/me", handleMe())
	mux.HandleFunc("/auth/logout", handleLogout(dashboardURL))
	mux.HandleFunc("/auth/refresh", handleRefresh(api, jwtSecret))
	mux.HandleFunc("/auth/github/install", handleGitHubInstall(githubAppSlug))
	mux.HandleFunc("/auth/github/setup", handleGitHubSetup(dashboardURL))
	mux.HandleFunc("/auth/github/connect", handleGitHubConnect(api))
	mux.HandleFunc("/auth/github/callback", handleGitHubCallback(api, dashboardURL))
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
func handleCallback(provider *OIDCProvider, api *handler.Client, jwtSecret, dashboardURL string) http.HandlerFunc {
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
			Email             string   `json:"email"`
			Name              string   `json:"name"`
			Picture           string   `json:"picture"`
			Groups            []string `json:"groups"`
			PreferredUsername  string   `json:"preferred_username"`
		}
		if err := idToken.Claims(&oidcClaims); err != nil {
			slog.Error("failed to extract claims", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		// Ensure personal workspace exists and user is a member.
		// This is idempotent — on first login it creates the workspace,
		// on subsequent logins it ensures the Rauthy group membership is intact.
		if oidcClaims.PreferredUsername == "" {
			slog.Warn("no preferred_username for personal workspace", "email", oidcClaims.Email)
			http.Redirect(w, r, dashboardURL+"/login?error=no_workspace", http.StatusTemporaryRedirect)
			return
		}

		// Set claims on context so auth.OutgoingContext() can propagate identity
		// to backend services via gRPC metadata. No JWT needed — backends trust
		// the gateway as the auth boundary.
		svcCtx := auth.WithClaims(r.Context(), &auth.Claims{
			Subject: idToken.Subject,
			Email:   oidcClaims.Email,
			Roles:   []auth.Role{auth.RoleUser},
		})

		personalWSID, isNewUser, err := api.EnsurePersonalWorkspace(svcCtx, idToken.Subject, oidcClaims.PreferredUsername)
		if err != nil {
			slog.Error("failed to ensure personal workspace", "error", err, "email", oidcClaims.Email)
			http.Error(w, "failed to create workspace", http.StatusInternalServerError)
			return
		}

		// Re-read the user's groups from Rauthy after ensuring personal workspace,
		// since EnsurePersonalWorkspace may have updated group membership.
		user, err := api.Rauthy.User(r.Context(), idToken.Subject)
		if err != nil {
			slog.Error("failed to fetch user after personal workspace setup", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}
		workspaces := auth.ParseRauthyGroups(user.Groups)

		slog.Info("personal workspace ensured", "email", oidcClaims.Email, "workspace", personalWSID)

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
			Subject:    idToken.Subject,
			Name:       oidcClaims.Name,
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

		redirectURL := dashboardURL
		if isNewUser {
			redirectURL = dashboardURL + "/?welcome=true"
		}
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
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
			"name":       claims.Name,
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

// handleRefresh re-reads the user's Rauthy groups and mints a new JWT.
// Called by the dashboard after workspace mutations (create, invite, etc.).
func handleRefresh(api *handler.Client, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := auth.FromContext(r.Context())
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if api.Rauthy == nil {
			http.Error(w, "token refresh not available", http.StatusServiceUnavailable)
			return
		}

		// Fetch current user from Rauthy to get updated group list.
		// Rauthy returns group names (e.g. "ws:myworkspace", "ws:myworkspace:admin"),
		// which ParseRauthyGroups can process directly.
		user, err := api.Rauthy.User(r.Context(), claims.Subject)
		if err != nil {
			slog.Error("failed to fetch user for token refresh", "error", err, "subject", claims.Subject)
			http.Error(w, "failed to refresh token", http.StatusInternalServerError)
			return
		}

		// Parse workspace memberships from group names
		workspaces := auth.ParseRauthyGroups(user.Groups)

		roles := []auth.Role{auth.RoleUser}
		for _, m := range workspaces {
			if m.Role == auth.WorkspaceRoleAdmin {
				roles = append(roles, auth.RoleAdmin)
				break
			}
		}

		newClaims := &auth.Claims{
			Subject:    claims.Subject,
			Name:       claims.Name,
			Email:      claims.Email,
			AvatarURL:  claims.AvatarURL,
			Roles:      roles,
			Workspaces: workspaces,
		}

		jwt, err := auth.NewToken(newClaims, jwtSecret, tokenExpiry)
		if err != nil {
			slog.Error("failed to create refreshed token", "error", err)
			http.Error(w, "failed to refresh token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     tokenCookieName,
			Value:    jwt,
			Path:     "/",
			MaxAge:   int(tokenExpiry.Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		slog.Info("token refreshed", "email", claims.Email, "workspaces", len(workspaces))
		w.WriteHeader(http.StatusOK)
	}
}

// handleGitHubInstall redirects to GitHub App installation page.
// Users can install the GitHub App on new accounts to make them available as sources.
func handleGitHubInstall(githubAppSlug string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if githubAppSlug == "" {
			http.Error(w, "GitHub App not configured", http.StatusServiceUnavailable)
			return
		}

		claims := auth.FromContext(r.Context())
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		installURL := fmt.Sprintf("https://github.com/apps/%s/installations/new", githubAppSlug)
		http.Redirect(w, r, installURL, http.StatusTemporaryRedirect)
	}
}

// handleGitHubSetup is the callback from GitHub after App installation.
// Returns an HTML page that signals the opener and closes the popup.
// Falls back to a redirect if not opened as a popup.
func handleGitHubSetup(dashboardURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html><html><head><title>GitHub App Installed</title></head><body><script>
if (window.opener) { window.opener.postMessage("github-app-installed", %q); window.close(); }
else { window.location.href = %q; }
</script><p>GitHub App installed. You can close this window.</p></body></html>`, dashboardURL, dashboardURL+"/?github=installed")
	}
}

// handleGitHubConnect initiates the GitHub OAuth flow to connect the user's GitHub account.
// The token is stored per-user and used for listing installations (githubSources query).
func handleGitHubConnect(api *handler.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if api.GitHubApp == nil {
			http.Error(w, "GitHub App not configured", http.StatusServiceUnavailable)
			return
		}

		state := generateState()
		http.SetCookie(w, &http.Cookie{
			Name:     githubStateCookieName,
			Value:    state,
			Path:     "/",
			MaxAge:   600, // 10 minutes
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		url := api.GitHubApp.OAuthURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// handleGitHubCallback exchanges the authorization code for a token and stores it.
func handleGitHubCallback(api *handler.Client, dashboardURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify state
		stateCookie, err := r.Cookie(githubStateCookieName)
		if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:   githubStateCookieName,
			Path:   "/",
			MaxAge: -1,
		})

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		token, err := api.GitHubApp.ExchangeCode(r.Context(), code)
		if err != nil {
			slog.Error("failed to exchange github code", "error", err)
			http.Error(w, "failed to connect GitHub account", http.StatusInternalServerError)
			return
		}

		// Store the token via deployer
		if err := api.StoreGitHubToken(r.Context(), claims.Subject, token); err != nil {
			slog.Error("failed to store github token", "error", err, "user", claims.Subject)
			http.Error(w, "failed to store GitHub token", http.StatusInternalServerError)
			return
		}

		slog.Info("github account connected", "user", claims.Subject)
		http.Redirect(w, r, dashboardURL+"/?github=account_connected", http.StatusTemporaryRedirect)
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
