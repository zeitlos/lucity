package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	rplog "github.com/railwayapp/railpack/core/logger"
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
	planFile, err := generatePlan(buildDir)
	if err != nil {
		return err
	}
	defer os.Remove(planFile)

	// 5. Build with buildctl
	cacheRef := cfg.Registry + ":buildcache"
	slog.Info("building image", "image", imageName, "cache", cacheRef)
	if err := buildWithBuildctl(cfg.BuildkitAddr, buildDir, planFile, imageName, cacheRef, cfg.Insecure); err != nil {
		return err
	}

	// 6. Extract digest from registry (buildctl outputs it)
	digest := "" // buildctl --metadata-file approach below
	metadataFile := filepath.Join(workDir, "build-metadata.json")
	if data, err := os.ReadFile(metadataFile); err == nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal(data, &metadata); err == nil {
			if d, ok := metadata["containerimage.digest"].(string); ok {
				digest = d
			}
		}
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

// generatePlan creates a railpack build plan and writes it to disk.
func generatePlan(buildDir string) (string, error) {
	a, err := app.NewApp(buildDir)
	if err != nil {
		return "", fmt.Errorf("failed to read app: %w", err)
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
		return "", fmt.Errorf("railpack plan generation failed: %s", errMsg)
	}

	planJSON, err := json.Marshal(result.Plan)
	if err != nil {
		return "", fmt.Errorf("failed to marshal build plan: %w", err)
	}

	planFile := filepath.Join(buildDir, "railpack-plan.json")
	if err := os.WriteFile(planFile, planJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write build plan: %w", err)
	}

	return planFile, nil
}

// buildWithBuildctl invokes buildctl to build and push the image via BuildKit.
// cacheRef is the registry reference for layer caching (e.g., "registry:5000/proj/svc:buildcache").
func buildWithBuildctl(buildkitAddr, buildDir, planFile, imageName, cacheRef string, insecure bool) error {
	args := []string{
		"--addr", buildkitAddr,
		"build",
		"--progress", "plain",
		"--frontend", "gateway.v0",
		"--opt", "source=ghcr.io/railwayapp/railpack-frontend",
		"--opt", "filename=railpack-plan.json",
		"--local", "context=" + buildDir,
		"--local", "dockerfile=" + buildDir,
		"--metadata-file", "/tmp/lucity-builds/build-metadata.json",
	}

	// Import layer cache from registry (cache miss on first build is handled gracefully)
	importCache := "type=registry,ref=" + cacheRef
	if insecure {
		importCache += ",registry.insecure=true"
	}
	args = append(args, "--import-cache", importCache)

	// Export layer cache to registry for future builds (mode=max includes all intermediate layers).
	// image-manifest=true forces a standard OCI image manifest instead of an image index —
	// required for Zot compatibility (https://github.com/project-zot/zot/issues/2728).
	exportCache := "type=registry,ref=" + cacheRef + ",mode=max,image-manifest=true"
	if insecure {
		exportCache += ",registry.insecure=true"
	}
	args = append(args, "--export-cache", exportCache)

	// Output configuration: build and push to registry
	output := fmt.Sprintf("type=image,name=%s,push=true", imageName)
	if insecure {
		output += ",registry.insecure=true"
	}
	args = append(args, "--output", output)

	cmd := exec.Command("buildctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buildctl build failed: %w", err)
	}

	return nil
}

func inClusterClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	return kubernetes.NewForConfig(config)
}
