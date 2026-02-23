package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

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
	engine           engine.Engine
	tracker          *build.Tracker
	registryURL      string
	registryToken    string
	registryInsecure bool
	workDir          string
}

// NewServer creates a new builder gRPC server.
func NewServer(eng engine.Engine, registryURL, registryToken string, registryInsecure bool, workDir string) *Server {
	return &Server{
		engine:           eng,
		tracker:          build.NewTracker(),
		registryURL:      registryURL,
		registryToken:    registryToken,
		registryInsecure: registryInsecure,
		workDir:          workDir,
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

// BuildLogs streams build log lines in real time. It sends existing lines
// from the given offset, then continues sending new lines until the build
// reaches a terminal phase and all lines have been sent.
func (s *Server) BuildLogs(req *builder.BuildLogsRequest, stream builder.BuilderService_BuildLogsServer) error {
	state := s.tracker.Get(req.BuildId)
	if state == nil {
		return status.Error(codes.NotFound, "build not found")
	}

	offset := int(req.Offset)

	for {
		lines := s.tracker.LogLines(req.BuildId, offset)
		for _, line := range lines {
			if err := stream.Send(&builder.BuildLogEntry{Line: line}); err != nil {
				return err
			}
		}
		offset += len(lines)

		// If the build is done and we've sent all lines, we're finished.
		if s.tracker.IsTerminal(req.BuildId) {
			// Drain any final lines that appeared between the last check and now.
			final := s.tracker.LogLines(req.BuildId, offset)
			for _, line := range final {
				if err := stream.Send(&builder.BuildLogEntry{Line: line}); err != nil {
					return err
				}
			}
			return nil
		}

		// Wait before checking for new lines.
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func (s *Server) DeleteImages(ctx context.Context, req *builder.DeleteImagesRequest) (*builder.DeleteImagesResponse, error) {
	if auth.FromContext(ctx) == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	slog.Info("DeleteImages called", "project", req.Project)

	repos, err := s.projectRepositories(ctx, req.Project)
	if err != nil {
		slog.Warn("failed to discover project repositories", "project", req.Project, "error", err)
		return &builder.DeleteImagesResponse{}, nil
	}

	if len(repos) == 0 {
		slog.Info("no repositories found for project", "project", req.Project)
		return &builder.DeleteImagesResponse{}, nil
	}

	var deleted []string
	for _, repo := range repos {
		if err := s.deleteRepository(ctx, repo); err != nil {
			slog.Warn("failed to delete repository", "repo", repo, "error", err)
			continue
		}
		slog.Info("deleted repository", "repo", repo)
		deleted = append(deleted, repo)
	}

	return &builder.DeleteImagesResponse{DeletedRepositories: deleted}, nil
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

	// Build + push
	s.tracker.Update(buildID, builder.BuildPhase_BUILD_PHASE_BUILDING)
	result, err := s.engine.Build(ctx, engine.BuildOpts{
		RepoPath:    repoPath,
		ImageName:   imageName,
		ContextPath: req.ContextPath,
		Token:       s.registryToken,
		SourceURL:   req.SourceUrl,
		GitSHA:      full,
		Insecure:    s.registryInsecure,
		LogFunc:     func(line string) { s.tracker.AppendLog(buildID, line) },
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
