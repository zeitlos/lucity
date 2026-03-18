package auth

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	githubTokenKey = "x-github-token"
	subjectKey     = "x-lucity-subject"
	emailKey       = "x-lucity-email"
	rolesKey       = "x-lucity-roles"
)

type githubTokenContextKey struct{}

// UnaryServerInterceptor returns a gRPC server interceptor that extracts
// user identity from trusted metadata headers set by the gateway.
// Claims are attached to the request context via WithClaims.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = extractClaims(ctx)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream interceptor that extracts
// user identity from trusted metadata headers set by the gateway.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := extractClaims(ss.Context())
		return handler(srv, &wrappedAuthStream{ServerStream: ss, ctx: ctx})
	}
}

func extractClaims(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	// Extract user identity from metadata headers
	subject := firstValue(md, subjectKey)
	if subject != "" {
		claims := &Claims{
			Subject: subject,
			Email:   firstValue(md, emailKey),
		}
		if rolesStr := firstValue(md, rolesKey); rolesStr != "" {
			for _, r := range strings.Split(rolesStr, ",") {
				claims.Roles = append(claims.Roles, Role(r))
			}
		}
		ctx = WithClaims(ctx, claims)
	}

	// Extract GitHub token if present (used by builder for repo cloning).
	if vals := md.Get(githubTokenKey); len(vals) > 0 {
		ctx = WithGitHubToken(ctx, vals[0])
	}

	return ctx
}

func firstValue(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
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

// OutgoingContext propagates user identity and GitHub token from the context
// as gRPC metadata for outgoing calls to backend services.
func OutgoingContext(ctx context.Context) context.Context {
	// Propagate user identity as plain metadata headers
	if claims := FromContext(ctx); claims != nil {
		ctx = metadata.AppendToOutgoingContext(ctx, subjectKey, claims.Subject)
		if claims.Email != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, emailKey, claims.Email)
		}
		if len(claims.Roles) > 0 {
			roles := make([]string, len(claims.Roles))
			for i, r := range claims.Roles {
				roles[i] = string(r)
			}
			ctx = metadata.AppendToOutgoingContext(ctx, rolesKey, strings.Join(roles, ","))
		}
	}

	// Propagate GitHub token if present
	ghToken := GitHubTokenFrom(ctx)
	if ghToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, githubTokenKey, ghToken)
	}
	return ctx
}

// WithGitHubToken attaches a GitHub token to the context for gRPC propagation.
// Used to pass installation tokens to the builder for repo cloning.
func WithGitHubToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, githubTokenContextKey{}, token)
}

// GitHubTokenFrom extracts the GitHub token from the context.
func GitHubTokenFrom(ctx context.Context) string {
	token, _ := ctx.Value(githubTokenContextKey{}).(string)
	return token
}

// Refresh token context helpers

type refreshTokenContextKey struct{}

// WithRefreshToken attaches a refresh token to the context.
func WithRefreshToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, refreshTokenContextKey{}, token)
}

// RefreshTokenFrom extracts the refresh token from the context.
func RefreshTokenFrom(ctx context.Context) string {
	token, _ := ctx.Value(refreshTokenContextKey{}).(string)
	return token
}

// ResponseWriter context helpers

type responseWriterContextKey struct{}

// WithResponseWriter attaches an http.ResponseWriter to the context.
func WithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, responseWriterContextKey{}, w)
}

// ResponseWriterFrom extracts the http.ResponseWriter from the context.
func ResponseWriterFrom(ctx context.Context) http.ResponseWriter {
	w, _ := ctx.Value(responseWriterContextKey{}).(http.ResponseWriter)
	return w
}
