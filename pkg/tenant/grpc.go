package tenant

import (
	"context"

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
// the workspace identifier from incoming metadata and attaches it to the context.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = extractWorkspace(ctx)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream interceptor that extracts
// the workspace identifier from incoming metadata and attaches it to the context.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := extractWorkspace(ss.Context())
		return handler(srv, &wrappedStream{ServerStream: ss, ctx: ctx})
	}
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
