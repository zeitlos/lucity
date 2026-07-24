package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/oidc"
	"github.com/zeitlos/lucity/pkg/session"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
)

const (
	stateCookieName     = "lucity_oauth_state"
	verifierCookieName  = "lucity_pkce_verifier"
	sessionCookieName   = "lucity_session"
	bootstrapCookieName = "lucity_bootstrap"
	directSignIn        = "social:github"
)

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

func secureCookies(dashboardURL string) bool {
	return strings.HasPrefix(dashboardURL, "https://")
}

func registerAuthRoutes(mux *http.ServeMux, provider *oidc.Provider, store *sessionStore, codec *session.Codec, conductorClient *conductor.Client, logtoClient *logto.Client, callbackURL, dashboardURL, githubAppSlug, oidcIssuer, oidcAudience, oidcCLIClientID string, ciVerifier *githubActionsVerifier) {
	secure := secureCookies(dashboardURL)
	mux.HandleFunc("/auth/config", handleAuthConfig(oidcIssuer, oidcAudience, oidcCLIClientID))
	mux.HandleFunc("/auth/login", handleLogin(provider, callbackURL, secure))
	mux.HandleFunc("/auth/callback", handleCallback(provider, store, codec, conductorClient, callbackURL, dashboardURL, secure))
	mux.HandleFunc("/auth/me", handleMe(logtoClient))
	mux.HandleFunc("/auth/logout", handleLogout(store, codec, dashboardURL))
	mux.HandleFunc("/auth/github/install", handleGitHubInstall(githubAppSlug))
	mux.HandleFunc("/auth/github/setup", handleGitHubSetup(dashboardURL))

	if ciVerifier != nil {
		mux.HandleFunc("/auth/ci/exchange", handleCIExchange(ciVerifier, conductorClient, oidcAudience))
		slog.Info("keyless CI deploy exchange enabled", "audience", ciVerifier.audience)
	}
}

func handleAuthConfig(issuer, audience, cliClientID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"issuer":      issuer,
			"endpoint":    strings.TrimSuffix(issuer, "/oidc"),
			"audience":    audience,
			"cliClientId": cliClientID,
		})
	}
}

func handleLogin(provider *oidc.Provider, callbackURL string, secure bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := generateState()
		verifier := generateCodeVerifier()
		setShortCookie(w, stateCookieName, state, secure)
		setShortCookie(w, verifierCookieName, verifier, secure)
		http.Redirect(w, r, provider.AuthCodeURL(callbackURL, state, codeChallenge(verifier)), http.StatusTemporaryRedirect)
	}
}

func handleCallback(provider *oidc.Provider, store *sessionStore, codec *session.Codec, conductorClient *conductor.Client, callbackURL, dashboardURL string, secure bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stateCookie, err := r.Cookie(stateCookieName)
		if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		clearCookie(w, stateCookieName)

		verifierCookie, err := r.Cookie(verifierCookieName)
		if err != nil {
			http.Error(w, "missing PKCE verifier", http.StatusBadRequest)
			return
		}
		clearCookie(w, verifierCookieName)

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		tokens, err := provider.Exchange(r.Context(), code, callbackURL, verifierCookie.Value)
		if err != nil {
			slog.Error("code exchange failed", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}
		if tokens.RefreshToken == "" {
			slog.Error("no refresh token in token response")
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		info, err := provider.UserInfo(r.Context(), tokens.AccessToken)
		if err != nil {
			slog.Error("userinfo request failed", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		svcCtx := auth.NewContext(r.Context(), &auth.Claims{Subject: info.Subject, Email: info.Email})
		_, created, err := conductorClient.EnsureAccount(svcCtx, info.Subject)
		if err != nil {
			slog.Error("failed to provision account", "error", err, "email", info.Email)
			http.Error(w, "failed to create workspace", http.StatusInternalServerError)
			return
		}

		if created && !bootstrapRetry(r) {
			slog.Info("new account provisioned; re-authenticating for org-aware session", "email", info.Email)
			setShortCookie(w, bootstrapCookieName, "1", secure)
			http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
			return
		}
		clearCookie(w, bootstrapCookieName)

		sid, err := store.create(r.Context(), tokens.RefreshToken)
		if err != nil {
			slog.Error("failed to create session", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}
		if err := codec.Save(w, sessionData{SID: sid, Sub: info.Subject, Name: info.Name, Email: info.Email, Picture: info.Picture, RefreshToken: tokens.RefreshToken}); err != nil {
			slog.Error("failed to save session", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		slog.Info("user authenticated", "email", info.Email)
		http.Redirect(w, r, dashboardURL, http.StatusTemporaryRedirect)
	}
}

func handleMe(logtoClient *logto.Client) http.HandlerFunc {
	type workspaceEntry struct {
		Workspace string             `json:"workspace"`
		Role      auth.WorkspaceRole `json:"role"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.FromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		workspaces := []workspaceEntry{}
		if logtoClient != nil {
			if orgs, err := logtoClient.UserOrganizations(r.Context(), claims.Subject); err == nil {
				for _, org := range orgs {
					role := auth.WorkspaceRoleUser
					if roles, err := logtoClient.MemberRoles(r.Context(), org.ID, claims.Subject); err == nil {
						for _, role2 := range roles {
							if role2.Name == "admin" {
								role = auth.WorkspaceRoleAdmin
								break
							}
						}
					}
					workspaces = append(workspaces, workspaceEntry{Workspace: org.Name, Role: role})
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":         claims.Subject,
			"name":       claims.Name,
			"email":      claims.Email,
			"avatarUrl":  claims.AvatarURL,
			"workspaces": workspaces,
		})
	}
}

func handleLogout(store *sessionStore, codec *session.Codec, dashboardURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data sessionData
		if err := codec.Load(r, &data); err == nil {
			store.delete(r.Context(), data.SID)
		}
		codec.Clear(w)

		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, dashboardURL+"/login", http.StatusTemporaryRedirect)
	}
}

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
		installURL := fmt.Sprintf("https://github.com/apps/%s/installations/new?redirect_url=%s", githubAppSlug, url.QueryEscape(setupURL))
		http.Redirect(w, r, installURL, http.StatusTemporaryRedirect)
	}
}

func handleGitHubSetup(dashboardURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html><html><head><title>GitHub App Installed</title></head><body><script>
if (window.opener) { window.opener.postMessage("github-app-installed", %q); window.close(); }
else { window.location.href = %q; }
</script><p>GitHub App installed. You can close this window.</p></body></html>`, dashboardURL, dashboardURL+"/?github=installed")
	}
}

func setShortCookie(w http.ResponseWriter, name, value string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Path: "/", MaxAge: -1})
}

func bootstrapRetry(r *http.Request) bool {
	cookie, err := r.Cookie(bootstrapCookieName)
	return err == nil && cookie.Value == "1"
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func codeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
