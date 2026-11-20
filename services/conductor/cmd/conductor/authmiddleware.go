package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/session"
	"github.com/zeitlos/lucity/pkg/tenant"
)

const accountTokenHeader = "X-Lucity-Account-Token"

type sessionContextKey struct{}

func withSession(ctx context.Context, data sessionData) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, data)
}

func sessionFromContext(ctx context.Context) (sessionData, bool) {
	data, ok := ctx.Value(sessionContextKey{}).(sessionData)
	return data, ok
}

func sessionAuth(tokens *sessionTokens, codec *session.Codec, verifier *auth.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if bearer := bearerToken(r); bearer != "" {
				claims, err := verifier.ValidateToken(ctx, bearer)
				if errors.Is(err, auth.ErrOrganizationLookup) {
					slog.WarnContext(ctx, "bearer token could not be checked", "error", err)
					writeGraphQLError(w, http.StatusServiceUnavailable, "could not verify your sign-in right now, retry in a moment; if it keeps failing, sign in again")
					return
				}
				if err != nil {
					slog.WarnContext(ctx, "bearer token rejected", "error", err)
				} else {
					ctx = auth.NewContext(ctx, claims)
					if account := r.Header.Get(accountTokenHeader); account != "" {
						ctx = auth.WithToken(ctx, account)
					}
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			var data sessionData
			if err := codec.Load(r, &data); err != nil {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			ctx = withSession(ctx, data)

			signOut := func(err error) {
				slog.InfoContext(ctx, "session ended; signing out", "subject", data.Sub, "error", err)
				codec.Clear(w)
				next.ServeHTTP(w, r)
			}

			claims := &auth.Claims{Subject: data.Sub, Name: data.Name, Email: data.Email, AvatarURL: data.Picture}

			if workspace := r.Header.Get(tenant.Header); workspace != "" {
				orgToken, err := tokens.workspaceToken(ctx, data, workspace)
				if errors.Is(err, errNoSession) {
					signOut(err)
					return
				}

				if err != nil {
					slog.WarnContext(ctx, "organization token unavailable; request will have no workspace", "workspace", workspace, "subject", data.Sub, "error", err)
				} else if tokenClaims, err := verifier.ValidateToken(ctx, orgToken); err != nil {
					slog.WarnContext(ctx, "organization token rejected; request will have no workspace", "workspace", workspace, "subject", data.Sub, "error", err)
				} else {
					claims.Workspaces = tokenClaims.Workspaces
				}
			}

			if err := tokens.check(ctx, data); errors.Is(err, errNoSession) {
				signOut(err)
				return
			}

			ctx = auth.WithTokenSource(ctx, func(ctx context.Context) string {
				account, err := tokens.accountToken(ctx, data)
				if err != nil {
					slog.WarnContext(ctx, "account token unavailable", "subject", data.Sub, "error", err)
					return ""
				}
				return account
			})

			ctx = auth.NewContext(ctx, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeGraphQLError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"errors": []map[string]string{{"message": message}}})
}

func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
