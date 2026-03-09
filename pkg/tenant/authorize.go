package tenant

import (
	"log/slog"
	"net/http"

	"github.com/zeitlos/lucity/pkg/auth"
)

// AuthorizeMiddleware validates that the authenticated user has access to
// the workspace specified in the X-Lucity-Workspace header.
// Must run after both auth.Middleware and tenant.Middleware.
func AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws := FromContext(r.Context())
		if ws == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims := auth.FromContext(r.Context())
		if claims == nil {
			// Not authenticated — let the GraphQL directive handle it.
			next.ServeHTTP(w, r)
			return
		}

		if !claims.IsMemberOf(ws) {
			slog.Warn("workspace access denied",
				"email", claims.Email,
				"workspace", ws,
			)
			http.Error(w, "forbidden: not a member of workspace", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
