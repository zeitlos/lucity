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
)

func executeBuild(cfg Config) error {
	if len(cfg.TargetRefs) == 0 || cfg.TargetRefs[0] == "" {
		return fmt.Errorf("BUILD_TARGET_REFS is empty")
	}

	workDir := "/tmp/lucity-builds"

	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("failed to create work dir: %w", err)
	}

	slog.Info("waiting for buildkit")

	if err := waitForBuildKit(cfg.BuildkitAddr); err != nil {
		return fmt.Errorf("buildkit not ready: %w", err)
	}

	slog.Info("buildkit ready")

	// 2. Clone the repository
	slog.Info("cloning repository", "url", cfg.SourceURL, "ref", cfg.GitRef)

	repoPath, err := cloneForBuild(workDir, cfg.SourceURL, cfg.GitHubToken)

	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	defer os.RemoveAll(repoPath)

	// 3. Normalize file timestamps to the commit time so BuildKit cache keys
	// are deterministic for the same commit (git clone sets mtimes to "now").
	if err := normalizeTimestamps(repoPath); err != nil {
		slog.Warn("failed to normalize timestamps", "error", err)
	}

	// 4. Remove .git directory — no longer needed and its contents differ
	// between clones of the same commit (pack files, index), which would
	// cause BuildKit COPY cache misses.
	os.RemoveAll(filepath.Join(repoPath, ".git"))

	// 5. Generate railpack plan
	buildDir := repoPath

	if cfg.ContextPath != "" {
		buildDir = filepath.Join(repoPath, cfg.ContextPath)
	}

	slog.Info("generating railpack plan", "dir", buildDir)

	buildPlan, err := generatePlan(buildDir)

	if err != nil {
		return err
	}

	// 6. Build with BuildKit Go client (bypasses gateway frontend so cache import works).
	// Image refs are pre-computed by the conductor (including the tag) and passed via
	// BUILD_TARGET_REFS. The runner doesn't derive tags from the repo.
	imageName := cfg.TargetRefs[0]
	cacheRef := stripTag(imageName) + ":buildcache"

	slog.Info("building image", "image", imageName, "cache", cacheRef)

	digest, err := buildWithBuildKit(context.Background(), cfg, buildDir, imageName, cacheRef, buildPlan)

	if err != nil {
		return err
	}

	slog.Info("build completed", "image", imageName, "digest", digest)

	return nil
}

// stripTag returns the image ref without its trailing tag.
// Port-safe: ignores colons that come before the last slash.
func stripTag(ref string) string {
	slashIdx := strings.LastIndex(ref, "/")
	colonIdx := strings.LastIndex(ref, ":")

	if colonIdx > slashIdx {
		return ref[:colonIdx]
	}

	return ref
}

func waitForBuildKit(addr string) error {
	network := "unix"
	dialAddr := strings.TrimPrefix(addr, "unix://")

	if strings.HasPrefix(addr, "tcp://") {
		network = "tcp"
		dialAddr = strings.TrimPrefix(addr, "tcp://")
	}

	for range 60 {
		conn, err := net.DialTimeout(network, dialAddr, time.Second)

		if err == nil {
			conn.Close()
			return nil
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("buildkit not available at %s after 60s", addr)
}

func cloneForBuild(workDir, sourceURL, token string) (string, error) {
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

// normalizeTimestamps sets all file modification times in the repo to the HEAD
// commit time. This ensures BuildKit COPY cache keys are deterministic for the
// same commit regardless of when the clone happened.
func normalizeTimestamps(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}

	commitTime := commit.Author.When

	return filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		return os.Chtimes(path, commitTime, commitTime)
	})
}

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
//
// The platform's OCI registry (Zot) is served over plain HTTP on the cluster-internal
// address, so all registry interactions are flagged insecure.
func buildWithBuildKit(ctx context.Context, cfg Config, buildDir, imageName, cacheRef string, buildPlan *plan.BuildPlan) (string, error) {
	var clientOpts []client.ClientOpt

	if cfg.BuildkitTLSCACert != "" {
		clientOpts = append(clientOpts,
			client.WithServerConfig(cfg.BuildkitServer, cfg.BuildkitTLSCACert),
			client.WithCredentials(cfg.BuildkitTLSCert, cfg.BuildkitTLSKey),
		)
	}

	c, err := client.New(ctx, cfg.BuildkitAddr, clientOpts...)
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
		"registry.insecure":     "true",
	}

	// Cache import from registry (cache miss on first build is handled gracefully)
	importCacheAttrs := map[string]string{
		"ref":               cacheRef,
		"registry.insecure": "true",
	}

	// Cache export to registry (mode=max includes all intermediate layers).
	// image-manifest=true forces a standard OCI image manifest instead of an image index —
	// required for Zot compatibility (https://github.com/project-zot/zot/issues/2728).
	exportCacheAttrs := map[string]string{
		"ref":               cacheRef,
		"mode":              "max",
		"image-manifest":    "true",
		"registry.insecure": "true",
	}

	// Registry auth: load Docker config from DOCKER_CONFIG env var (set to
	// /etc/registry-auth in the K8s Job pod, backed by the registry-auth Secret).
	// The Job pod runs trusted platform code (not user code), so mounting
	// credentials here is safe. User code only executes inside BuildKit RUN steps
	// on the buildkitd pod, which is a separate container.
	dockerCfg, err := dockerconfig.Load("")
	if err != nil {
		slog.Warn("failed to load docker config for registry auth", "error", err)
	}

	var sessionAttachables []session.Attachable
	if dockerCfg != nil {
		sessionAttachables = append(sessionAttachables, authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{
			AuthConfigProvider: authprovider.LoadAuthConfig(dockerCfg),
		}))
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
