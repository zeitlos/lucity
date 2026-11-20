package auth

import "context"

type tokenContextKey struct{}

type githubTokenContextKey struct{}

// WithToken attaches a raw JWT token to the context.
func WithToken(ctx context.Context, token string) context.Context {
	return WithTokenSource(ctx, func(context.Context) string { return token })
}

// WithTokenSource attaches a function that obtains the token only when
// TokenFrom asks for it.
func WithTokenSource(ctx context.Context, source func(context.Context) string) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, source)
}

// TokenFrom returns the context's token, or an empty string if it has none.
func TokenFrom(ctx context.Context) string {
	source, ok := ctx.Value(tokenContextKey{}).(func(context.Context) string)
	if !ok {
		return ""
	}
	return source(ctx)
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
