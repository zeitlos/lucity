package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/services/conductor/internal/api/webhook/github"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

// Handler holds the dependencies for webhook event processing.
type Handler struct {
	GitHubApp *ghpkg.App
	Platform  platform.Interface
	Conductor *conductor.Client
}

type Server struct {
	server *http.Server
	port   string
}

func NewServer(port, webhookSecret string, handler *Handler) *Server {
	secret := []byte(webhookSecret)
	mux := http.NewServeMux()

	mux.HandleFunc("/webhooks/github", handleGitHub(secret, handler))

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

func handleGitHub(secret []byte, h *Handler) http.HandlerFunc {
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

		if event.Type == "push" && h != nil {
			go h.handlePush(event)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received": true}`))
	}
}

func (h *Handler) handlePush(event *github.Event) {
	pushedBranch := strings.TrimPrefix(event.Ref, "refs/heads/")

	if event.InstallationID == 0 {
		slog.Warn("push: no installation ID in event, cannot mint token")
		return
	}

	ctx := context.Background()

	ghToken, err := h.GitHubApp.InstallationToken(ctx, event.InstallationID)

	if err != nil {
		slog.Error("push: failed to mint installation token", "error", err)
		return
	}

	ctx = auth.NewContext(ctx, &auth.Claims{
		Subject: pushActor(event),
	})
	ctx = auth.WithGitHubToken(ctx, ghToken)

	repoURL := fmt.Sprintf("https://github.com/%s", event.RepoFullName)

	matches, err := h.Platform.ServicesByRepo(ctx, repoURL)

	if err != nil {
		slog.Error("push: failed to find services by repo", "repo", repoURL, "error", err)
		return
	}

	matched := 0

	for _, match := range matches {
		if match.InstallationID != 0 && match.InstallationID != event.InstallationID {
			slog.Warn("push: installation mismatch, skipping service",
				"service", match.ID, "expected", match.InstallationID, "got", event.InstallationID)
			continue
		}

		if !match.AutoDeploy {
			continue
		}

		trackedBranch := match.Branch

		if trackedBranch == "" {
			trackedBranch = event.DefaultBranch
		}

		if trackedBranch != pushedBranch {
			continue
		}

		matched++

		slog.Info("push: triggering deploy",
			"workspace", match.ID.Workspace,
			"project", match.ID.Project,
			"service", match.ID.Name,
			"environment", match.ID.Environment,
			"branch", pushedBranch,
			"sha", event.CommitSHA,
		)

		if _, err := h.Conductor.DeployPush(ctx, match.ID, event.CommitSHA); err != nil {
			slog.Warn("push: deploy failed", "service", match.ID, "error", err)
		}
	}

	if matched == 0 {
		slog.Info("push: no matching services", "repo", repoURL, "branch", pushedBranch)
	}
}

func pushActor(event *github.Event) string {
	if event.Sender != "" {
		return event.Sender
	}

	return "webhook"
}

func (s *Server) Label() string {
	return "Webhook HTTP"
}

func (s *Server) Start() error {
	slog.Info("webhook server listening", "url", fmt.Sprintf("http://localhost:%s/webhooks/github", s.port))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
