package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

type Server struct {
	server *http.Server
	port   string
}

func NewServer(port string, webhookHandler *stripelib.WebhookHandler) *Server {
	mux := http.NewServeMux()
	mux.Handle("/webhooks/stripe", webhookHandler)
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
	return "Stripe Webhook HTTP"
}

func (s *Server) Start() error {
	slog.Info("stripe webhook server listening", "url", fmt.Sprintf("http://localhost:%s/webhooks/stripe", s.port))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
