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
	subjectKey     = "x-lucity-subject"
	emailKey       = "x-lucity-email"
	rolesKey       = "x-lucity-roles"
)

type githubTokenContextKey struct{}

// InterceptorOption configures the gRPC server interceptor.
type InterceptorOption func(*interceptorConfig)

type interceptorConfig struct {
	verifier   *InternalVerifier
	requireJWT bool
}

// WithInternalVerifier configures the interceptor to validate internal ES256 JWTs.
func WithInternalVerifier(v *InternalVerifier) InterceptorOption {
	return func(c *interceptorConfig) { c.verifier = v }
}

// WithRequireJWT configures the interceptor to reject calls without a valid JWT.
// When false, unauthenticated calls fall back to legacy plain metadata extraction.
func WithRequireJWT(require bool) InterceptorOption {
	return func(c *interceptorConfig) { c.requireJWT = require }
}

// UnaryServerInterceptor returns a gRPC server interceptor that extracts
// user identity from internal JWTs or (legacy) plain metadata headers.
func UnaryServerInterceptor(opts ...InterceptorOption) grpc.UnaryServerInterceptor {
	cfg := &interceptorConfig{}
	for _, o := range opts {
		o(cfg)
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, err := extractAuth(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream interceptor that extracts
// user identity from internal JWTs or (legacy) plain metadata headers.
func StreamServerInterceptor(opts ...InterceptorOption) grpc.StreamServerInterceptor {
	cfg := &interceptorConfig{}
	for _, o := range opts {
		o(cfg)
	}

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := extractAuth(ss.Context(), cfg)
		if err != nil {
			return err
		}
		return handler(srv, &wrappedAuthStream{ServerStream: ss, ctx: ctx})
	}
}

// extractAuth validates a JWT if a verifier is configured, otherwise falls back to legacy headers.
func extractAuth(ctx context.Context, cfg *interceptorConfig) (context.Context, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	// Try JWT validation first
	if cfg.verifier != nil {
		if authHeader := firstValue(md, "authorization"); strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			intClaims, err := cfg.verifier.Validate(tokenStr)
			if err != nil {
				return ctx, status.Errorf(codes.Unauthenticated, "invalid internal token: %v", err)
			}
			ctx = WithClaims(ctx, &intClaims.Claims)
			ctx = WithActiveWorkspace(ctx, intClaims.Workspace)
			// Extract GitHub token from plain metadata (still needed by builder)
			if ghToken := firstValue(md, githubTokenKey); ghToken != "" {
				ctx = WithGitHubToken(ctx, ghToken)
			}
			return ctx, nil
		}
	}

	// No JWT found
	if cfg.requireJWT {
		return ctx, status.Errorf(codes.Unauthenticated, "missing internal authorization token")
	}

	// Legacy path: extract from plain metadata headers
	if cfg.verifier != nil {
		slog.Warn("gRPC call using legacy plain metadata auth (deprecated)")
	}
	ctx = extractClaims(ctx)
	return ctx, nil
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
// If an Issuer is in the context, mints a cryptographically signed ES256 JWT.
// Otherwise, falls back to plain metadata headers (legacy mode).
func OutgoingContext(ctx context.Context) context.Context {
	claims := FromContext(ctx)

	// Try JWT minting if an issuer is available
	if issuer := IssuerFrom(ctx); issuer != nil && claims != nil {
		workspace := ActiveWorkspaceFrom(ctx)
		token, err := issuer.MintToken(claims, workspace)
		if err != nil {
			slog.Error("failed to mint internal JWT, falling back to legacy metadata", "error", err)
		} else {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
			// Still propagate GitHub token as plain metadata (needed by builder)
			if ghToken := GitHubTokenFrom(ctx); ghToken != "" {
				ctx = metadata.AppendToOutgoingContext(ctx, githubTokenKey, ghToken)
			}
			return ctx
		}
	}

	// Legacy: propagate user identity as plain metadata headers
	if claims != nil {
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

// OutgoingSystemContext is like OutgoingContext but mints a system JWT for service-initiated calls.
// Used by the webhook service for calls that have no user JWT.
func OutgoingSystemContext(ctx context.Context, subject, workspace, scope string) context.Context {
	if issuer := IssuerFrom(ctx); issuer != nil {
		token, err := issuer.MintSystemToken(subject, workspace, scope)
		if err != nil {
			slog.Error("failed to mint system JWT", "error", err)
		} else {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
			if ghToken := GitHubTokenFrom(ctx); ghToken != "" {
				ctx = metadata.AppendToOutgoingContext(ctx, githubTokenKey, ghToken)
			}
			return ctx
		}
	}

	// Legacy fallback
	ctx = metadata.AppendToOutgoingContext(ctx, subjectKey, subject)
	ctx = metadata.AppendToOutgoingContext(ctx, rolesKey, string(RoleUser))
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
