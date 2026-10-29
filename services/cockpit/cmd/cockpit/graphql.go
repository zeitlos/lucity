package main

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/zeitlos/lucity/services/cockpit/internal/api/graphql"
	"github.com/zeitlos/lucity/services/cockpit/internal/api/graphql/directive"
	"github.com/zeitlos/lucity/services/cockpit/internal/cockpit"
)

func graphQLHandler(client *cockpit.Client) http.Handler {
	srv := handler.New(graphql.NewExecutableSchema(graphql.Config{
		Resolvers: &graphql.Resolver{Cockpit: client},
		Directives: graphql.DirectiveRoot{
			Constraint: directive.New().Validate,
		},
	}))

	srv.AddTransport(transport.POST{})

	return srv
}
