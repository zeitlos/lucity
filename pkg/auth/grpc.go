package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authHeader = "authorization"

// UnaryServerInterceptor returns a gRPC server interceptor that extracts
// and validates JWT tokens from the "authorization" metadata key.
// Valid claims are attached to the request context via WithClaims.
func UnaryServerInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get(authHeader)
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		token := values[0]
		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := ParseToken(token, jwtSecret)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx = WithClaims(ctx, claims)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream interceptor that extracts
// and validates JWT tokens from the "authorization" metadata key.
func StreamServerInterceptor(jwtSecret string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get(authHeader)
		if len(values) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization token")
		}

		token := values[0]
		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := ParseToken(token, jwtSecret)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx = WithClaims(ctx, claims)
		return handler(srv, &wrappedAuthStream{ServerStream: ss, ctx: ctx})
	}
}

type wrappedAuthStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *wrappedAuthStream) Context() context.Context {
	return s.ctx
}

// TokenFromContext extracts the raw JWT token string from the HTTP request context.
// This is used by the gateway to propagate the token to gRPC calls.
type tokenContextKey struct{}

// WithToken attaches a raw JWT token to the context.
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, token)
}

// TokenFrom extracts the raw JWT token from the context.
func TokenFrom(ctx context.Context) string {
	token, _ := ctx.Value(tokenContextKey{}).(string)
	return token
}

// OutgoingCredentials returns a grpc.CallOption that attaches the JWT token
// from the context as gRPC metadata for outgoing calls.
func OutgoingContext(ctx context.Context) context.Context {
	token := TokenFrom(ctx)
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, authHeader, token)
}
