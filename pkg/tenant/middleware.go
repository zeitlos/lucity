package tenant

import (
	"net/http"

	"github.com/zeitlos/lucity/pkg/auth"
)

// Middleware derives the workspace identifier from the authenticated claims and
// attaches it to the request context. Every token is scoped to a single
// workspace, so the token is the only source. It does NOT reject requests
// without a workspace — AuthorizeMiddleware and the GraphQL role directive do.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ws string

		if claims, err := auth.FromContext(r.Context()); err == nil && len(claims.Workspaces) == 1 {
			ws = claims.Workspaces[0].Workspace
		}

		if ws != "" {
			ctx := NewContext(r.Context(), ws)
			ctx = auth.WithActiveWorkspace(ctx, ws)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
