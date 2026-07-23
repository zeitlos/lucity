package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
)

// TODO(stage-6b): delete these cookie names + directSignIn. They belong to the
// server-side OIDC login/callback/refresh flow (handleLogin/handleCallback/
// handleRefresh/handleLogout) which is removed in stage-6b; the dashboard and
// CLI now run the OIDC flow themselves and present bearer tokens.
const (
	stateCookieName    = "lucity_oauth_state"
	verifierCookieName = "lucity_pkce_verifier"
	sessionCookieName  = "lucity_session" // HMAC-signed session JWT (auth claims)
	tokenCookieName    = "lucity_token"   // Logto opaque access token (Account API)
	refreshCookieName  = "lucity_refresh"
)

// directSignIn skips Logto's connector picker and goes straight to GitHub,
// the only configured sign-in method.
const directSignIn = "social:github"

type OIDCProvider struct {
	provider    *oidc.Provider
	verifier    *oidc.IDTokenVerifier
	oauthConfig oauth2.Config
	httpClient  *http.Client
}

func NewOIDCProvider(ctx context.Context, issuerURL, discoveryURL, clientID, clientSecret, callbackURL string) (*OIDCProvider, error) {
	var httpClient *http.Client
	if discoveryURL != "" {
		var err error
		httpClient, err = newIssuerRewriteClient(issuerURL, discoveryURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create internal HTTP client: %w", err)
		}
		ctx = oidc.ClientContext(ctx, httpClient)
		slog.Info("OIDC using internal discovery URL", "issuer", issuerURL, "discovery", discoveryURL)
	}

	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC provider at %s: %w", issuerURL, err)
	}

	oauthConfig := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  callbackURL,
		Scopes: []string{
			oidc.ScopeOpenID, "profile", "email", oidc.ScopeOfflineAccess,
			"identities",
			"urn:logto:scope:organizations",
			"urn:logto:scope:organization_roles",
		},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCProvider{
		provider:    provider,
		verifier:    verifier,
		oauthConfig: oauthConfig,
		httpClient:  httpClient,
	}, nil
}

// httpContext returns a context with the internal HTTP client injected, so that
// server-side calls (token exchange, JWKS refresh) route through internal DNS.
func (p *OIDCProvider) httpContext(ctx context.Context) context.Context {
	if p.httpClient != nil {
		return oidc.ClientContext(ctx, p.httpClient)
	}
	return ctx
}

// issuerRewriteTransport rewrites HTTP requests from the public issuer host to
// an internal service URL. This lets the OIDC library validate the issuer normally
// while all HTTP traffic stays cluster-internal.
type issuerRewriteTransport struct {
	publicHost  string
	internalURL *url.URL
	base        http.RoundTripper
}

func (t *issuerRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == t.publicHost {
		req = req.Clone(req.Context())
		req.URL.Scheme = t.internalURL.Scheme
		req.URL.Host = t.internalURL.Host
	}
	return t.base.RoundTrip(req)
}

func newIssuerRewriteClient(issuerURL, discoveryURL string) (*http.Client, error) {
	pub, err := url.Parse(issuerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer URL %q: %w", issuerURL, err)
	}
	internal, err := url.Parse(discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("invalid discovery URL %q: %w", discoveryURL, err)
	}

	return &http.Client{
		Transport: &issuerRewriteTransport{
			publicHost:  pub.Host,
			internalURL: internal,
			base:        http.DefaultTransport,
		},
	}, nil
}

// TODO(stage-6b): delete secureCookies — its only remaining callers set cookies
// on the login/callback/refresh handlers and newTokenRefresher, all removed.
//
// secureCookies returns true if cookies should have the Secure flag set,
// derived from whether the dashboard URL uses HTTPS.
func secureCookies(dashboardURL string) bool {
	return strings.HasPrefix(dashboardURL, "https://")
}

