package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
)

const githubActionsIssuer = "https://token.actions.githubusercontent.com"

type githubActionsVerifier struct {
	audience string

	mu       sync.Mutex
	verifier *oidc.IDTokenVerifier
}

func newGitHubActionsVerifier(audience string) *githubActionsVerifier {
	return &githubActionsVerifier{audience: audience}
}

func originFromURL(raw string) string {
	parsed, err := url.Parse(raw)

	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return raw
	}

	return parsed.Scheme + "://" + parsed.Host
}

type githubActionsClaims struct {
	Repository string `json:"repository"`
	Ref        string `json:"ref"`
	SHA        string `json:"sha"`
	Workflow   string `json:"workflow"`
	RunID      string `json:"run_id"`
	Actor      string `json:"actor"`
	EventName  string `json:"event_name"`
}

func (v *githubActionsVerifier) verify(ctx context.Context, rawToken string) (*githubActionsClaims, error) {
	verifier, err := v.idTokenVerifier(ctx)

	if err != nil {
		return nil, err
	}

	idToken, err := verifier.Verify(ctx, rawToken)

	if err != nil {
		return nil, err
	}

	var claims githubActionsClaims

	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (v *githubActionsVerifier) idTokenVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.verifier != nil {
		return v.verifier, nil
	}

	provider, err := oidc.NewProvider(ctx, githubActionsIssuer)

	if err != nil {
		return nil, err
	}

	v.verifier = provider.Verifier(&oidc.Config{ClientID: v.audience})

	return v.verifier, nil
}

func handleCIExchange(verifier *githubActionsVerifier, conductorClient *conductor.Client, sessionSecret string, ttl time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 16<<10)).Decode(&body); err != nil || body.Token == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}

		claims, err := verifier.verify(r.Context(), body.Token)
		if err != nil {
			slog.WarnContext(r.Context(), "ci exchange: token verification failed", "error", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if claims.Repository == "" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		match, err := conductorClient.MatchCIDeploy(r.Context(), claims.Repository, claims.Ref)
		if err != nil {
			slog.WarnContext(r.Context(), "ci exchange: no matching service",
				"repo", claims.Repository, "ref", claims.Ref, "error", err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		memberships := make([]auth.WorkspaceMembership, 0, len(match.Workspaces))
		for _, workspace := range match.Workspaces {
			memberships = append(memberships, auth.WorkspaceMembership{
				Workspace: workspace,
				Role:      auth.WorkspaceRoleDeployer,
			})
		}

		token, err := mintMachineToken(sessionSecret, &auth.Claims{
			Subject:    conductor.GitHubActionsSubject(claims.Repository),
			Name:       "GitHub Actions",
			Workspaces: memberships,
		}, ttl)
		if err != nil {
			slog.ErrorContext(r.Context(), "ci exchange: mint token failed", "error", err)
			http.Error(w, "token issue failed", http.StatusInternalServerError)
			return
		}

		services := make([]string, len(match.Services))
		for i, id := range match.Services {
			services[i] = id.String()
		}

		slog.InfoContext(r.Context(), "ci deploy token issued",
			"repo", claims.Repository,
			"ref", claims.Ref,
			"sha", claims.SHA,
			"workflow", claims.Workflow,
			"run_id", claims.RunID,
			"actor", claims.Actor,
			"services", services,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"token":     token,
			"expiresAt": time.Now().Add(ttl).Format(time.RFC3339),
			"workspace": match.Workspaces[0],
			"services":  services,
		})
	}
}
