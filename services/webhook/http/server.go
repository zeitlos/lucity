package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type Server struct {
	server *http.Server
	port   string
}

func NewServer(port string) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/webhook/github", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("received github webhook", "method", r.Method, "path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received": true}`))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return &Server{
		port: port,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		},
	}
}

func (s *Server) Label() string {
	return "Webhook HTTP"
}

func (s *Server) Start() error {
	slog.Info("webhook server listening", "url", fmt.Sprintf("http://localhost:%s/webhook/github", s.port))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
