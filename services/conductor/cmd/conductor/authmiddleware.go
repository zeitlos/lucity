package main

import (
	"net/http"
	"strings"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/tenant"
)

const (
	sessionCookieMaxAge = 30 * 24 * 3600
	accountTokenHeader  = "X-Lucity-Account-Token"
)

func sessionAuth(store *sessionStore, verifier *auth.Verifier, secret string, secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := auth.WithResponseWriter(r.Context(), w)

			if bearer := bearerToken(r); bearer != "" {
				if claims, err := verifier.ValidateToken(ctx, bearer); err == nil {
					ctx = auth.NewContext(ctx, claims)
					if acct := r.Header.Get(accountTokenHeader); acct != "" {
						ctx = auth.WithToken(ctx, acct)
					}
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			data, err := openSession(secret, cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			sess := store.get(data)

			claims := &auth.Claims{Subject: data.Sub, Name: data.Name, Email: data.Email, AvatarURL: data.Picture}
			var rotated string

			if workspace := r.Header.Get(tenant.Header); workspace != "" {
				if orgToken, rot, err := store.orgToken(ctx, sess, workspace); err == nil {
					if rot != "" {
						rotated = rot
					}
					if tokenClaims, err := verifier.ValidateToken(ctx, orgToken); err == nil {
						claims.Workspaces = tokenClaims.Workspaces
					}
				}
			}

			if account, rot, err := store.accountAPIToken(ctx, sess); err == nil {
				if rot != "" {
					rotated = rot
				}
				ctx = auth.WithToken(ctx, account)
			}

			if rotated != "" {
				data.RefreshToken = sess.refreshToken
				if sealed, err := sealSession(secret, *data); err == nil {
					setSessionCookie(w, sealed, secure)
				}
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

func setSessionCookie(w http.ResponseWriter, value string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   sessionCookieMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Path: "/", MaxAge: -1})
}
