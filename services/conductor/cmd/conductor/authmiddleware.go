package main

import (
	"context"
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
			ctx := auth.WithResponseWriter(r.Context(), w)

			if bearer := bearerToken(r); bearer != "" {
				if claims, err := verifier.ValidateToken(ctx, bearer); err == nil {
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

			claims := &auth.Claims{Subject: data.Sub, Name: data.Name, Email: data.Email, AvatarURL: data.Picture}

			if workspace := r.Header.Get(tenant.Header); workspace != "" {
				if orgToken, err := store.orgToken(ctx, data.SID, workspace); err == nil {
					if tokenClaims, err := verifier.ValidateToken(ctx, orgToken); err == nil {
						claims.Workspaces = tokenClaims.Workspaces
					}
				}
			}

			if account, err := store.accountAPIToken(ctx, data.SID); err == nil {
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

func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
