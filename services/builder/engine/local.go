package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
)

// LocalEngine builds images locally using railpack and docker buildx.
// Uses the railpack Go library for detection/plan generation and the
// railpack BuildKit frontend (ghcr.io/railwayapp/railpack-frontend)
// for building via docker buildx.
type LocalEngine struct{}

// NewLocalEngine creates a LocalEngine.
func NewLocalEngine() *LocalEngine {
	return &LocalEngine{}
}

func (e *LocalEngine) Detect(ctx context.Context, repoPath string) ([]DetectResult, error) {
	return Detect(ctx, repoPath)
}

func (e *LocalEngine) Build(ctx context.Context, opts BuildOpts) (*BuildResult, error) {
	buildDir := opts.RepoPath
	if opts.ContextPath != "" {
		buildDir = filepath.Join(opts.RepoPath, opts.ContextPath)
	}

	// Generate build plan using railpack library
	a, err := app.NewApp(buildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read app: %w", err)
	}

	env := app.NewEnvironment(nil)
	result := core.GenerateBuildPlan(a, env, &core.GenerateBuildPlanOptions{})
	if !result.Success || result.Plan == nil {
		errMsg := "unknown error"
		if errs := errorLogs(result.Logs); len(errs) > 0 {
			errMsg = strings.Join(errs, "; ")
		}
		return nil, fmt.Errorf("railpack plan generation failed: %s", errMsg)
	}

	// Write build plan to a temp file for the BuildKit frontend
	planJSON, err := json.Marshal(result.Plan)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal build plan: %w", err)
	}

	planFile := filepath.Join(buildDir, "railpack-plan.json")
	if err := os.WriteFile(planFile, planJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to write build plan: %w", err)
	}
	defer os.Remove(planFile)

	// Build the image and load it into the local Docker image store.
	// We split build and push into two steps because BuildKit's push (via the
	// docker driver) uses the Docker daemon's credential store inside the VM,
	// which doesn't reliably pick up credentials set via DOCKER_CONFIG or
	// docker login on the host. By loading first, then pushing with `docker push`
	// (which runs on the host and respects DOCKER_CONFIG), we get reliable auth.
	slog.Info("building image with railpack frontend", "image", opts.ImageName, "dir", buildDir)

	args := []string{"buildx", "build",
		"--build-arg", "BUILDKIT_SYNTAX=ghcr.io/railwayapp/railpack-frontend",
		"-f", planFile,
		"--tag", opts.ImageName,
		"--load",
		"--progress", "plain",
	}

	args = append(args, buildDir)

	buildCmd := exec.CommandContext(ctx, "docker", args...)
	buildCmd.Dir = buildDir

	buildOutput, err := runAndStream(buildCmd, opts.LogFunc)
	if err != nil {
		slog.Error("build failed", "error", err, "output", string(buildOutput))
		return nil, fmt.Errorf("build failed: %w: %s", err, string(buildOutput))
	}

	// Apply OCI labels in a post-build step. The railpack BuildKit frontend
	// ignores --label flags passed to `docker buildx build`, so we layer them
	// on top with a tiny inline Dockerfile. This only updates image metadata,
	// no new filesystem layers are created.
	if err := applyLabels(ctx, opts); err != nil {
		return nil, fmt.Errorf("failed to apply labels: %w", err)
	}

	slog.Info("build completed, pushing image", "image", opts.ImageName)

	digest, err := pushImage(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &BuildResult{
		ImageRef: opts.ImageName,
		Digest:   digest,
	}, nil
}

// pushImage saves the image from the Docker daemon to a tarball and pushes it
// to the registry using crane. This runs entirely on the host, avoiding Docker
// Desktop's daemon-level TLS enforcement which breaks pushes to HTTP registries
// like a local Zot instance.
func pushImage(ctx context.Context, opts BuildOpts) (string, error) {
	// Save the image from the Docker daemon to a tarball
	tarFile, err := os.CreateTemp("", "image-*.tar")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tarPath := tarFile.Name()
	tarFile.Close()
	defer os.Remove(tarPath)

	saveCmd := exec.CommandContext(ctx, "docker", "save", opts.ImageName, "-o", tarPath)
	if output, err := saveCmd.CombinedOutput(); err != nil {
		slog.Error("docker save failed", "error", err, "output", string(output))
		return "", fmt.Errorf("docker save failed: %w: %s", err, string(output))
	}

	// Push using crane (runs on the host, supports --insecure for HTTP registries)
	args := []string{"push", tarPath, opts.ImageName}
	if opts.Insecure {
		args = append(args, "--insecure")
	}

	pushCmd := exec.CommandContext(ctx, "crane", args...)
	pushOutput, err := runAndStream(pushCmd, opts.LogFunc)
	if err != nil {
		slog.Error("push failed", "error", err, "output", string(pushOutput))
		return "", fmt.Errorf("push failed: %w: %s", err, string(pushOutput))
	}
	slog.Info("push completed", "image", opts.ImageName)

	// crane outputs "registry/repo@sha256:..." on success — extract the digest
	digest := extractCraneDigest(string(pushOutput))
	return digest, nil
}

// runAndStream runs a command, streaming each output line to logFunc (if non-nil),
// and returns all combined output for error reporting.
func runAndStream(cmd *exec.Cmd, logFunc func(string)) ([]byte, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	var output []byte
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // handle long lines
	for scanner.Scan() {
		line := scanner.Text()
		output = append(output, line...)
		output = append(output, '\n')
		if logFunc != nil {
			logFunc(line)
		}
	}

	// Drain any remaining data if scanner hit an error
	if scanner.Err() != nil {
		remaining, _ := io.ReadAll(stdout)
		output = append(output, remaining...)
	}

	if err := cmd.Wait(); err != nil {
		return output, err
	}
	return output, nil
}

// extractCraneDigest parses the digest from crane push output.
// crane outputs lines like "registry/repo@sha256:abc123..." on success.
func extractCraneDigest(output string) string {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "@sha256:"); idx >= 0 {
			return line[idx+1:] // return "sha256:..."
		}
	}
	return ""
}

// applyLabels re-tags an already-loaded image with OCI labels using an inline
// Dockerfile. The railpack BuildKit frontend ignores --label flags, so we apply
// them in a separate build step that only updates image metadata (no new layers).
func applyLabels(ctx context.Context, opts BuildOpts) error {
	var labels []string
	if opts.SourceURL != "" {
		labels = append(labels, fmt.Sprintf("LABEL org.opencontainers.image.source=%q", opts.SourceURL))
	}
	if opts.GitSHA != "" {
		labels = append(labels, fmt.Sprintf("LABEL org.opencontainers.image.revision=%q", opts.GitSHA))
	}
	if len(labels) == 0 {
		return nil
	}

	dockerfile := fmt.Sprintf("FROM %s\n%s\n", opts.ImageName, strings.Join(labels, "\n"))

	cmd := exec.CommandContext(ctx, "docker", "build",
		"--tag", opts.ImageName,
		"--file", "-",
		".",
	)
	cmd.Stdin = strings.NewReader(dockerfile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("label application failed", "error", err, "output", string(output))
		return fmt.Errorf("docker build for labels failed: %w: %s", err, string(output))
	}
	return nil
}