// TODO(stage-6b): prune registerAuthRoutes to only /auth/config, /auth/ci/exchange,
// and /auth/github/{install,setup}. Delete the /auth/{login,callback,me,logout,
// refresh,cli/exchange} registrations and the now-unused params (provider,
// sessionSecret, and the `secure` local). Also unthread provider + sessionSecret
// from NewGraphQLServer (graphql.go) and main.go.
//
// registerAuthRoutes adds OIDC auth endpoints to the mux.
func registerAuthRoutes(mux *http.ServeMux, provider *OIDCProvider, conductor *conductor.Client, logtoClient *logto.Client, sessionSecret, dashboardURL, githubAppSlug, oidcIssuer, oidcAudience, oidcDashboardClientID, oidcCLIClientID string, ciVerifier *githubActionsVerifier, ciSessionTTL time.Duration) {
	secure := secureCookies(dashboardURL)
	mux.HandleFunc("/auth/config", handleAuthConfig(oidcIssuer, oidcAudience, oidcDashboardClientID, oidcCLIClientID))
	mux.HandleFunc("/auth/login", handleLogin(provider, secure))                                                            // TODO(stage-6b): delete
	mux.HandleFunc("/auth/callback", handleCallback(provider, conductor, logtoClient, sessionSecret, dashboardURL, secure)) // TODO(stage-6b): delete
	mux.HandleFunc("/auth/me", handleMe())                                                                                  // TODO(stage-6b): delete
	mux.HandleFunc("/auth/logout", handleLogout(dashboardURL))                                                              // TODO(stage-6b): delete
	mux.HandleFunc("/auth/refresh", handleRefresh(provider, logtoClient, sessionSecret, secure))                            // TODO(stage-6b): delete
	mux.HandleFunc("/auth/cli/exchange", handleCLIExchange(sessionSecret))                                                  // TODO(stage-6b): delete
	mux.HandleFunc("/auth/github/install", handleGitHubInstall(githubAppSlug))
	mux.HandleFunc("/auth/github/setup", handleGitHubSetup(dashboardURL))

	if ciVerifier != nil {
		mux.HandleFunc("/auth/ci/exchange", handleCIExchange(ciVerifier, conductor, oidcAudience))
		slog.Info("keyless CI deploy exchange enabled", "audience", ciVerifier.audience)
	}
}

// handleAuthConfig advertises the OIDC issuer and API audience so non-interactive
// clients can exchange credentials directly with the identity provider.
func handleAuthConfig(issuer, audience, dashboardClientID, cliClientID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"issuer":            issuer,
			"endpoint":          strings.TrimSuffix(issuer, "/oidc"),
			"audience":          audience,
			"dashboardClientId": dashboardClientID,
			"cliClientId":       cliClientID,
		})
	}
}

// TODO(stage-6b): delete handleLogin — the dashboard and CLI initiate the OIDC
// flow themselves; the server no longer redirects to the authorization page.
//
// handleLogin redirects to the OIDC provider's authorization page with PKCE.
func handleLogin(provider *OIDCProvider, secure bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliPort := r.URL.Query().Get("cli_port")
		cliState := r.URL.Query().Get("cli_state")
		if cliPort != "" {
			if !validCLIPort(cliPort) || !validCLIState(cliState) {
				http.Error(w, "invalid cli parameters", http.StatusBadRequest)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     cliCookieName,
				Value:    cliPort + ":" + cliState,
				Path:     "/",
				MaxAge:   600,
				HttpOnly: true,
				Secure:   secure,
				SameSite: http.SameSiteLaxMode,
			})
		}

		state := generateState()
		verifier := generateCodeVerifier()
		challenge := codeChallenge(verifier)

		http.SetCookie(w, &http.Cookie{
			Name:     stateCookieName,
			Value:    state,
			Path:     "/",
			MaxAge:   600, // 10 minutes
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     verifierCookieName,
			Value:    verifier,
			Path:     "/",
			MaxAge:   600,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})

		url := provider.oauthConfig.AuthCodeURL(state,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
			oauth2.SetAuthURLParam("direct_sign_in", directSignIn),
		)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// TODO(stage-6b): delete handleCallback — the OIDC redirect callback is handled
