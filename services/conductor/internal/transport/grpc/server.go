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
	deployerproto "github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/tenant"
	"github.com/zeitlos/lucity/services/conductor/internal/api/handler"
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

// SuspendWorkspace forwards to the deployer service. The request and
// response shapes mirror the deployer proto exactly so cashier's
// existing call sites only need to swap the import path.
func (s *Service) SuspendWorkspace(ctx context.Context, req *conductor.SuspendWorkspaceRequest) (*conductor.SuspendWorkspaceResponse, error) {
	if s.api == nil || s.api.Deployer == nil {
		return nil, status.Error(codes.FailedPrecondition, "deployer not wired")
	}
	if req.GetWorkspace() == "" {
		return nil, status.Error(codes.InvalidArgument, "workspace required")
	}

	// Re-issue with the same auth + tenant context that the handler
	// uses for outgoing gRPC calls. Cashier's incoming call is
	// already authenticated by the unary interceptor below; we mint
	// a fresh outgoing context here for the deployer hop.
	outCtx := auth.OutgoingContext(ctx)
	outCtx = tenant.WithWorkspace(outCtx, req.GetWorkspace())
	outCtx = tenant.OutgoingContext(outCtx)

	if _, err := s.api.Deployer.SuspendWorkspace(outCtx, &deployerproto.SuspendWorkspaceRequest{
		Workspace: req.GetWorkspace(),
		Suspended: req.GetSuspended(),
	}); err != nil {
		return nil, fmt.Errorf("deployer suspend: %w", err)
	}

	return &conductor.SuspendWorkspaceResponse{}, nil
}

// ListResourceAllocations forwards to the deployer service and
// translates the response into the conductor.proto shape. ResourceTier
// values are stable across the two protos so the conversion is a
// straight cast.
func (s *Service) ListResourceAllocations(ctx context.Context, req *conductor.ListResourceAllocationsRequest) (*conductor.ListResourceAllocationsResponse, error) {
	if s.api == nil || s.api.Deployer == nil {
		return nil, status.Error(codes.FailedPrecondition, "deployer not wired")
	}

	outCtx := auth.OutgoingContext(ctx)
	resp, err := s.api.Deployer.ListResourceAllocations(outCtx, &deployerproto.ListResourceAllocationsRequest{})
	if err != nil {
		return nil, fmt.Errorf("deployer list resource allocations: %w", err)
	}

	out := &conductor.ListResourceAllocationsResponse{
		Allocations: make([]*conductor.ResourceAllocation, 0, len(resp.Allocations)),
	}
	for _, a := range resp.Allocations {
		out.Allocations = append(out.Allocations, &conductor.ResourceAllocation{
			Workspace:      a.GetWorkspace(),
			Project:        a.GetProject(),
			Environment:    a.GetEnvironment(),
			Tier:           conductor.ResourceTier(a.GetTier()),
			CpuMillicores:  a.GetCpuMillicores(),
			MemoryMb:       a.GetMemoryMb(),
			DiskMb:         a.GetDiskMb(),
		})
	}
	return out, nil
}

// Server is a graceful.Server-compatible wrapper around grpc.Server.
type Server struct {
	server *grpc.Server
	addr   string
}

// NewServer registers Service on a new grpc.Server with the standard
// internal-JWT auth + tenant interceptors.
func NewServer(addr string, svc *Service, verifier *auth.InternalVerifier) *Server {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			auth.UnaryServerInterceptor(verifier),
			tenant.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			auth.StreamServerInterceptor(verifier),
			tenant.StreamServerInterceptor(),
		),
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
