package tenant

import (
	"context"
	"fmt"
)

// Header is the HTTP header used to pass the workspace identifier.
const Header = "X-Lucity-Workspace"

// MetadataKey is the gRPC metadata key for workspace propagation (lowercase per gRPC convention).
const MetadataKey = "x-lucity-workspace"

type contextKey struct{}

// WithWorkspace attaches a workspace identifier to the context.
func WithWorkspace(ctx context.Context, workspace string) context.Context {
	return context.WithValue(ctx, contextKey{}, workspace)
}

// FromContext extracts the workspace identifier from the context.
// Returns an empty string if no workspace is set.
func FromContext(ctx context.Context) string {
	ws, _ := ctx.Value(contextKey{}).(string)
	return ws
}

// Require extracts the workspace from the context and returns an error if not set.
func Require(ctx context.Context) (string, error) {
	ws := FromContext(ctx)
	if ws == "" {
		return "", fmt.Errorf("missing required %s header", Header)
	}
	return ws, nil
}
