package graphql

import (
	"context"

	gqlgen "github.com/99designs/gqlgen/graphql"
)

func bucketUsageRequested(ctx context.Context) bool {
	return gqlgen.FieldRequested(ctx, "sizeBytes") || gqlgen.FieldRequested(ctx, "objectCount")
}
