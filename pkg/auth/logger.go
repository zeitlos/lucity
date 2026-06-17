package auth

import (
	"context"
	"log/slog"
)

type authHandler struct {
	slog.Handler
}

func (h authHandler) Handle(ctx context.Context, r slog.Record) error {
	claims, err := FromContext(ctx)

	if err != nil {
		return nil
	}

	user := claims.Email

	if user == "" {
		user = claims.Subject
	}

	r.AddAttrs(slog.String("user", user))

	return h.Handler.Handle(ctx, r)
}

func (h authHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return authHandler{
		Handler: h.Handler.WithAttrs(attrs),
	}
}

func (h authHandler) WithGroup(name string) slog.Handler {
	return authHandler{
		Handler: h.Handler.WithGroup(name),
	}
}
