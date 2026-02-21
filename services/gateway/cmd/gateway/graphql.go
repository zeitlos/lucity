package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	gatewaygraphql "github.com/zeitlos/lucity/services/gateway/graphql"
	"github.com/zeitlos/lucity/services/gateway/graphql/directive"
	"github.com/zeitlos/lucity/services/gateway/graphql/model"
	"github.com/zeitlos/lucity/services/gateway/handler"

	"github.com/zeitlos/lucity/pkg/auth"
	gh "github.com/zeitlos/lucity/pkg/github"

	gqlgen "github.com/99designs/gqlgen/graphql"
	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/ast"
)

type GraphQLServer struct {
	server *http.Server
	port   string
}

func NewGraphQLServer(port string, api *handler.Client, githubApp *gh.App, jwtSecret, dashboardURL string) *GraphQLServer {
	resolver := gatewaygraphql.Resolver{
		API: api,
	}

	constraintDir := directive.New()

	srv := gqlhandler.New(gatewaygraphql.NewExecutableSchema(gatewaygraphql.Config{
		Resolvers: &resolver,
		Directives: gatewaygraphql.DirectiveRoot{
			Constraint: constraintDir.Validate,
			HasRole: func(ctx context.Context, obj interface{}, next gqlgen.Resolver, role []model.Role) (interface{}, error) {
				claims := auth.FromContext(ctx)

				// Allow ANONYMOUS access
				for _, r := range role {
					if r == model.RoleAnonymous {
						return next(ctx)
					}
				}

				if claims == nil {
					return nil, fmt.Errorf("unauthorized")
				}

				for _, required := range role {
					if claims.HasRole(auth.Role(required)) {
						return next(ctx)
					}
				}

				return nil, fmt.Errorf("forbidden: insufficient role")
			},
		},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin:  func(r *http.Request) bool { return true },
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
			// Extract JWT from connectionParams for WebSocket auth.
			token, _ := initPayload["Authorization"].(string)
			if token == "" {
				return ctx, &initPayload, nil
			}
			token = strings.TrimPrefix(token, "Bearer ")
			claims, err := auth.ParseToken(token, jwtSecret)
			if err != nil {
				return ctx, &initPayload, nil
			}
			ctx = auth.WithClaims(ctx, claims)
			ctx = auth.WithToken(ctx, token)
			return ctx, &initPayload, nil
		},
	})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})

	mux := http.NewServeMux()

	// Health endpoints
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

	// Auth endpoints
	registerAuthRoutes(mux, githubApp, jwtSecret, dashboardURL)

	// REST API endpoints
	mux.HandleFunc("/api/eject/", ejectHandler(api))

	// GraphQL endpoints
	mux.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	mux.Handle("/graphql", srv)

	// Apply auth middleware then CORS
	authMiddleware := auth.Middleware(jwtSecret)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", dashboardURL},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	return &GraphQLServer{
		port: port,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: corsHandler.Handler(authMiddleware(mux)),
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
