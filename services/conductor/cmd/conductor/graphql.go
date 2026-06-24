package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/logto"
	gatewaygraphql "github.com/zeitlos/lucity/services/conductor/internal/api/graphql"
	"github.com/zeitlos/lucity/services/conductor/internal/api/graphql/directive"
	"github.com/zeitlos/lucity/services/conductor/internal/api/graphql/model"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/tenant"

	"github.com/99designs/gqlgen/graphql"
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

const (
	hasRoleDirective        = "hasRole"
	allowSuspendedDirective = "allowSuspended"
)

func NewGraphQLServer(port string, conductorClient *conductor.Client, oidcProvider *OIDCProvider, verifier *auth.Verifier, logtoClient *logto.Client, internalIssuer *auth.Issuer, sessionSecret, dashboardURL, githubAppSlug string, grpcComponents []grpcComponent) *GraphQLServer {
	resolver := gatewaygraphql.Resolver{
		Conductor: conductorClient,
	}

	constraintDir := directive.New()

	srv := gqlhandler.New(gatewaygraphql.NewExecutableSchema(gatewaygraphql.Config{
		Resolvers: &resolver,
		Directives: gatewaygraphql.DirectiveRoot{
			Constraint: constraintDir.Validate,
			AllowSuspended: func(ctx context.Context, obj interface{}, next gqlgen.Resolver) (interface{}, error) {
				return next(ctx)
			},
			HasRole: func(ctx context.Context, obj interface{}, next gqlgen.Resolver, required model.Role) (interface{}, error) {
				if required == model.RoleAnonymous {
					return next(ctx)
				}

				claims, err := auth.FromContext(ctx)

				if err != nil {
					return nil, err
				}

				if required == model.RoleAuthenticated {
					return next(ctx)
				}

				converted, err := convertRole(required)

				if err != nil {
					return nil, err
				}

				workspace, err := tenant.FromContext(ctx)

				if err != nil {
					return nil, err
				}

				if !claims.WorkspaceRoleIn(workspace).Satisfies(converted) {
					return nil, fmt.Errorf("forbidden: insufficient role")
				}

				oc := gqlgen.GetOperationContext(ctx)

				// Check workspace suspension for mutations (queries are never blocked).
				if oc.Operation == nil || oc.Operation.Operation != ast.Mutation {
					return next(ctx)
				}

				fc := gqlgen.GetFieldContext(ctx)
				allowSuspended := false

				if fc != nil && fc.Field.Definition != nil {
					for _, d := range fc.Field.Definition.Directives {
						if d.Name == allowSuspendedDirective {
							// Skip if the field is marked @allowSuspended (checked via AST, not directive order).
							allowSuspended = true
							break
						}
					}
				}

				if !allowSuspended {
					org, err := logtoClient.OrganizationByName(ctx, workspace)

					if err != nil {
						return nil, err
					}

					suspended := false

					if raw, exists := org.CustomData["suspended"]; exists {
						if v, ok := raw.(bool); ok {
							suspended = v
						} else {
							// key is present but value malformed
							suspended = true
						}
					}

					if suspended {
						slog.Warn("mutation blocked: workspace suspended", "workspace", workspace, "operation", oc.OperationName)

						return nil, fmt.Errorf("workspace suspended: update your payment method to continue")
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
		dashboardURL:            true,
	}
	// The browser sends the origin without path, so also allow the base URL.
	if u, err := url.Parse(dashboardURL); err == nil {
		allowedOrigins[u.Scheme+"://"+u.Host] = true
	}

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		PingPongInterval:      15 * time.Second,
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
					ctx = auth.NewContext(ctx, claims)
					ctx = auth.WithToken(ctx, token)
				}
			}

			// Inject internal JWT issuer for gRPC calls made during subscriptions.
			if internalIssuer != nil {
				ctx = auth.WithIssuer(ctx, internalIssuer)
			}

			// Workspace: browser can't send custom headers on WS upgrade,
			// so read from connectionParams.
			if ws, ok := initPayload[tenant.Header].(string); ok && ws != "" {
				ctx = tenant.NewContext(ctx, ws)
				ctx = auth.WithActiveWorkspace(ctx, ws)
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

		if oc.Operation == nil || oc.Operation.Operation != ast.Mutation {
			return next(ctx)
		}

		slog.InfoContext(ctx, "graphql mutation", "operation", oc.OperationName)

		return next(ctx)
	})

	// Enforcing IDs stay within workspace provided by header.
	srv.AroundFields(func(ctx context.Context, next graphql.Resolver) (any, error) {
		fc := graphql.GetFieldContext(ctx)

		for _, arg := range fc.Args {
			err := walkWorkspaceScoped(reflect.ValueOf(arg), func(scoped platform.WorkspaceScoped) error {
				callerWorkspace, err := tenant.FromContext(ctx)

				if err != nil {
					return err
				}

				if scoped.WorkspaceID() != callerWorkspace {
					slog.WarnContext(ctx, "workspaces of id and header don't match", "id", scoped.WorkspaceID(), "header", callerWorkspace)

					return errors.New("not found")
				}

				return nil
			})

			if err != nil {
				return nil, err
			}
		}

		return next(ctx)
	})

	srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		var dbProv *conductor.DatabaseProvisioningError
		if errors.As(err, &dbProv) {
			return &gqlerror.Error{
				Message:    "Database is provisioning",
				Extensions: map[string]interface{}{"code": "DATABASE_PROVISIONING"},
			}
		}

		slog.ErrorContext(ctx, "error during graphql operation", "error", err)

		return gqlgen.DefaultErrorPresenter(ctx, err)
	})

	srv.AroundRootFields(func(ctx context.Context, next graphql.RootResolver) graphql.Marshaler {
		rc := graphql.GetRootFieldContext(ctx)

		if rc == nil {
			return next(ctx)
		}

		if strings.HasPrefix(rc.Field.Name, "__") {
			return next(ctx)
		}

		def := rc.Field.Definition

		if def == nil || def.Directives.ForName(hasRoleDirective) == nil {
			slog.ErrorContext(ctx, "root field has no auth directive", "type", rc.Object, "field", rc.Field.Name)

			graphql.AddError(ctx, gqlerror.Errorf("%s.%s has no auth directive", rc.Object, rc.Field.Name))

			return graphql.Null
		}

		return next(ctx)
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
	registerAuthRoutes(mux, oidcProvider, conductorClient, logtoClient, sessionSecret, dashboardURL, githubAppSlug)

	// GraphQL endpoints
	mux.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	mux.Handle("/graphql", srv)

	authMiddleware := auth.Middleware(verifier)

	// Issuer middleware injects the internal JWT issuer into the request context.
	// This enables auth.OutgoingContext to mint ES256 JWTs for gRPC calls.
	issuerMiddleware := func(next http.Handler) http.Handler {
		if internalIssuer == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := auth.WithIssuer(r.Context(), internalIssuer)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// TODO: Validate if CORS is even needed. Dashboard and conductor run under the same hostname.
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
					issuerMiddleware(
						tenant.Middleware(
							tenant.AuthorizeMiddleware(mux),
						),
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

func walkWorkspaceScoped(v reflect.Value, visit func(platform.WorkspaceScoped) error) error {
	if !v.IsValid() || !v.CanInterface() {
		return nil
	}

	if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) && v.IsNil() {
		return nil
	}

	if scoped, ok := v.Interface().(platform.WorkspaceScoped); ok {
		return visit(scoped)
	}

	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		return walkWorkspaceScoped(v.Elem(), visit)
	case reflect.Struct:
		for i := range v.NumField() {
			if v.Type().Field(i).PkgPath != "" {
				continue
			}

			if err := walkWorkspaceScoped(v.Field(i), visit); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			if err := walkWorkspaceScoped(v.Index(i), visit); err != nil {
				return err
			}
		}
	}

	return nil
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

func convertRole(role model.Role) (auth.WorkspaceRole, error) {
	switch role {
	case model.RoleWorkspaceAdmin:
		return auth.WorkspaceRoleAdmin, nil
	case model.RoleWorkspaceMember:
		return auth.WorkspaceRoleUser, nil
	}

	return "", fmt.Errorf("unknown role: %q", role)
}
