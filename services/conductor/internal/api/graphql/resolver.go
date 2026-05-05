package graphql

import "github.com/zeitlos/lucity/services/conductor/internal/api/handler"

//go:generate go run github.com/99designs/gqlgen generate

type Resolver struct {
	API *handler.Client
}
