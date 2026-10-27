package auth

import "context"

type tokenContextKey struct{}

type githubTokenContextKey struct{}

// WithToken attaches a raw JWT token to the context.
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, token)
}

// TokenFrom extracts the raw JWT token from the context.
func TokenFrom(ctx context.Context) string {
	token, _ := ctx.Value(tokenContextKey{}).(string)
	return token
}

// WithGitHubToken attaches a GitHub token to the context.
func WithGitHubToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, githubTokenContextKey{}, token)
}

// GitHubTokenFrom extracts the GitHub token from the context.
func GitHubTokenFrom(ctx context.Context) string {
	token, _ := ctx.Value(githubTokenContextKey{}).(string)
	return token
}
