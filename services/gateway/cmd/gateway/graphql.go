package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	gatewaygraphql "github.com/zeitlos/lucity/services/gateway/graphql"
	"github.com/zeitlos/lucity/services/gateway/graphql/directive"
	"github.com/zeitlos/lucity/services/gateway/graphql/model"
	"github.com/zeitlos/lucity/services/gateway/handler"

	gqlgen "github.com/99designs/gqlgen/graphql"
	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/ast"
)

type GraphQLServer struct {
	server *http.Server
	port   string
}

func NewGraphQLServer(port string, api *handler.Client) *GraphQLServer {
	resolver := gatewaygraphql.Resolver{
		API: api,
	}

	constraintDir := directive.New()

	srv := gqlhandler.New(gatewaygraphql.NewExecutableSchema(gatewaygraphql.Config{
		Resolvers: &resolver,
		Directives: gatewaygraphql.DirectiveRoot{
			Constraint: constraintDir.Validate,
			HasRole: func(ctx context.Context, obj interface{}, next gqlgen.Resolver, role []model.Role) (interface{}, error) {
				// No-op: auth will be enforced at operation level in a future phase
				return next(ctx)
			},
		},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	mux.Handle("/graphql", srv)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	return &GraphQLServer{
		port: port,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: corsHandler.Handler(mux),
		},
	}
}

func (s *GraphQLServer) Start() error {
	slog.Info("graphql playground enabled", "url", fmt.Sprintf("http://localhost:%s/playground", s.port))
	return s.server.ListenAndServe()
}

func (s *GraphQLServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *GraphQLServer) Label() string {
	return "GraphQL"
}
