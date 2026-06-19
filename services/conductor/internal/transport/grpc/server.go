package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/conductor"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

// Service implements the ConductorService gRPC server — the inbound surface
// the conductor exposes to other platform services (today: cashier). It is
// the only entry point to the deployer / platform clients from outside the
// conductor binary.
type Service struct {
	conductor.UnimplementedConductorServiceServer

	platform platform.Interface
	deployer deployer.Interface
}

func NewService(platform platform.Interface, deployer deployer.Interface) *Service {
	return &Service{
		platform: platform,
		deployer: deployer,
	}
}

// SuspendWorkspace flips the suspended bit on every environment owned by
// the workspace. Each env is its own helm release, so the call fans out one
// helm upgrade per env; failures on a single env are logged and skipped so
// one bad env doesn't block the rest.
func (s *Service) SuspendWorkspace(ctx context.Context, req *conductor.SuspendWorkspaceRequest) (*conductor.SuspendWorkspaceResponse, error) {
	ws := req.GetWorkspace()

	if ws == "" {
		return nil, status.Error(codes.InvalidArgument, "workspace required")
	}

	projects, err := s.platform.Projects(ctx, ws)

	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	for _, project := range projects {
		envs, err := s.platform.Environments(ctx, project.ID)

		if err != nil {
			slog.Warn("suspend: list environments failed", "project", project.ID, "error", err)
			continue
		}

		for _, env := range envs {
			if _, err := s.deployer.Environments().Suspend(ctx, env.ID, req.GetSuspended()); err != nil {
				slog.Warn("suspend: env failed", "env", env.ID, "error", err)
			}
		}
	}

	return &conductor.SuspendWorkspaceResponse{}, nil
}

// ListResourceAllocations enumerates platform-managed namespaces and returns
// the per-env resource usage Cashier needs for metering.
func (s *Service) ListResourceAllocations(ctx context.Context, req *conductor.ListResourceAllocationsRequest) (*conductor.ListResourceAllocationsResponse, error) {
	allocations, err := s.platform.ResourceAllocations(ctx)

	if err != nil {
		return nil, fmt.Errorf("list resource allocations: %w", err)
	}

	out := &conductor.ListResourceAllocationsResponse{
		Allocations: make([]*conductor.ResourceAllocation, 0, len(allocations)),
	}

	for _, allocation := range allocations {
		out.Allocations = append(out.Allocations, &conductor.ResourceAllocation{
			Workspace:     allocation.EnvironmentID.Workspace,
			Project:       allocation.EnvironmentID.Project,
			Environment:   allocation.EnvironmentID.Name,
			Namespace:     allocation.Namespace,
			Tier:          tierToProto(allocation.Tier),
			CpuMillicores: int32(allocation.CPUMillicores),
			MemoryMb:      int32(allocation.MemoryMB),
			DiskMb:        int32(allocation.DiskMB),
		})
	}

	return out, nil
}

func tierToProto(t platform.ResourceTier) conductor.ResourceTier {
	switch t {
	case platform.ProductionTier:
		return conductor.ResourceTier_RESOURCE_TIER_PRODUCTION
	case platform.EcoTier:
		return conductor.ResourceTier_RESOURCE_TIER_ECO
	default:
		return conductor.ResourceTier_RESOURCE_TIER_UNSPECIFIED
	}
}

// Server is a graceful.Server-compatible wrapper around grpc.Server.
type Server struct {
	server *grpc.Server
	addr   string
}

// NewServer registers Service on a new grpc.Server with the internal-JWT
// auth interceptor. Cashier's two RPCs (SuspendWorkspace,
// ListResourceAllocations) carry workspace explicitly, so no tenant
// metadata interceptor is needed.
func NewServer(addr string, svc *Service, verifier *auth.InternalVerifier) *Server {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(auth.UnaryServerInterceptor(verifier)),
		grpc.ChainStreamInterceptor(auth.StreamServerInterceptor(verifier)),
	)
	conductor.RegisterConductorServiceServer(s, svc)
	return &Server{server: s, addr: addr}
}

func (s *Server) Label() string { return "conductor-grpc" }

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)

	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}

	return s.server.Serve(lis)
}

func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return errors.New("conductor-grpc: forced stop after timeout")
	}
}
