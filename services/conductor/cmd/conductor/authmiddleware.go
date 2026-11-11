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

const (
	sessionCookieMaxAge = 30 * 24 * 3600
	accountTokenHeader  = "X-Lucity-Account-Token"
)

type sessionIDContextKey struct{}

func withSessionID(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, sessionIDContextKey{}, sid)
}

func sessionIDFromContext(ctx context.Context) (string, bool) {
	sid, ok := ctx.Value(sessionIDContextKey{}).(string)
	return sid, ok
}

func sessionAuth(store *sessionStore, codec *session.Codec, verifier *auth.Verifier) func(http.Handler) http.Handler {
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
			store.ensure(ctx, data.SID, data.RefreshToken)
			ctx = withSessionID(ctx, data.SID)

			signOut := func(err error) {
				slog.InfoContext(ctx, "session ended; signing out", "subject", data.Sub, "error", err)
				codec.Clear(w)
				next.ServeHTTP(w, r)
			}

			claims := &auth.Claims{Subject: data.Sub, Name: data.Name, Email: data.Email, AvatarURL: data.Picture}

			if workspace := r.Header.Get(tenant.Header); workspace != "" {
				orgToken, err := store.orgToken(ctx, data.SID, workspace)
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

			account, err := store.accountAPIToken(ctx, data.SID)
			if errors.Is(err, errNoSession) {
				signOut(err)
				return
			}
			if err == nil {
				ctx = auth.WithToken(ctx, account)
			}

			if rotated, ok := store.refreshToken(ctx, data.SID); ok && rotated != data.RefreshToken {
				data.RefreshToken = rotated
				_ = codec.Save(w, data)
			}

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
