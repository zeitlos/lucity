package graphql

import (
	"github.com/zeitlos/lucity/services/cockpit/internal/cockpit"
)

//go:generate go run github.com/99designs/gqlgen generate

type Resolver struct {
	Cockpit *cockpit.Client
}
