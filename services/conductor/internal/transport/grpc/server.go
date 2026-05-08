package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/conductor"
	"github.com/zeitlos/lucity/services/conductor/internal/api/handler"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
)

// Service implements the ConductorService gRPC server. It is the
// inbound surface the conductor exposes to other platform services
// (today: cashier).
//
// The implementation is intentionally thin during the migration: the
// heavy lifting still happens in the (separately-deployed) deployer
// service. Conductor proxies the SuspendWorkspace call to the
// existing deployer gRPC client that the handler already holds.
// Phase 4+ will replace the proxy with a direct in-process
// implementation once deployer's logic is fully absorbed.
type Service struct {
	conductor.UnimplementedConductorServiceServer
	api *handler.Client
}

// NewService builds the gRPC service. It reuses the handler.Client's
// existing deployer connection so SuspendWorkspace's behavior is
// byte-identical to today's deployer.SuspendWorkspace path.
func NewService(api *handler.Client) *Service {
	return &Service{api: api}
}

// SuspendWorkspace forwards into the inproc deployer.
func (s *Service) SuspendWorkspace(ctx context.Context, req *conductor.SuspendWorkspaceRequest) (*conductor.SuspendWorkspaceResponse, error) {
	if s.api == nil || s.api.Deployer == nil {
		return nil, status.Error(codes.FailedPrecondition, "deployer not wired")
	}
	if req.GetWorkspace() == "" {
		return nil, status.Error(codes.InvalidArgument, "workspace required")
	}

	if err := s.api.Deployer.SuspendWorkspace(ctx, req.GetWorkspace(), req.GetSuspended()); err != nil {
		return nil, fmt.Errorf("deployer suspend: %w", err)
	}

	return &conductor.SuspendWorkspaceResponse{}, nil
}

// ListResourceAllocations forwards into the inproc deployer and
// translates the response into the conductor.proto shape.
func (s *Service) ListResourceAllocations(ctx context.Context, req *conductor.ListResourceAllocationsRequest) (*conductor.ListResourceAllocationsResponse, error) {
	if s.api == nil || s.api.Deployer == nil {
		return nil, status.Error(codes.FailedPrecondition, "deployer not wired")
	}

	allocations, err := s.api.Deployer.ListResourceAllocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("deployer list resource allocations: %w", err)
	}

	out := &conductor.ListResourceAllocationsResponse{
		Allocations: make([]*conductor.ResourceAllocation, 0, len(allocations)),
	}
	for _, a := range allocations {
		out.Allocations = append(out.Allocations, &conductor.ResourceAllocation{
			Workspace:     a.Workspace,
			Project:       a.Project,
			Environment:   a.Environment,
			Tier:          tierToConductorProto(a.Tier),
			CpuMillicores: int32(a.CPUMillicores),
			MemoryMb:      int32(a.MemoryMB),
			DiskMb:        int32(a.DiskMB),
		})
	}
	return out, nil
}

// tierToConductorProto converts data.ResourceTier to the conductor proto enum.
func tierToConductorProto(t data.ResourceTier) conductor.ResourceTier {
	switch t {
	case data.ResourceTierProduction:
		return conductor.ResourceTier_RESOURCE_TIER_PRODUCTION
	case data.ResourceTierEco:
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

// NewServer registers Service on a new grpc.Server with the
// internal-JWT auth interceptor. The two RPCs cashier calls
// (SuspendWorkspace, ListResourceAllocations) carry workspace as an
// explicit field, so no tenant-metadata interceptor is needed.
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
