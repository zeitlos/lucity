package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	githubTokenKey = "x-github-token"
)

type githubTokenContextKey struct{}

// UnaryServerInterceptor returns a gRPC server interceptor that validates
// internal ES256 JWTs and extracts user identity from the token claims.
// When a verifier is provided, all calls must include a valid JWT.
func UnaryServerInterceptor(verifier *InternalVerifier) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, err := extractAuth(ctx, verifier)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream interceptor that validates
// internal ES256 JWTs and extracts user identity from the token claims.
// When a verifier is provided, all calls must include a valid JWT.
func StreamServerInterceptor(verifier *InternalVerifier) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := extractAuth(ss.Context(), verifier)
		if err != nil {
			return err
		}
		return handler(srv, &wrappedAuthStream{ServerStream: ss, ctx: ctx})
	}
}

// extractAuth validates a JWT from the authorization header.
// Returns Unauthenticated if the verifier is configured and no valid JWT is present.
func extractAuth(ctx context.Context, verifier *InternalVerifier) (context.Context, error) {
	if verifier == nil {
		return ctx, nil
	}

	md, _ := metadata.FromIncomingContext(ctx)

	authHeader := firstValue(md, "authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ctx, status.Errorf(codes.Unauthenticated, "missing internal authorization token")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	intClaims, err := verifier.Validate(tokenStr)
	if err != nil {
		return ctx, status.Errorf(codes.Unauthenticated, "invalid internal token: %v", err)
	}

	ctx = WithClaims(ctx, &intClaims.Claims)
	ctx = WithActiveWorkspace(ctx, intClaims.Workspace)

	// Extract GitHub token from plain metadata (needed by builder for repo cloning)
	if ghToken := firstValue(md, githubTokenKey); ghToken != "" {
		ctx = WithGitHubToken(ctx, ghToken)
	}

	return ctx, nil
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

// OutgoingContext mints an ES256 JWT containing user claims and workspace,
// and attaches it to outgoing gRPC metadata. Requires an Issuer in the context.
func OutgoingContext(ctx context.Context) context.Context {
	issuer := IssuerFrom(ctx)
	claims := FromContext(ctx)
	if issuer == nil || claims == nil {
		return ctx
	}

	workspace := ActiveWorkspaceFrom(ctx)
	token, err := issuer.MintToken(claims, workspace)
	if err != nil {
		slog.Error("failed to mint internal JWT", "error", err)
		return ctx
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	if ghToken := GitHubTokenFrom(ctx); ghToken != "" {
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
