package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/gateway/handler"
)

// ejectHandler returns an HTTP handler that streams an ejected project as a zip download.
func ejectHandler(api *handler.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := auth.FromContext(r.Context())
		if claims == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract project ID from URL: /api/eject/{projectId}
		// Project IDs are "org/name", so the path has two segments after /api/eject/.
		projectID := strings.TrimPrefix(r.URL.Path, "/api/eject/")
		projectID = strings.TrimSuffix(projectID, "/")
		if projectID == "" {
			http.Error(w, "project ID required", http.StatusBadRequest)
			return
		}

		archive, err := api.Eject(r.Context(), projectID)
		if err != nil {
			slog.Error("eject failed", "project", projectID, "error", err)
			http.Error(w, "eject failed", http.StatusInternalServerError)
			return
		}

		shortName := projectID
		if parts := strings.SplitN(projectID, "/", 2); len(parts) == 2 {
			shortName = parts[1]
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf(`attachment; filename="%s-ejected.zip"`, shortName))
		w.Header().Set("Content-Length", strconv.Itoa(len(archive)))
		w.Write(archive)
	}
}
