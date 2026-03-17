package auth

import (
	"net/http"
	"strings"
)

const (
	sessionCookieName = "lucity_session" // HMAC-signed session JWT (auth claims)
	tokenCookieName   = "lucity_token"   // Logto opaque access token (Account API)
)

// Middleware returns an HTTP middleware that extracts a session JWT from the
// Authorization header or session cookie and attaches claims to the request context.
// Also reads the Logto access token cookie for Account API calls.
// It does NOT reject unauthenticated requests — that's the GraphQL directive's job.
func Middleware(verifier *Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionToken := extractSessionToken(r)
			if sessionToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := verifier.ValidateToken(r.Context(), sessionToken)
			if err != nil {
				// Invalid token — treat as unauthenticated, let the directive handle it
				next.ServeHTTP(w, r)
				return
			}

			ctx := WithClaims(r.Context(), claims)

			// Store the Logto access token for Account API calls (e.g. GitHub token retrieval)
			if logtoToken := extractLogtoToken(r); logtoToken != "" {
				ctx = WithToken(ctx, logtoToken)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractSessionToken gets the session JWT from the Authorization header or session cookie.
func extractSessionToken(r *http.Request) string {
	// Check Authorization header first
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Fall back to session cookie
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		return cookie.Value
	}

	return ""
}

// extractLogtoToken reads the Logto opaque access token from the cookie.
func extractLogtoToken(r *http.Request) string {
	if cookie, err := r.Cookie(tokenCookieName); err == nil {
		return cookie.Value
	}
	return ""
}
