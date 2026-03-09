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
	"strconv"
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
	githubWSCookieName = "lucity_github_ws"
)

// registerAuthRoutes adds OIDC auth endpoints to the mux.
func registerAuthRoutes(mux *http.ServeMux, provider *OIDCProvider, api *handler.Client, jwtSecret, dashboardURL, githubAppSlug string) {
	mux.HandleFunc("/auth/login", handleLogin(provider))
	mux.HandleFunc("/auth/callback", handleCallback(provider, api, jwtSecret, dashboardURL))
	mux.HandleFunc("/auth/me", handleMe())
	mux.HandleFunc("/auth/logout", handleLogout(dashboardURL))
	mux.HandleFunc("/auth/refresh", handleRefresh(api, jwtSecret))
	mux.HandleFunc("/auth/github/install", handleGitHubInstall(githubAppSlug))
	mux.HandleFunc("/auth/github/setup", handleGitHubSetup(api, dashboardURL))
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

		// Parse workspace memberships from Rauthy groups
		workspaces := auth.ParseRauthyGroups(oidcClaims.Groups)

		// Auto-create personal workspace for new users with no workspace memberships
		if len(workspaces) == 0 {
			if oidcClaims.PreferredUsername == "" {
				slog.Warn("user has no workspaces and no preferred_username for personal workspace", "email", oidcClaims.Email)
				http.Redirect(w, r, dashboardURL+"/login?error=no_workspace", http.StatusTemporaryRedirect)
				return
			}

			// Create a service-level context with auth for gRPC calls
			svcCtx := auth.WithClaims(r.Context(), &auth.Claims{
				Subject: idToken.Subject,
				Email:   oidcClaims.Email,
				Roles:   []auth.Role{auth.RoleUser},
			})

			wsID, err := api.EnsurePersonalWorkspace(svcCtx, idToken.Subject, oidcClaims.PreferredUsername)
			if err != nil {
				slog.Error("failed to create personal workspace", "error", err, "email", oidcClaims.Email)
				http.Error(w, "failed to create workspace", http.StatusInternalServerError)
				return
			}
			workspaces = []auth.WorkspaceMembership{{Workspace: wsID, Role: auth.WorkspaceRoleAdmin}}
			slog.Info("personal workspace created for new user", "email", oidcClaims.Email, "workspace", wsID)
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

		// Fetch current user from Rauthy to get updated group list
		user, err := api.Rauthy.User(r.Context(), claims.Subject)
		if err != nil {
			slog.Error("failed to fetch user for token refresh", "error", err, "subject", claims.Subject)
			http.Error(w, "failed to refresh token", http.StatusInternalServerError)
			return
		}

		// Resolve group IDs to group names
		allGroups, err := api.Rauthy.Groups(r.Context())
		if err != nil {
			slog.Error("failed to fetch groups for token refresh", "error", err)
			http.Error(w, "failed to refresh token", http.StatusInternalServerError)
			return
		}

		groupIDToName := make(map[string]string, len(allGroups))
		for _, g := range allGroups {
			groupIDToName[g.ID] = g.Name
		}

		groupNames := make([]string, 0, len(user.Groups))
		for _, gid := range user.Groups {
			if name, ok := groupIDToName[gid]; ok {
				groupNames = append(groupNames, name)
			}
		}

		// Parse workspace memberships
		workspaces := auth.ParseRauthyGroups(groupNames)

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
// Sets a cookie to remember which workspace to link after installation.
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

		workspace := r.URL.Query().Get("workspace")
		if workspace == "" {
			http.Error(w, "missing workspace parameter", http.StatusBadRequest)
			return
		}

		// Verify user is admin of this workspace
		if claims.WorkspaceRoleIn(workspace) != auth.WorkspaceRoleAdmin {
			http.Error(w, "forbidden: workspace admin role required", http.StatusForbidden)
			return
		}

		// Set cookie to remember workspace during GitHub redirect
		http.SetCookie(w, &http.Cookie{
			Name:     githubWSCookieName,
			Value:    workspace,
			Path:     "/",
			MaxAge:   600, // 10 minutes
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		installURL := fmt.Sprintf("https://github.com/apps/%s/installations/new", githubAppSlug)
		http.Redirect(w, r, installURL, http.StatusTemporaryRedirect)
	}
}

// handleGitHubSetup is the callback from GitHub after App installation.
// Links the installation to the workspace stored in the cookie.
func handleGitHubSetup(api *handler.Client, dashboardURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		installationIDStr := r.URL.Query().Get("installation_id")
		if installationIDStr == "" {
			http.Error(w, "missing installation_id", http.StatusBadRequest)
			return
		}

		installationID, err := strconv.ParseInt(installationIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid installation_id", http.StatusBadRequest)
			return
		}

		// Read workspace from cookie
		wsCookie, err := r.Cookie(githubWSCookieName)
		if err != nil || wsCookie.Value == "" {
			http.Error(w, "missing workspace context — please try again from workspace settings", http.StatusBadRequest)
			return
		}
		workspace := wsCookie.Value

		// Clear cookie
		http.SetCookie(w, &http.Cookie{
			Name:   githubWSCookieName,
			Path:   "/",
			MaxAge: -1,
		})

		// Verify user is admin
		if claims.WorkspaceRoleIn(workspace) != auth.WorkspaceRoleAdmin {
			http.Error(w, "forbidden: workspace admin role required", http.StatusForbidden)
			return
		}

		// Create a context with workspace tenant for the handler
		ctx := auth.WithClaims(r.Context(), claims)

		// Link installation via handler (which needs tenant context)
		if _, err := api.LinkGitHubInstallationDirect(ctx, workspace, installationID); err != nil {
			slog.Error("failed to link github installation", "error", err, "workspace", workspace, "installation_id", installationID)
			http.Error(w, "failed to link GitHub installation", http.StatusInternalServerError)
			return
		}

		slog.Info("github installation linked via setup callback", "workspace", workspace, "installation_id", installationID)
		http.Redirect(w, r, dashboardURL+"/workspace/settings?github=linked", http.StatusTemporaryRedirect)
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