// client-side (dashboard /callback route, CLI loopback). First-login side
// effects (username + personal workspace provisioning) now run lazily via
// conductor.EnsureAccount in the workspaces resolver.
//
// handleCallback exchanges the auth code for tokens, verifies the ID token,
// extracts claims, and creates a session. Stores two cookies:
// - lucity_session: HMAC-signed JWT with auth claims and workspace memberships
// - lucity_token: Logto opaque access token for Account API calls (e.g. GitHub token)
func handleCallback(provider *OIDCProvider, conductor *conductor.Client, logtoClient *logto.Client, sessionSecret, dashboardURL string, secure bool) http.HandlerFunc {
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

		// Exchange code for OAuth2 token (with PKCE verifier).
		// Use httpContext so the token exchange routes through internal DNS.
		oauth2Token, err := provider.oauthConfig.Exchange(provider.httpContext(r.Context()), code,
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

		idToken, err := provider.verifier.Verify(provider.httpContext(r.Context()), rawIDToken)
		if err != nil {
			slog.Error("failed to verify id token", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		// Extract claims from the ID token for user identity
		var oidcClaims struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Picture  string `json:"picture"`
			Username string `json:"username"`
		}
		if err := idToken.Claims(&oidcClaims); err != nil {
			slog.Error("failed to extract claims", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		// The Logto username is the stable source of truth for workspace ID.
		// On first login it's empty — derive it from the social identity (GitHub/
		// GitLab/Bitbucket login), sanitize it, and persist it on Logto with
		// conflict resolution. On subsequent logins, read it straight from the
		// ID token claims.
		username := oidcClaims.Username
		if username == "" && logtoClient != nil {
			socialLogin, err := logtoClient.SocialLogin(r.Context(), idToken.Subject)
			if err != nil {
				slog.Error("failed to fetch social login from Logto", "error", err)
			} else if socialLogin != "" {
				set, err := logtoClient.EnsureUsername(r.Context(), idToken.Subject, socialLogin)
				if err != nil {
					slog.Error("failed to set Logto username", "error", err, "login", socialLogin)
				} else {
					slog.Info("username set from social login", "username", set, "login", socialLogin, "email", oidcClaims.Email)
					username = set
				}
			}
		}
		if username == "" {
			slog.Warn("no username for personal workspace", "email", oidcClaims.Email)
			http.Redirect(w, r, dashboardURL+"/login?error=no_workspace", http.StatusTemporaryRedirect)
			return
		}

		// Set claims on context so auth.OutgoingContext() can propagate identity
		// to backend services via gRPC metadata. No JWT needed — backends trust
		// the gateway as the auth boundary.
		svcCtx := auth.NewContext(r.Context(), &auth.Claims{
			Subject: idToken.Subject,
			Email:   oidcClaims.Email,
		})

		personalWSID, isNewUser, err := conductor.EnsurePersonalWorkspace(svcCtx, idToken.Subject, username)
		if err != nil {
			slog.Error("failed to ensure personal workspace", "error", err, "email", oidcClaims.Email)
			http.Error(w, "failed to create workspace", http.StatusInternalServerError)
			return
		}

		slog.Info("personal workspace ensured", "email", oidcClaims.Email, "workspace", personalWSID)

		// Build workspace memberships for the session token.
		// Fetch all orgs the user belongs to (not just the personal one).
		var workspaces []auth.WorkspaceMembership
		if logtoClient != nil {
			userOrgs, err := logtoClient.UserOrganizations(r.Context(), idToken.Subject)
			if err != nil {
				slog.Warn("failed to fetch user organizations for session", "error", err)
			} else {
				for _, org := range userOrgs {
					role := auth.WorkspaceRoleUser
					// Check if user has admin role in this org
					roles, rolesErr := logtoClient.MemberRoles(r.Context(), org.ID, idToken.Subject)
					if rolesErr == nil {
						for _, r := range roles {
							if r.Name == "admin" {
								role = auth.WorkspaceRoleAdmin
								break
							}
						}
					}
					workspaces = append(workspaces, auth.WorkspaceMembership{
						Workspace: org.Name, // org name = workspace ID
						Role:      role,
					})
				}
			}
		}

		// Mint session JWT with all claims and workspace memberships.
		sessionClaims := &auth.Claims{
			Subject:    idToken.Subject,
			Name:       oidcClaims.Name,
			Email:      oidcClaims.Email,
			AvatarURL:  oidcClaims.Picture,
			Workspaces: workspaces,
		}
		sessionToken, err := mintSessionToken(sessionSecret, sessionClaims)
		if err != nil {
			slog.Error("failed to mint session token", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		// Session cookie: HMAC JWT verified by auth middleware
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    sessionToken,
			Path:     "/",
			MaxAge:   7 * 24 * 3600, // 7 days
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})

		// Logto access token cookie: opaque token for Account API calls (e.g. GitHub token)
		http.SetCookie(w, &http.Cookie{
			Name:     tokenCookieName,
			Value:    oauth2Token.AccessToken,
			Path:     "/",
			MaxAge:   7 * 24 * 3600,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})

		// Store refresh token for silent token renewal
		slog.Debug("oauth2 token response", "has_refresh_token", oauth2Token.RefreshToken != "", "has_access_token", oauth2Token.AccessToken != "", "token_type", oauth2Token.TokenType, "expiry", oauth2Token.Expiry)
		if oauth2Token.RefreshToken != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     refreshCookieName,
				Value:    oauth2Token.RefreshToken,
				Path:     "/",
				MaxAge:   30 * 24 * 3600, // 30 days
				HttpOnly: true,
				Secure:   secure,
				SameSite: http.SameSiteLaxMode,
			})
		}

		slog.Info("user authenticated", "email", oidcClaims.Email)

		if cliCookie, err := r.Cookie(cliCookieName); err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:   cliCookieName,
				Path:   "/",
				MaxAge: -1,
			})

			cliPort, cliState, ok := strings.Cut(cliCookie.Value, ":")
			if ok && validCLIPort(cliPort) && validCLIState(cliState) {
				code, err := sealCLICode(sessionSecret, cliHandoff{
					Token:        sessionToken,
					RefreshToken: oauth2Token.RefreshToken,
					LogtoToken:   oauth2Token.AccessToken,
					ExpiresAt:    time.Now().Add(sessionExpiry),
				})
				if err != nil {
					slog.Error("failed to seal cli handoff code", "error", err)
				} else {
					cliURL := fmt.Sprintf("http://127.0.0.1:%s/callback?code=%s&state=%s",
						cliPort, url.QueryEscape(code), url.QueryEscape(cliState))
					http.Redirect(w, r, cliURL, http.StatusTemporaryRedirect)
					return
				}
			}
		}

		redirectURL := dashboardURL
		if isNewUser {
			redirectURL = dashboardURL + "/?welcome=true"
		}
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// TODO(stage-6b): delete handleMe (/auth/me). The dashboard and CLI read the
// profile + memberships straight from the IdP (ID token / userinfo).
//
// handleMe returns the current user's profile from the JWT claims in context.
func handleMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.FromContext(r.Context())

		if err != nil {
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
			"id":         claims.Subject,
			"name":       claims.Name,
			"email":      claims.Email,
			"avatarUrl":  claims.AvatarURL,
			"workspaces": workspaces,
		})
	}
}

