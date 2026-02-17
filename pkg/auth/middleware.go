package auth

import (
	"net/http"
	"strings"
)

const cookieName = "lucity_token"

// Middleware returns an HTTP middleware that extracts a JWT from the
// Authorization header or cookie and attaches claims to the request context.
// It does NOT reject unauthenticated requests — that's the GraphQL directive's job.
func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := ParseToken(tokenString, secret)
			if err != nil {
				// Invalid token — treat as unauthenticated, let the directive handle it
				next.ServeHTTP(w, r)
				return
			}

			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken gets the JWT from the Authorization header or cookie.
func extractToken(r *http.Request) string {
	// Check Authorization header first
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Fall back to cookie
	if cookie, err := r.Cookie(cookieName); err == nil {
		return cookie.Value
	}

	return ""
}
