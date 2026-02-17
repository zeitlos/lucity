package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	rplog "github.com/railwayapp/railpack/core/logger"
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
	a, err := app.NewApp(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read app: %w", err)
	}

	env := app.NewEnvironment(nil)
	result := core.GenerateBuildPlan(a, env, &core.GenerateBuildPlanOptions{})

	if !result.Success || len(result.DetectedProviders) == 0 {
		return nil, nil
	}

	provider := result.DetectedProviders[0]

	startCmd := ""
	if result.Plan != nil {
		startCmd = result.Plan.Deploy.StartCmd
	}

	framework := detectFramework(provider, result.Metadata, repoPath)

	slog.Info("detected service",
		"provider", provider,
		"framework", framework,
		"startCommand", startCmd,
		"providers", result.DetectedProviders,
	)

	return []DetectResult{{
		Name:          serviceName(framework, provider),
		Provider:      provider,
		Framework:     framework,
		StartCommand:  startCmd,
		SuggestedPort: defaultPort(provider),
	}}, nil
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

	// Login to registry before building (buildx push needs credentials)
	loginCmd := exec.CommandContext(ctx, "docker", "login", registryHost(opts.ImageName),
		"-u", "x-access-token", "--password-stdin")
	loginCmd.Stdin = strings.NewReader(opts.Token)

	if output, err := loginCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("docker login failed: %w: %s", err, output)
	}

	// Build and push with docker buildx using the railpack frontend.
	// This builds the image via BuildKit and pushes directly to registry.
	slog.Info("building image with railpack frontend", "image", opts.ImageName, "dir", buildDir)

	buildCmd := exec.CommandContext(ctx, "docker", "buildx", "build",
		"--build-arg", "BUILDKIT_SYNTAX=ghcr.io/railwayapp/railpack-frontend",
		"-f", planFile,
		"--output", fmt.Sprintf("type=image,name=%s,push=true", opts.ImageName),
		"--progress", "plain",
		"--metadata-file", filepath.Join(buildDir, "buildx-metadata.json"),
		buildDir,
	)
	buildCmd.Dir = buildDir

	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		slog.Error("build failed", "error", err, "output", string(buildOutput))
		return nil, fmt.Errorf("build failed: %w: %s", err, string(buildOutput))
	}
	slog.Info("build and push completed", "image", opts.ImageName)

	// Extract digest from buildx metadata
	digest := extractBuildxDigest(filepath.Join(buildDir, "buildx-metadata.json"))
	if digest == "" {
		// Fallback: parse from build output
		digest = extractDigest(string(buildOutput))
	}

	return &BuildResult{
		ImageRef: opts.ImageName,
		Digest:   digest,
	}, nil
}

// extractBuildxDigest reads the digest from docker buildx metadata JSON.
func extractBuildxDigest(metadataFile string) string {
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return ""
	}
	defer os.Remove(metadataFile)

	var metadata struct {
		Digest string `json:"containerimage.digest"`
	}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return ""
	}
	return metadata.Digest
}

// registryHost extracts the registry host from an image reference.
// "ghcr.io/user/proj/svc:tag" → "ghcr.io"
func registryHost(imageRef string) string {
	parts := strings.SplitN(imageRef, "/", 2)
	if len(parts) > 0 && strings.Contains(parts[0], ".") {
		return parts[0]
	}
	return "docker.io"
}

// extractDigest parses the digest from docker push output.
// Looks for "digest: sha256:..." in the output.
func extractDigest(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(strings.ToLower(line), "digest:"); idx >= 0 {
			rest := strings.TrimSpace(line[idx+7:])
			// Take just the sha256:... part
			if parts := strings.Fields(rest); len(parts) > 0 {
				return parts[0]
			}
		}
	}
	return ""
}

// detectFramework determines the specific framework from the provider and metadata.
func detectFramework(provider string, metadata map[string]string, repoPath string) string {
	switch provider {
	case "node":
		// Railpack sets "nodeSPAFramework" for Vite, Astro, Angular, CRA, React Router
		if fw := metadata["nodeSPAFramework"]; fw != "" {
			return fw
		}
		// Check package.json for non-SPA frameworks
		if hasPackageDep(repoPath, "next") {
			return "nextjs"
		}
		if hasPackageDep(repoPath, "nuxt") {
			return "nuxt"
		}
		if hasPackageDep(repoPath, "@remix-run/node") {
			return "remix"
		}
		if hasPackageDep(repoPath, "svelte") {
			return "svelte"
		}
		return "node"
	case "python":
		if fileExists(repoPath, "manage.py") {
			return "django"
		}
		if fileContains(repoPath, "requirements.txt", "fastapi") ||
			fileContains(repoPath, "pyproject.toml", "fastapi") {
			return "fastapi"
		}
		if fileContains(repoPath, "requirements.txt", "flask") ||
			fileContains(repoPath, "pyproject.toml", "flask") {
			return "flask"
		}
		return "python"
	case "golang":
		return "go"
	case "rust":
		return "rust"
	case "ruby":
		if fileExists(repoPath, "config/routes.rb") {
			return "rails"
		}
		return "ruby"
	case "php":
		if fileExists(repoPath, "artisan") {
			return "laravel"
		}
		return "php"
	case "java":
		return "java"
	case "elixir":
		if fileExists(repoPath, "mix.exs") && fileContains(repoPath, "mix.exs", "phoenix") {
			return "phoenix"
		}
		return "elixir"
	case "dotnet":
		return "dotnet"
	case "deno":
		return "deno"
	default:
		return provider
	}
}

// defaultPort returns a framework-appropriate default port.
func defaultPort(provider string) int {
	switch provider {
	case "node":
		return 3000
	case "python":
		return 8000
	case "golang":
		return 8080
	case "ruby":
		return 3000
	case "php":
		return 8080
	case "elixir":
		return 4000
	case "java":
		return 8080
	case "rust":
		return 8080
	case "deno":
		return 8000
	case "dotnet":
		return 5000
	default:
		return 8080
	}
}

// serviceName generates a suggested service name from the framework/provider.
func serviceName(framework, provider string) string {
	if framework != "" && framework != provider {
		return "web"
	}
	return "web"
}

// hasPackageDep checks if a package.json contains a dependency.
func hasPackageDep(repoPath, dep string) bool {
	data, err := os.ReadFile(filepath.Join(repoPath, "package.json"))
	if err != nil {
		return false
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	if _, ok := pkg.Dependencies[dep]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies[dep]; ok {
		return true
	}
	return false
}

// fileExists checks if a file exists relative to the repo path.
func fileExists(repoPath, relPath string) bool {
	_, err := os.Stat(filepath.Join(repoPath, relPath))
	return err == nil
}

// fileContains checks if a file contains a substring.
func fileContains(repoPath, relPath, substr string) bool {
	data, err := os.ReadFile(filepath.Join(repoPath, relPath))
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), strings.ToLower(substr))
}

// errorLogs extracts error-level messages from railpack logs.
func errorLogs(logs []rplog.Msg) []string {
	var errs []string
	for _, l := range logs {
		if l.Level == rplog.Error {
			errs = append(errs, l.Msg)
		}
	}
	return errs
}
