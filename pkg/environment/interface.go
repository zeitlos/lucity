package environment

import (
	"context"

	"github.com/zeitlos/lucity/pkg/platform"
)

type Interface interface {
	Ensure(ctx context.Context, id platform.EnvironmentID, tier platform.ResourceTier) error
	Delete(ctx context.Context, id platform.EnvironmentID) error
}
