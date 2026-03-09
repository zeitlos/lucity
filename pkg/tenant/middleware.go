package tenant

import "net/http"

// Middleware extracts the workspace identifier from the X-Lucity-Workspace
// header and attaches it to the request context. It does NOT reject requests
// without the header — validation happens via Require at the handler level.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws := r.Header.Get(Header)
		if ws != "" {
			ctx := WithWorkspace(r.Context(), ws)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
