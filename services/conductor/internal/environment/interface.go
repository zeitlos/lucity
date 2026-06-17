package environment

import (
	"context"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Interface interface {
	Ensure(ctx context.Context, id platform.EnvironmentID, tier platform.ResourceTier) error
	Delete(ctx context.Context, id platform.EnvironmentID) error
}
