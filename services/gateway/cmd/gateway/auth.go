package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	gh "github.com/zeitlos/lucity/pkg/github"
)

const (
	stateCookieName = "lucity_oauth_state"
	tokenCookieName = "lucity_token"
	tokenExpiry     = 7 * 24 * time.Hour // 1 week
)

// registerAuthRoutes adds OAuth and auth endpoints to the mux.
func registerAuthRoutes(mux *http.ServeMux, app *gh.App, jwtSecret, dashboardURL string) {
	mux.HandleFunc("/auth/github", handleGitHubLogin(app))
	mux.HandleFunc("/auth/github/callback", handleGitHubCallback(app, jwtSecret, dashboardURL))
	mux.HandleFunc("/auth/me", handleMe(jwtSecret))
	mux.HandleFunc("/auth/logout", handleLogout(dashboardURL))
}

// handleGitHubLogin redirects to GitHub's OAuth authorization page.
func handleGitHubLogin(app *gh.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := generateState()

		http.SetCookie(w, &http.Cookie{
			Name:     stateCookieName,
			Value:    state,
			Path:     "/",
			MaxAge:   600, // 10 minutes
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, app.OAuthURL(state), http.StatusTemporaryRedirect)
	}
}

// handleGitHubCallback exchanges the auth code for a token and creates a session.
func handleGitHubCallback(app *gh.App, jwtSecret, dashboardURL string) http.HandlerFunc {
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

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		// Exchange code for token
		token, err := app.ExchangeCode(r.Context(), code)
		if err != nil {
			slog.Error("failed to exchange code", "error", err)
			http.Error(w, "authentication failed", http.StatusInternalServerError)
			return
		}

		// Get user profile
		user, err := app.GetUser(r.Context(), token)
		if err != nil {
			slog.Error("failed to get user", "error", err)
			http.Error(w, "failed to get user profile", http.StatusInternalServerError)
			return
		}

		// Create JWT
		claims := &auth.Claims{
			Subject:     user.Name,
			Email:       user.Email,
			GitHubLogin: user.Login,
			AvatarURL:   user.AvatarURL,
			Roles:       []auth.Role{auth.RoleUser},
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

		slog.Info("user authenticated", "login", user.Login)
		http.Redirect(w, r, dashboardURL, http.StatusTemporaryRedirect)
	}
}

// handleMe returns the current user's profile from the JWT.
func handleMe(jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"login":     claims.GitHubLogin,
			"name":      claims.Subject,
			"email":     claims.Email,
			"avatarUrl": claims.AvatarURL,
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
