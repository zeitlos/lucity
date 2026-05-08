package tenant

import (
	"context"
	"log/slog"
)

type tenantHandler struct {
	slog.Handler
}

func (h tenantHandler) Handle(ctx context.Context, r slog.Record) error {
	workspace, err := FromContext(ctx)

	if err != nil {
		return nil
	}

	r.AddAttrs(slog.String("workspace", workspace))

	return h.Handler.Handle(ctx, r)
}

func (h tenantHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return tenantHandler{
		Handler: h.Handler.WithAttrs(attrs),
	}
}

func (h tenantHandler) WithGroup(name string) slog.Handler {
	return tenantHandler{
		Handler: h.Handler.WithGroup(name),
	}
}
