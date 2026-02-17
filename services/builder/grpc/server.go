package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/services/builder/build"
	"github.com/zeitlos/lucity/services/builder/engine"
)

// Server implements the BuilderService gRPC API.
type Server struct {
	builder.UnimplementedBuilderServiceServer
	engine        engine.Engine
	tracker       *build.Tracker
	registryURL   string
	registryToken string
	workDir       string
}

// NewServer creates a new builder gRPC server.
func NewServer(eng engine.Engine, registryURL, registryToken, workDir string) *Server {
	return &Server{
		engine:        eng,
		tracker:       build.NewTracker(),
		registryURL:   registryURL,
		registryToken: registryToken,
		workDir:       workDir,
	}
}

func (s *Server) DetectServices(ctx context.Context, req *builder.DetectServicesRequest) (*builder.DetectServicesResponse, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	slog.Info("DetectServices called", "source_url", req.SourceUrl)

	// Clone the repo
	repoPath, err := s.cloneRepo(ctx, req.SourceUrl, req.GitRef, claims.GitHubToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to clone repo: %v", err)
	}
	defer os.RemoveAll(repoPath)

	// Run detection
	results, err := s.engine.Detect(ctx, repoPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "detection failed: %v", err)
	}

	var services []*builder.DetectedService
	for _, r := range results {
		services = append(services, &builder.DetectedService{
			Name:          r.Name,
			Provider:      r.Provider,
			Framework:     r.Framework,
			StartCommand:  r.StartCommand,
			SuggestedPort: int32(r.SuggestedPort),
		})
	}

	return &builder.DetectServicesResponse{Services: services}, nil
}

func (s *Server) StartBuild(ctx context.Context, req *builder.StartBuildRequest) (*builder.StartBuildResponse, error) {
	claims := auth.FromContext(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	slog.Info("StartBuild called",
		"source_url", req.SourceUrl,
		"service", req.Service,
		"registry", req.Registry,
	)

	buildID := uuid.NewString()
	s.tracker.Create(buildID)

	// Run the build asynchronously
	go s.runBuild(buildID, claims.GitHubToken, req)

	return &builder.StartBuildResponse{BuildId: buildID}, nil
}

func (s *Server) BuildStatus(ctx context.Context, req *builder.BuildStatusRequest) (*builder.BuildStatusResponse, error) {
	state := s.tracker.Get(req.BuildId)
	if state == nil {
		return nil, status.Error(codes.NotFound, "build not found")
	}

	return &builder.BuildStatusResponse{
		Phase:    state.Phase,
		ImageRef: state.ImageRef,
		Digest:   state.Digest,
		Error:    state.Error,
	}, nil
}

// runBuild executes the full build pipeline in a background goroutine.
func (s *Server) runBuild(buildID, token string, req *builder.StartBuildRequest) {
	ctx := context.Background()

	// Clone
	s.tracker.Update(buildID, builder.BuildPhase_BUILD_PHASE_CLONING)
	repoPath, err := s.cloneRepo(ctx, req.SourceUrl, req.GitRef, token)
	if err != nil {
		s.tracker.Fail(buildID, fmt.Sprintf("clone failed: %v", err))
		return
	}
	defer os.RemoveAll(repoPath)

	// Get git SHA for the image tag and OCI labels
	full := fullSHA(repoPath)
	tag := full
	if len(tag) >= 7 {
		tag = tag[:7]
	}
	imageName := req.Registry + ":" + tag

	// Build + push (use registry token for GHCR push — GitHub App OAuth tokens can't push to GHCR)
	s.tracker.Update(buildID, builder.BuildPhase_BUILD_PHASE_BUILDING)
	result, err := s.engine.Build(ctx, engine.BuildOpts{
		RepoPath:    repoPath,
		ImageName:   imageName,
		ContextPath: req.ContextPath,
		Token:       s.registryToken,
		SourceURL:   req.SourceUrl,
		GitSHA:      full,
	})
	if err != nil {
		s.tracker.Fail(buildID, fmt.Sprintf("build failed: %v", err))
		return
	}

	s.tracker.Succeed(buildID, result.ImageRef, result.Digest)
	slog.Info("build succeeded", "build_id", buildID, "image", result.ImageRef)
}

// cloneRepo clones a source repository to a temp directory.
func (s *Server) cloneRepo(ctx context.Context, sourceURL, gitRef, token string) (string, error) {
	tmpDir, err := os.MkdirTemp(s.workDir, "build-*")
	if err != nil {
		return "", fmt.Errorf("failed to create work dir: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL: sourceURL,
		Auth: &githttp.BasicAuth{
			Username: "x-access-token",
			Password: token,
		},
		Depth:        1,
		SingleBranch: true,
	}

	slog.Info("cloning repo", "url", sourceURL, "ref", gitRef)
	_, err = git.PlainCloneContext(ctx, tmpDir, false, cloneOpts)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("git clone failed: %w", err)
	}

	return tmpDir, nil
}

// fullSHA returns the full git SHA of HEAD in the given repo path.
func fullSHA(repoPath string) string {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "latest"
	}
	head, err := repo.Head()
	if err != nil {
		return "latest"
	}
	return head.Hash().String()
}
