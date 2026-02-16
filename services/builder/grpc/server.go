package grpc

import (
	"context"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/builder"
)

type Server struct {
	builder.UnimplementedBuilderServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) BuildImage(ctx context.Context, req *builder.BuildImageRequest) (*builder.BuildImageResponse, error) {
	slog.Info("BuildImage called",
		"source_url", req.SourceUrl,
		"git_ref", req.GitRef,
		"service", req.Service,
		"registry", req.Registry,
	)

	return &builder.BuildImageResponse{
		ImageRef: req.Registry + "/" + req.Service + ":" + req.GitRef,
		Digest:   "sha256:mock1234567890abcdef",
	}, nil
}
