package grpc

import (
	"context"
	"net"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/tenant"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	server *grpc.Server
	addr   string
}

func NewGRPCServer(addr string, svc *Server, verifier *auth.InternalVerifier) *GRPCServer {
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
	builder.RegisterBuilderServiceServer(s, svc)

	return &GRPCServer{
		server: s,
		addr:   addr,
	}
}

func (s *GRPCServer) Label() string {
	return "gRPC"
}

func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.server.Serve(lis)
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.server.GracefulStop()
	return nil
}
