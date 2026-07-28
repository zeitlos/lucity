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

func handleCIExchange(verifier *githubActionsVerifier, conductorClient *conductor.Client, audience string) http.HandlerFunc {
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

		match, err := conductorClient.MatchCIDeploy(r.Context(), claims.Repository)
		if err != nil {
			slog.WarnContext(r.Context(), "ci exchange: no matching service",
				"repo", claims.Repository, "ref", claims.Ref, "error", err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		token, expiresAt, err := conductorClient.IssueCIDeployToken(r.Context(), claims.Repository, match.Workspaces[0], audience)
		if err != nil {
			slog.ErrorContext(r.Context(), "ci exchange: issue token failed", "error", err)
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
			"expiresAt": expiresAt.Format(time.RFC3339),
			"workspace": match.Workspaces[0],
			"services":  services,
		})
	}
}
