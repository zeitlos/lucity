package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Version is set at build time via ldflags.
var Version = "dev"

type componentStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type versionResponse struct {
	Version    string            `json:"version"`
	Components []componentStatus `json:"components"`
}

type grpcComponent struct {
	name string
	conn *grpc.ClientConn
}

func versionHandler(components []grpcComponent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := versionResponse{
			Version: Version,
		}

		for _, c := range components {
			status := checkGRPC(r.Context(), c.conn)
			resp.Components = append(resp.Components, componentStatus{
				Name:   c.name,
				Status: status,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func checkGRPC(parent context.Context, conn *grpc.ClientConn) string {
	state := conn.GetState()
	if state == connectivity.Ready {
		return "UP"
	}

	// Trigger connection attempt and wait briefly.
	conn.Connect()
	ctx, cancel := context.WithTimeout(parent, 500*time.Millisecond)
	defer cancel()

	for {
		if !conn.WaitForStateChange(ctx, state) {
			break
		}
		state = conn.GetState()
		if state == connectivity.Ready {
			return "UP"
		}
		if state == connectivity.TransientFailure || state == connectivity.Shutdown {
			break
		}
	}

	return "DOWN"
}
