package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	gatewaygraphql "github.com/zeitlos/lucity/services/gateway/graphql"
	"github.com/zeitlos/lucity/services/gateway/graphql/directive"
	"github.com/zeitlos/lucity/services/gateway/graphql/model"
	"github.com/zeitlos/lucity/services/gateway/handler"
	"github.com/zeitlos/lucity/services/gateway/logto"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/tenant"

	gqlgen "github.com/99designs/gqlgen/graphql"
	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type GraphQLServer struct {
	server *http.Server
	port   string
}

func NewGraphQLServer(port string, api *handler.Client, oidcProvider *OIDCProvider, verifier *auth.Verifier, logtoClient *logto.Client, sessionSecret, dashboardURL, githubAppSlug string, grpcComponents []grpcComponent) *GraphQLServer {
	resolver := gatewaygraphql.Resolver{
		API: api,
	}

	constraintDir := directive.New()

	type allowSuspendedKeyType struct{}
	allowSuspendedKey := allowSuspendedKeyType{}

	srv := gqlhandler.New(gatewaygraphql.NewExecutableSchema(gatewaygraphql.Config{
		Resolvers: &resolver,
		Directives: gatewaygraphql.DirectiveRoot{
			Constraint: constraintDir.Validate,
			AllowSuspended: func(ctx context.Context, obj interface{}, next gqlgen.Resolver) (interface{}, error) {
				ctx = context.WithValue(ctx, allowSuspendedKey, true)
				return next(ctx)
			},
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

				hasRole := false
				for _, required := range role {
					if claims.HasRole(auth.Role(required)) {
						hasRole = true
						break
					}
				}
				if !hasRole {
					return nil, fmt.Errorf("forbidden: insufficient role")
				}

				// Check workspace suspension for mutations (queries are never blocked).
				oc := gqlgen.GetOperationContext(ctx)
				if oc.Operation != nil && oc.Operation.Operation == ast.Mutation {
					if ctx.Value(allowSuspendedKey) == nil {
						ws := tenant.FromContext(ctx)
						if ws != "" && api.Logto != nil {
							org, err := api.Logto.Organization(ctx, ws)
							if err == nil && org.CustomData != nil {
								if suspended, ok := org.CustomData["suspended"].(bool); ok && suspended {
									slog.Warn("mutation blocked: workspace suspended", "workspace", ws, "operation", oc.OperationName)
									return nil, fmt.Errorf("workspace suspended: update your payment method to continue")
								}
							}
						}
					}
				}

				return next(ctx)
			},
		},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})

	allowedOrigins := map[string]bool{
		"http://localhost:5173": true,
		dashboardURL:           true,
	}
	// The browser sends the origin without path, so also allow the base URL.
	if u, err := url.Parse(dashboardURL); err == nil {
		allowedOrigins[u.Scheme+"://"+u.Host] = true
	}

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return allowedOrigins[origin]
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
			// Auth: prefer connectionParams token (non-browser clients),
			// fall back to session cookie on the HTTP upgrade request
			// (already in ctx from auth.Middleware).
			token, _ := initPayload["Authorization"].(string)
			if token != "" {
				token = strings.TrimPrefix(token, "Bearer ")
				if claims, err := verifier.ValidateToken(ctx, token); err == nil {
					ctx = auth.WithClaims(ctx, claims)
					ctx = auth.WithToken(ctx, token)
				}
			}

			// Workspace: browser can't send custom headers on WS upgrade,
			// so read from connectionParams.
			if ws, ok := initPayload[tenant.Header].(string); ok && ws != "" {
				ctx = tenant.WithWorkspace(ctx, ws)
			}

			return ctx, &initPayload, nil
		},
	})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.FixedComplexityLimit(200))

	// Audit logging for mutations
	srv.AroundOperations(func(ctx context.Context, next gqlgen.OperationHandler) gqlgen.ResponseHandler {
		oc := gqlgen.GetOperationContext(ctx)
		if oc.Operation != nil && oc.Operation.Operation == ast.Mutation {
			claims := auth.FromContext(ctx)
			email := "anonymous"
			if claims != nil {
				email = claims.Email
			}
			workspace := tenant.FromContext(ctx)
			slog.Info("graphql mutation",
				"operation", oc.OperationName,
				"user", email,
				"workspace", workspace,
			)
		}
		return next(ctx)
	})

	srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		var dbProv *handler.DatabaseProvisioningError
		if errors.As(err, &dbProv) {
			return &gqlerror.Error{
				Message:    "Database is provisioning",
				Extensions: map[string]interface{}{"code": "DATABASE_PROVISIONING"},
			}
		}
		return gqlgen.DefaultErrorPresenter(ctx, err)
	})

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

	// Version endpoint
	mux.HandleFunc("/version", versionHandler(grpcComponents))

	// Auth endpoints
	registerAuthRoutes(mux, oidcProvider, api, verifier, logtoClient, sessionSecret, dashboardURL, githubAppSlug)

	// REST API endpoints
	mux.HandleFunc("/api/eject/", ejectHandler(api))

	// GraphQL endpoints
	mux.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	mux.Handle("/graphql", srv)

	// Apply middleware chain: rate limit → CORS → security headers → auth → tenant
	authMiddleware := auth.Middleware(verifier)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", dashboardURL},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", tenant.Header},
		AllowCredentials: true,
	})

	handler := rateLimitMiddleware(
		corsHandler.Handler(
			securityHeadersMiddleware(
				authMiddleware(
					tenant.Middleware(
						tenant.AuthorizeMiddleware(mux),
					),
				),
			),
		),
	)

	return &GraphQLServer{
		port: port,
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
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

// securityHeadersMiddleware adds standard security headers to all responses.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware implements a simple per-IP token bucket rate limiter.
// Each IP gets 100 requests per second with a burst of 200.
func rateLimitMiddleware(next http.Handler) http.Handler {
	type bucket struct {
		tokens   float64
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*bucket)
	)

	const (
		rate      = 100.0 // tokens per second
		burstSize = 200.0
	)

	// Clean up stale entries periodically
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, b := range clients {
				if time.Since(b.lastSeen) > 10*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health checks and version endpoint
		if strings.HasPrefix(r.URL.Path, "/health") || r.URL.Path == "/version" {
			next.ServeHTTP(w, r)
			return
		}

		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = strings.Split(fwd, ",")[0]
			ip = strings.TrimSpace(ip)
		}

		mu.Lock()
		b, exists := clients[ip]
		now := time.Now()
		if !exists {
			b = &bucket{tokens: burstSize, lastSeen: now}
			clients[ip] = b
		}

		// Refill tokens based on elapsed time
		elapsed := now.Sub(b.lastSeen).Seconds()
		b.tokens += elapsed * rate
		if b.tokens > burstSize {
			b.tokens = burstSize
		}
		b.lastSeen = now

		if b.tokens < 1 {
			mu.Unlock()
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		b.tokens--
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