// TODO(stage-6b): delete handleLogout (/auth/logout). Sign-out is client-side
// (Logto signOut); there are no server session cookies to clear.
//
// handleLogout clears the session cookies.
func handleLogout(dashboardURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   sessionCookieName,
			Path:   "/",
			MaxAge: -1,
		})
		http.SetCookie(w, &http.Cookie{
			Name:   tokenCookieName,
			Path:   "/",
			MaxAge: -1,
		})
		http.SetCookie(w, &http.Cookie{
			Name:   refreshCookieName,
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

// TODO(stage-6b): delete handleRefresh (/auth/refresh). Token refresh is done by
// the clients directly against the IdP; the server no longer holds refresh
// tokens or re-mints session JWTs.
//
// handleRefresh uses the Logto refresh token to obtain a new access token
// and re-mints the session JWT with fresh workspace memberships.
// Called by the dashboard when the access token expires.
func handleRefresh(provider *OIDCProvider, logtoClient *logto.Client, sessionSecret string, secure bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		jsonRequest := strings.Contains(r.Header.Get("Content-Type"), "application/json")

		var refreshToken string
		if jsonRequest {
			var body struct {
				RefreshToken string `json:"refreshToken"`
			}
			if err := json.NewDecoder(io.LimitReader(r.Body, 16<<10)).Decode(&body); err != nil {
				http.Error(w, "no refresh token", http.StatusUnauthorized)
				return
			}
			refreshToken = body.RefreshToken
		} else {
			refreshCookie, err := r.Cookie(refreshCookieName)
			if err == nil {
				refreshToken = refreshCookie.Value
			}
		}
		if refreshToken == "" {
			http.Error(w, "no refresh token", http.StatusUnauthorized)
			return
		}

		// Use the OIDC provider's oauth config for refresh — the refresh token
		// was issued for the OIDC public client, not the M2M app.
		tokenSource := provider.oauthConfig.TokenSource(provider.httpContext(r.Context()), &oauth2.Token{
			RefreshToken: refreshToken,
		})
		oauth2Token, err := tokenSource.Token()
		if err != nil {
			slog.Error("failed to refresh access token", "error", err)
			http.Error(w, "failed to refresh token", http.StatusUnauthorized)
			return
		}
		newAccessToken := oauth2Token.AccessToken
		newRefreshToken := oauth2Token.RefreshToken

		var reMintedToken string
		var reMintedExpiry time.Time

		// Re-mint session JWT if we have existing claims (refreshes workspace memberships)
		claims, err := auth.FromContext(r.Context())

		if err == nil && logtoClient != nil {
			userOrgs, err := logtoClient.UserOrganizations(r.Context(), claims.Subject)
			if err == nil {
				var workspaces []auth.WorkspaceMembership
				for _, org := range userOrgs {
					role := auth.WorkspaceRoleUser
					roles, rolesErr := logtoClient.MemberRoles(r.Context(), org.ID, claims.Subject)
					if rolesErr == nil {
						for _, r := range roles {
							if r.Name == "admin" {
								role = auth.WorkspaceRoleAdmin
								break
							}
						}
					}
					workspaces = append(workspaces, auth.WorkspaceMembership{
						Workspace: org.Name,
						Role:      role,
					})
				}
				claims.Workspaces = workspaces
			}

			sessionToken, err := mintSessionToken(sessionSecret, claims)
			if err == nil {
				reMintedToken = sessionToken
				reMintedExpiry = time.Now().Add(sessionExpiry)
				http.SetCookie(w, &http.Cookie{
					Name:     sessionCookieName,
					Value:    sessionToken,
					Path:     "/",
					MaxAge:   7 * 24 * 3600,
					HttpOnly: true,
					Secure:   secure,
					SameSite: http.SameSiteLaxMode,
				})
			}
		}

		// Update the Logto access token cookie
		http.SetCookie(w, &http.Cookie{
			Name:     tokenCookieName,
			Value:    newAccessToken,
			Path:     "/",
			MaxAge:   7 * 24 * 3600,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})

		if newRefreshToken != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     refreshCookieName,
				Value:    newRefreshToken,
				Path:     "/",
				MaxAge:   30 * 24 * 3600,
				HttpOnly: true,
				Secure:   secure,
				SameSite: http.SameSiteLaxMode,
			})
		}

		slog.Info("token refreshed")

		if jsonRequest {
			rotatedRefresh := newRefreshToken
			if rotatedRefresh == "" {
				rotatedRefresh = refreshToken
			}

			resp := map[string]string{
				"token":        reMintedToken,
				"refreshToken": rotatedRefresh,
				"logtoToken":   newAccessToken,
			}
			if reMintedToken != "" {
				resp["expiresAt"] = reMintedExpiry.Format(time.RFC3339)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// handleGitHubInstall redirects to GitHub App installation page.
// Users can install the GitHub App on new accounts to make them available as sources.
// Passes redirect_url so GitHub redirects back to our setup handler after installation,
// regardless of how the App's "Setup URL" is configured on GitHub.
func handleGitHubInstall(githubAppSlug string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if githubAppSlug == "" {
			http.Error(w, "GitHub App not configured", http.StatusServiceUnavailable)
			return
		}

		if _, err := auth.FromContext(r.Context()); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Use the request's scheme+host to build the setup callback URL.
		// This ensures the redirect works in both dev (localhost:8080) and prod (lucity.cloud).
		scheme := r.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			if r.TLS != nil {
				scheme = "https"
			} else {
				scheme = "http"
			}
		}
		host := r.Header.Get("X-Forwarded-Host")
		if host == "" {
			host = r.Host
		}
		setupURL := fmt.Sprintf("%s://%s/auth/github/setup", scheme, host)

		installURL := fmt.Sprintf("https://github.com/apps/%s/installations/new?redirect_url=%s",
			githubAppSlug, url.QueryEscape(setupURL))
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

// TODO(stage-6b): delete newTokenRefresher and drop the tokenRefresher argument
// from conductor.New (main.go). Clients now send a fresh Account-API token on
// each request (X-Lucity-Account-Token), so the server never refreshes Logto
// tokens on their behalf. Also revisit userGitHubToken's refresher branch
// (internal/conductor/github.go), which becomes dead once this is nil.
func newTokenRefresher(provider *OIDCProvider, secure bool) conductor.TokenRefresher {
	return func(ctx context.Context, refreshToken string) (string, error) {
		tokenSource := provider.oauthConfig.TokenSource(provider.httpContext(ctx), &oauth2.Token{
			RefreshToken: refreshToken,
		})
		oauth2Token, err := tokenSource.Token()
		if err != nil {
			return "", fmt.Errorf("token refresh failed: %w", err)
		}

		// Write updated cookies to the response
		w := auth.ResponseWriterFrom(ctx)
		if w != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     tokenCookieName,
				Value:    oauth2Token.AccessToken,
				Path:     "/",
				MaxAge:   7 * 24 * 3600,
				HttpOnly: true,
				Secure:   secure,
				SameSite: http.SameSiteLaxMode,
			})
			if oauth2Token.RefreshToken != "" {
				http.SetCookie(w, &http.Cookie{
					Name:     refreshCookieName,
					Value:    oauth2Token.RefreshToken,
					Path:     "/",
					MaxAge:   30 * 24 * 3600,
					HttpOnly: true,
					Secure:   secure,
					SameSite: http.SameSiteLaxMode,
				})
			}
		}

		slog.Info("logto access token refreshed via token refresher")
		return oauth2Token.AccessToken, nil
	}
}

// TODO(stage-6b): delete generateState, generateCodeVerifier, and codeChallenge
// — they are only used by handleLogin's server-side PKCE, which is removed.
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
