package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	_ "github.com/moby/buildkit/util/grpcutil/encoding/proto"
	"github.com/moby/buildkit/util/progress/progressui"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	rpbuildkit "github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	rplog "github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/tonistiigi/fsutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/zeitlos/lucity/services/builder/build"
)

// runBuildConfig holds env-based configuration for the build runner.
type runBuildConfig struct {
	BuildID     string
	SourceURL   string
	GitRef      string
	Registry    string
	ContextPath string
	Insecure    bool
	BuildkitAddr string
	GitHubToken string
	Namespace   string
}

func loadRunBuildConfig() runBuildConfig {
	return runBuildConfig{
		BuildID:      os.Getenv("BUILD_ID"),
		SourceURL:    os.Getenv("BUILD_SOURCE_URL"),
		GitRef:       os.Getenv("BUILD_GIT_REF"),
		Registry:     os.Getenv("BUILD_REGISTRY"),
		ContextPath:  os.Getenv("BUILD_CONTEXT_PATH"),
		Insecure:     os.Getenv("BUILD_INSECURE") == "true",
		BuildkitAddr: os.Getenv("BUILDKIT_ADDR"),
		GitHubToken:  os.Getenv("GITHUB_TOKEN"),
		Namespace:    os.Getenv("BUILD_NAMESPACE"),
	}
}

// runBuild is the entry point for the build runner that runs inside K8s Job pods.
// It clones the repo, generates a railpack plan, builds via BuildKit, and pushes
// the image. Results are annotated on the Job for the builder service to read.
func runBuild() {
	cfg := loadRunBuildConfig()

	slog.Info("build runner starting",
		"build_id", cfg.BuildID,
		"source_url", cfg.SourceURL,
		"registry", cfg.Registry,
	)

	// Create in-cluster K8s client for annotating the Job
	k8sClient, err := inClusterClient()
	if err != nil {
		slog.Error("failed to create k8s client", "error", err)
		os.Exit(1)
	}

	if err := executeBuild(cfg, k8sClient); err != nil {
		slog.Error("build failed", "error", err)

		// Annotate Job with error
		if annotateErr := build.AnnotateJobError(k8sClient, cfg.Namespace, cfg.BuildID, err.Error()); annotateErr != nil {
			slog.Error("failed to annotate job with error", "error", annotateErr)
		}

		os.Exit(1)
	}
}

func executeBuild(cfg runBuildConfig, k8sClient kubernetes.Interface) error {
	workDir := "/tmp/lucity-builds"
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("failed to create work dir: %w", err)
	}

	// 1. Wait for BuildKit to be ready
	slog.Info("waiting for buildkit sidecar")
	if err := waitForBuildKit(cfg.BuildkitAddr); err != nil {
		return fmt.Errorf("buildkit not ready: %w", err)
	}
	slog.Info("buildkit ready")

	// 2. Clone the repository
	slog.Info("cloning repository", "url", cfg.SourceURL, "ref", cfg.GitRef)
	repoPath, err := cloneForBuild(workDir, cfg.SourceURL, cfg.GitRef, cfg.GitHubToken)
	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}
	defer os.RemoveAll(repoPath)

	// 3. Determine git SHA for image tag
	sha := buildFullSHA(repoPath)
	tag := sha
	if len(tag) >= 7 {
		tag = tag[:7]
	}
	imageName := cfg.Registry + ":" + tag
	slog.Info("image name determined", "image", imageName, "sha", sha)

	// 4. Generate railpack plan
	buildDir := repoPath
	if cfg.ContextPath != "" {
		buildDir = filepath.Join(repoPath, cfg.ContextPath)
	}

	slog.Info("generating railpack plan", "dir", buildDir)
	buildPlan, err := generatePlan(buildDir)
	if err != nil {
		return err
	}

	// 5. Build with BuildKit Go client (bypasses gateway frontend so cache import works)
	cacheRef := cfg.Registry + ":buildcache"
	slog.Info("building image", "image", imageName, "cache", cacheRef)
	digest, err := buildWithBuildKit(context.Background(), cfg.BuildkitAddr, buildDir, imageName, cacheRef, buildPlan, cfg.Insecure)
	if err != nil {
		return err
	}

	slog.Info("build completed", "image", imageName, "digest", digest)

	// 7. Annotate Job with result
	if err := build.AnnotateJobResult(k8sClient, cfg.Namespace, cfg.BuildID, imageName, digest); err != nil {
		return fmt.Errorf("failed to annotate job: %w", err)
	}

	slog.Info("build result annotated on job", "build_id", cfg.BuildID)
	return nil
}

