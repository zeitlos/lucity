package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/zeitlos/lucity/services/webhook/github"
)

type Server struct {
	server *http.Server
	port   string
}

func NewServer(port, webhookSecret string) *Server {
	secret := []byte(webhookSecret)
	mux := http.NewServeMux()

	mux.HandleFunc("/webhook/github", handleGitHub(secret))

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

func handleGitHub(secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		event, err := github.ValidateAndParse(secret, r)
		if err != nil {
			slog.Warn("webhook validation failed", "error", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		slog.Info("webhook received",
			"type", event.Type,
			"action", event.Action,
			"repo", event.RepoFullName,
			"ref", event.Ref,
			"sha", event.CommitSHA,
			"sender", event.Sender,
		)

		// TODO: route events to builder/packager/deployer
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received": true}`))
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
