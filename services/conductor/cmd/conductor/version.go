package main

import (
	"encoding/json"
	"net/http"
)

// Version is set at build time via ldflags.
// This string is used as a fallback if the flag is not set (e.g. during local development).
var Version = "26.7.3-rc.1-39-g5ce734f.local"

type versionResponse struct {
	Version string `json:"version"`
}

func versionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(versionResponse{Version: Version})
	}
}
