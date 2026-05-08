package graphql

import (
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
)

//go:generate go run github.com/99designs/gqlgen generate

type Resolver struct {
	Conductor *conductor.Client
}
