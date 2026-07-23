package tenant

import (
	"net/http"

	"github.com/zeitlos/lucity/pkg/auth"
)

// Middleware extracts the workspace identifier from the X-Lucity-Workspace
// header and attaches it to the request context. It does NOT reject requests
// without the header — validation happens via Require at the handler level.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(stage-6b): drop the X-Lucity-Workspace header fallback. It exists
		// for legacy HS256 tokens that carried multiple workspaces and needed the
		// header to pick the active one. Once HS256 is gone, every bearer is a
		// single-workspace org token (or a workspace-less account token), so the
		// workspace should come solely from the token — delete the header read
		// and keep only the claims-derived branch.
		ws := r.Header.Get(Header)

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