// waitForBuildKit waits for the BuildKit Unix socket to become available.
func waitForBuildKit(addr string) error {
	socketPath := strings.TrimPrefix(addr, "unix://")

	for i := 0; i < 60; i++ {
		conn, err := net.DialTimeout("unix", socketPath, time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("buildkit socket not available at %s after 60s", socketPath)
}

// cloneForBuild clones a repo for the build runner.
func cloneForBuild(workDir, sourceURL, gitRef, token string) (string, error) {
	tmpDir, err := os.MkdirTemp(workDir, "build-*")
	if err != nil {
		return "", fmt.Errorf("failed to create work dir: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cloneOpts := &git.CloneOptions{
		URL: sourceURL,
		Auth: &githttp.BasicAuth{
			Username: "x-access-token",
			Password: token,
		},
		Depth:        1,
		SingleBranch: true,
	}

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
		go func() {
			<-done
			os.RemoveAll(tmpDir)
		}()
		return "", fmt.Errorf("git clone timed out: %w", ctx.Err())
	}
}

// buildFullSHA returns the full git SHA of HEAD.
func buildFullSHA(repoPath string) string {
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

// generatePlan creates a railpack build plan from the source directory.
func generatePlan(buildDir string) (*plan.BuildPlan, error) {
	a, err := app.NewApp(buildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read app: %w", err)
	}

	env := app.NewEnvironment(nil)
	result := core.GenerateBuildPlan(a, env, &core.GenerateBuildPlanOptions{})
	if !result.Success || result.Plan == nil {
		errMsg := "unknown error"
		var errs []string
		for _, l := range result.Logs {
			if l.Level == rplog.Error {
				errs = append(errs, l.Msg)
			}
		}
		if len(errs) > 0 {
			errMsg = strings.Join(errs, "; ")
		}
		return nil, fmt.Errorf("railpack plan generation failed: %s", errMsg)
	}

	return result.Plan, nil
}

// buildWithBuildKit converts the railpack plan to LLB and solves directly with the
// BuildKit Go client. This bypasses the gateway frontend, which fixes cache import —
// the railpack frontend never forwarded cache-imports to its inner solve call.
func buildWithBuildKit(ctx context.Context, buildkitAddr, buildDir, imageName, cacheRef string, buildPlan *plan.BuildPlan, insecure bool) (string, error) {
	c, err := client.New(ctx, buildkitAddr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to buildkit: %w", err)
	}
	defer c.Close()

	// Convert railpack plan to LLB
	buildPlatform := specs.Platform{OS: "linux", Architecture: "amd64"}
	llbState, image, err := rpbuildkit.ConvertPlanToLLB(buildPlan, rpbuildkit.ConvertPlanOptions{
		BuildPlatform: buildPlatform,
	})
	if err != nil {
		return "", fmt.Errorf("failed to convert plan to LLB: %w", err)
	}

	imageBytes, err := json.Marshal(image)
	if err != nil {
		return "", fmt.Errorf("failed to marshal image config: %w", err)
	}

	def, err := llbState.Marshal(ctx, llb.LinuxAmd64)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LLB: %w", err)
	}

	// Build context
	appFS, err := fsutil.NewFS(buildDir)
	if err != nil {
		return "", fmt.Errorf("failed to create build context: %w", err)
	}

	// Output: build and push to registry
	exportAttrs := map[string]string{
		"name":                  imageName,
		"push":                  "true",
		"containerimage.config": string(imageBytes),
	}
	if insecure {
		exportAttrs["registry.insecure"] = "true"
	}

	// Cache import from registry (cache miss on first build is handled gracefully)
	importCacheAttrs := map[string]string{"ref": cacheRef}
	if insecure {
		importCacheAttrs["registry.insecure"] = "true"
	}

	// Cache export to registry (mode=max includes all intermediate layers).
	// image-manifest=true forces a standard OCI image manifest instead of an image index —
	// required for Zot compatibility (https://github.com/project-zot/zot/issues/2728).
	exportCacheAttrs := map[string]string{
		"ref":            cacheRef,
		"mode":           "max",
		"image-manifest": "true",
	}
	if insecure {
		exportCacheAttrs["registry.insecure"] = "true"
	}

	// Registry auth: load Docker config from DOCKER_CONFIG env (set to /etc/registry-auth
	// in the K8s Job pod, backed by the registry-auth Secret).
	dockerCfg, err := dockerconfig.Load("")
	if err != nil {
		slog.Warn("failed to load docker config for registry auth", "error", err)
	}

	var sessionAttachables []session.Attachable
	if dockerCfg != nil {
		sessionAttachables = append(sessionAttachables, authprovider.NewDockerAuthProvider(dockerCfg, nil))
	}

	solveOpts := client.SolveOpt{
		LocalMounts: map[string]fsutil.FS{
			"context": appFS,
		},
		Session: sessionAttachables,
		Exports: []client.ExportEntry{
			{
				Type:  client.ExporterImage,
				Attrs: exportAttrs,
			},
		},
		CacheImports: []client.CacheOptionsEntry{
			{Type: "registry", Attrs: importCacheAttrs},
		},
		CacheExports: []client.CacheOptionsEntry{
			{Type: "registry", Attrs: exportCacheAttrs},
		},
	}

	// Stream build progress
	ch := make(chan *client.SolveStatus)
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		display, err := progressui.NewDisplay(os.Stdout, progressui.PlainMode)
		if err != nil {
			for range ch {
			}
			return
		}
		display.UpdateFrom(ctx, ch)
	}()

	startTime := time.Now()
	resp, err := c.Solve(ctx, def, solveOpts, ch)
	<-progressDone

	if err != nil {
		return "", fmt.Errorf("buildkit solve failed: %w", err)
	}

	slog.Info("buildkit solve completed", "duration", time.Since(startTime).Round(time.Millisecond))
	return resp.ExporterResponse["containerimage.digest"], nil
}

func inClusterClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	return kubernetes.NewForConfig(config)
}
