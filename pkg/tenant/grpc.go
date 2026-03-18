package tenant

import (
	"context"

	"github.com/zeitlos/lucity/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// OutgoingContext appends the workspace identifier from the context to
// outgoing gRPC metadata. If no workspace is set, the context is returned unchanged.
func OutgoingContext(ctx context.Context) context.Context {
	ws := FromContext(ctx)
	if ws == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, MetadataKey, ws)
}

// UnaryServerInterceptor returns a gRPC server interceptor that extracts
// the workspace identifier from the auth context (JWT-validated) or incoming metadata.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = resolveWorkspace(ctx)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream interceptor that extracts
// the workspace identifier from the auth context (JWT-validated) or incoming metadata.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := resolveWorkspace(ss.Context())
		return handler(srv, &wrappedStream{ServerStream: ss, ctx: ctx})
	}
}

// resolveWorkspace reads workspace from auth context (set by JWT validation) first,
// falling back to plain metadata extraction for legacy callers.
func resolveWorkspace(ctx context.Context) context.Context {
	if ws := auth.ActiveWorkspaceFrom(ctx); ws != "" {
		return WithWorkspace(ctx, ws)
	}
	return extractWorkspace(ctx)
}

func extractWorkspace(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	values := md.Get(MetadataKey)
	if len(values) == 0 {
		return ctx
	}
	return WithWorkspace(ctx, values[0])
}

// wrappedStream overrides Context() to return our enriched context.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
