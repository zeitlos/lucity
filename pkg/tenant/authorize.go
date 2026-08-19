package tenant

import (
	"log/slog"
	"net/http"

	"github.com/zeitlos/lucity/pkg/auth"
)

// AuthorizeMiddleware validates that the authenticated user has access to
// the workspace in the request context.
// Must run after the session auth middleware and tenant.Middleware.
func AuthorizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := FromContext(r.Context())

		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := auth.FromContext(r.Context())

		if err != nil {
			// Not authenticated. Let the GraphQL directive handle it.
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
