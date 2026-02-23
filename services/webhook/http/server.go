package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/packager"
	webhook "github.com/zeitlos/lucity/services/webhook"
	"github.com/zeitlos/lucity/services/webhook/github"
)

// Handler holds the dependencies for webhook event processing.
type Handler struct {
	GitHubApp *ghpkg.App
	Pipeline  *webhook.Pipeline
	JWTSecret string
}

type Server struct {
	server *http.Server
	port   string
}

func NewServer(port, webhookSecret string, handler *Handler) *Server {
	secret := []byte(webhookSecret)
	mux := http.NewServeMux()

	mux.HandleFunc("/webhook/github", handleGitHub(secret, handler))

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

// handlePush processes a push event: matches repos to projects and triggers builds.
func (h *Handler) handlePush(event *github.Event) {
	// Only deploy pushes to the default branch → development environment.
	refBranch := strings.TrimPrefix(event.Ref, "refs/heads/")
	if refBranch != event.DefaultBranch {
		slog.Debug("push: ignoring non-default branch", "ref", event.Ref, "default", event.DefaultBranch)
		return
	}

	if event.InstallationID == 0 {
		slog.Warn("push: no installation ID in event, cannot mint token")
		return
	}

	// Mint an installation token for GitHub API access.
	ctx := context.Background()
	ghToken, err := h.GitHubApp.InstallationToken(ctx, event.InstallationID)
	if err != nil {
		slog.Error("push: failed to get installation token", "error", err)
		return
	}

	// Extract the owner from the repo full name (e.g., "zeitlos/myapp" → "zeitlos").
	owner := event.RepoFullName
	if i := strings.IndexByte(owner, '/'); i >= 0 {
		owner = owner[:i]
	}

	// Create a JWT for gRPC auth so downstream services accept the request.
	claims := &auth.Claims{
		Subject:     "webhook",
		GitHubLogin: owner,
		Roles:       []auth.Role{auth.RoleUser},
		GitHubToken: ghToken,
	}
	jwt, err := auth.NewToken(claims, h.JWTSecret, 30*time.Minute)
	if err != nil {
		slog.Error("push: failed to create JWT", "error", err)
		return
	}

	ctx = auth.WithToken(ctx, jwt)
	ctx = auth.OutgoingContext(ctx)

	// List all projects and find services matching this repo.
	resp, err := h.Pipeline.Packager.ListProjects(ctx, &packager.ListProjectsRequest{})
	if err != nil {
		slog.Error("push: failed to list projects", "error", err)
		return
	}

	repoURL := fmt.Sprintf("https://github.com/%s", event.RepoFullName)
	environment := "development"

	for _, proj := range resp.Projects {
		for _, svc := range proj.Services {
			if !matchesRepo(svc.SourceUrl, repoURL) {
				continue
			}

			slog.Info("push: triggering deploy",
				"project", proj.Name,
				"service", svc.Name,
				"environment", environment,
				"sha", event.CommitSHA,
			)

			go h.Pipeline.Run(ctx, proj.Name, svc.Name, environment, event.CommitSHA, svc.SourceUrl, svc.ContextPath)
		}
	}
}

// matchesRepo checks if a service's source URL matches a repo URL.
// Handles trailing .git, case differences, and protocol variations.
func matchesRepo(serviceURL, repoURL string) bool {
	normalize := func(u string) string {
		u = strings.TrimSuffix(u, ".git")
		u = strings.TrimSuffix(u, "/")
		u = strings.ToLower(u)
		return u
	}
	return normalize(serviceURL) == normalize(repoURL)
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
