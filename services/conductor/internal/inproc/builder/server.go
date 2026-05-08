package builder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/builder/build"
	"github.com/zeitlos/lucity/services/conductor/internal/builder/engine"
	"github.com/zeitlos/lucity/services/conductor/internal/data"
)

// ErrBuildNotFound is returned when a build ID can't be located.
var ErrBuildNotFound = errors.New("build not found")

// Server implements the builder logic. It used to be a gRPC server
// reachable via a bufconn pipe; today the handler holds it directly
// and calls Go methods.
type Server struct {
	engine           engine.Engine
	tracker          build.Tracker
	registryURL      string
	registryUsername string
	registryPassword string
	registryInsecure bool
	workDir          string
}

// NewServer creates a new builder gRPC server.
func NewServer(eng engine.Engine, tracker build.Tracker, registryURL, registryUsername, registryPassword string, registryInsecure bool, workDir string) *Server {
	return &Server{
		engine:           eng,
		tracker:          tracker,
		registryURL:      registryURL,
		registryUsername: registryUsername,
		registryPassword: registryPassword,
		registryInsecure: registryInsecure,
		workDir:          workDir,
	}
}

// DetectServices clones the source repo and runs Railpack to detect
// services and frameworks.
func (s *Server) DetectServices(ctx context.Context, sourceURL, gitRef string) ([]data.DetectedService, error) {
	slog.Info("DetectServices called", "source_url", sourceURL)

	ghToken := auth.GitHubTokenFrom(ctx)
	repoPath, err := s.cloneRepo(ctx, sourceURL, gitRef, ghToken)
	if err != nil {
		slog.Error("clone failed", "source_url", sourceURL, "error", err)
		return nil, fmt.Errorf("failed to clone repo: %w", err)
	}
	defer os.RemoveAll(repoPath)

	results, err := s.engine.Detect(ctx, repoPath)
	if err != nil {
		slog.Error("detection failed", "source_url", sourceURL, "error", err)
		return nil, fmt.Errorf("detection failed: %w", err)
	}
	slog.Info("detection complete", "source_url", sourceURL, "services_found", len(results))

	out := make([]data.DetectedService, 0, len(results))
	for _, r := range results {
		out = append(out, data.DetectedService{
			Name:          r.Name,
			Provider:      r.Provider,
			Framework:     r.Framework,
			StartCommand:  r.StartCommand,
			SuggestedPort: r.SuggestedPort,
		})
	}
	return out, nil
}

// StartBuild kicks off an async build job and returns its ID.
func (s *Server) StartBuild(ctx context.Context, sourceURL, gitRef, service, registry, contextPath string) (string, error) {
	slog.Info("StartBuild called", "source_url", sourceURL, "service", service, "registry", registry)

	buildID := uuid.NewString()
	s.tracker.Create(buildID)

	ghToken := auth.GitHubTokenFrom(ctx)
	go s.runBuild(buildID, ghToken, sourceURL, gitRef, registry, contextPath)
	return buildID, nil
}

// BuildStatus returns the current state of an async build.
func (s *Server) BuildStatus(ctx context.Context, buildID string) (data.BuildStatus, error) {
	state := s.tracker.Get(buildID)
	if state == nil {
		return data.BuildStatus{}, ErrBuildNotFound
	}
	return data.BuildStatus{
		Phase:    state.Phase,
		ImageRef: state.ImageRef,
		Digest:   state.Digest,
		Error:    state.Error,
	}, nil
}

// BuildLogs returns a channel of build log lines, starting from the
// given offset. The channel is closed when the build reaches a
// terminal phase and all lines have been emitted, or when ctx is
// cancelled.
func (s *Server) BuildLogs(ctx context.Context, buildID string, offset int) (<-chan string, error) {
	if s.tracker.Get(buildID) == nil {
		return nil, ErrBuildNotFound
	}

	out := make(chan string, 32)
	go func() {
		defer close(out)
		for {
			lines := s.tracker.LogLines(buildID, offset)
			for _, line := range lines {
				select {
				case out <- line:
				case <-ctx.Done():
					return
				}
			}
			offset += len(lines)

			if s.tracker.IsTerminal(buildID) {
				// Drain any final lines that appeared between the last check and now.
				for _, line := range s.tracker.LogLines(buildID, offset) {
					select {
					case out <- line:
					case <-ctx.Done():
						return
					}
				}
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(200 * time.Millisecond):
			}
		}
	}()
	return out, nil
}

// DeleteImages removes all OCI repositories belonging to a project.
func (s *Server) DeleteImages(ctx context.Context, workspace, project string) ([]string, error) {
	slog.Info("DeleteImages called", "project", project)

	repos, err := s.projectRepositories(ctx, workspace, project)
	if err != nil {
		slog.Warn("failed to discover project repositories", "project", project, "error", err)
		return nil, nil
	}
	if len(repos) == 0 {
		slog.Info("no repositories found for project", "project", project)
		return nil, nil
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
	return deleted, nil
}

// maxBuildDuration is the maximum time to wait for a build to complete.
const maxBuildDuration = 30 * time.Minute

// runBuild executes the full build pipeline in a background goroutine.
// Creates a K8s Job and polls for completion — the Job pod handles clone/build/push.
func (s *Server) runBuild(buildID, token, sourceURL, gitRef, registry, contextPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), maxBuildDuration)
	defer cancel()

	s.tracker.Update(buildID, data.BuildPhaseBuilding)
	result, err := s.engine.Build(ctx, engine.BuildOpts{
		BuildID:     buildID,
		ContextPath: contextPath,
		SourceURL:   sourceURL,
		GitRef:      gitRef,
		GitHubToken: token,
		Registry:    registry,
		Insecure:    s.registryInsecure,
	})
	if err != nil {
		slog.Error("build failed", "build_id", buildID, "error", err)
		s.tracker.Fail(buildID, fmt.Sprintf("build failed: %v", err))
		return
	}
	s.tracker.Succeed(buildID, result.ImageRef, result.Digest)
	slog.Info("build succeeded", "build_id", buildID, "image", result.ImageRef)
}

// cloneRepo clones a source repository to a temp directory.
// It wraps go-git's PlainCloneContext in a goroutine because go-git does not
// reliably cancel on context expiry during HTTP I/O.
func (s *Server) cloneRepo(ctx context.Context, sourceURL, gitRef, token string) (string, error) {
	tmpDir, err := os.MkdirTemp(s.workDir, "build-*")
	if err != nil {
		return "", fmt.Errorf("failed to create work dir: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL:          sourceURL,
		Depth:        1,
		SingleBranch: true,
	}
	if token != "" {
		cloneOpts.Auth = &githttp.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	slog.Info("cloning repo", "url", sourceURL, "ref", gitRef)

	type cloneResult struct{ err error }
	done := make(chan cloneResult, 1)
	go func() {
		_, err := git.PlainCloneContext(ctx, tmpDir, false, cloneOpts)
		done <- cloneResult{err}
	}()

	select {
	case result := <-done:
		if result.err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("git clone failed: %w", result.err)
		}
		return tmpDir, nil
	case <-ctx.Done():
		// go-git doesn't always honour context cancellation during HTTP I/O.
		// Return immediately so the gRPC handler isn't blocked.
		slog.Warn("clone context expired, returning early", "url", sourceURL, "error", ctx.Err())
		go func() {
			<-done // wait for goroutine to finish before cleaning up
			os.RemoveAll(tmpDir)
		}()
		return "", fmt.Errorf("git clone timed out: %w", ctx.Err())
	}
}

