package tenant

import (
	"context"
	"errors"
)

// Header is the HTTP header used to pass the workspace identifier.
const Header = "X-Lucity-Workspace"

// MetadataKey is the gRPC metadata key for workspace propagation (lowercase per gRPC convention).
const MetadataKey = "x-lucity-workspace"

type contextKey struct{}

func NewContext(ctx context.Context, workspace string) context.Context {
	return context.WithValue(ctx, contextKey{}, workspace)
}

func FromContext(ctx context.Context) (string, error) {
	workspace, set := ctx.Value(contextKey{}).(string)

	if !set || workspace == "" {
		return "", errors.New("workspace not specified")
	}

	return workspace, nil
}
