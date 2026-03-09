package grpc

import (
	"context"
	"net"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/tenant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server *grpc.Server
	addr   string
}

func NewGRPCServer(addr, jwtSecret string, svc *Server) *GRPCServer {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			auth.UnaryServerInterceptor(jwtSecret),
			tenant.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			auth.StreamServerInterceptor(jwtSecret),
			tenant.StreamServerInterceptor(),
		),
	)
	deployer.RegisterDeployerServiceServer(s, svc)
	reflection.Register(s)

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
