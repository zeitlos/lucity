package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LocalEngine builds images locally using railpack CLI and Docker.
// Requires railpack and Docker to be available on the host.
type LocalEngine struct{}

// NewLocalEngine creates a LocalEngine.
func NewLocalEngine() *LocalEngine {
	return &LocalEngine{}
}

// railpackPlan is the JSON output structure from `railpack plan`.
type railpackPlan struct {
	Providers []string          `json:"providers"`
	Metadata  map[string]string `json:"metadata"`
	Plan      struct {
		Deploy struct {
			StartCmd string `json:"startCmd"`
		} `json:"deploy"`
	} `json:"plan"`
}

func (e *LocalEngine) Detect(ctx context.Context, repoPath string) ([]DetectResult, error) {
	// Run railpack plan to get detection info
	cmd := exec.CommandContext(ctx, "railpack", "plan", ".", "--format", "json")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("railpack plan failed: %w: %s", err, stderr.String())
	}

	var plan railpackPlan
	if err := json.Unmarshal(stdout.Bytes(), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse railpack plan output: %w", err)
	}

	provider := ""
	if len(plan.Providers) > 0 {
		provider = plan.Providers[0]
	}

	if provider == "" {
		return nil, nil // no provider detected
	}

	framework := detectFramework(provider, plan.Metadata, repoPath)

	return []DetectResult{{
		Name:          serviceName(framework, provider),
		Provider:      provider,
		Framework:     framework,
		StartCommand:  plan.Plan.Deploy.StartCmd,
		SuggestedPort: defaultPort(provider),
	}}, nil
}

func (e *LocalEngine) Build(ctx context.Context, opts BuildOpts) (*BuildResult, error) {
	buildDir := opts.RepoPath
	if opts.ContextPath != "" {
		buildDir = filepath.Join(opts.RepoPath, opts.ContextPath)
	}

	// Build with railpack
	slog.Info("building image with railpack", "image", opts.ImageName, "dir", buildDir)
	buildCmd := exec.CommandContext(ctx, "railpack", "build", ".", "--name", opts.ImageName)
	buildCmd.Dir = buildDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return nil, fmt.Errorf("railpack build failed: %w", err)
	}

	// Login to registry
	loginCmd := exec.CommandContext(ctx, "docker", "login", registryHost(opts.ImageName),
		"-u", "x-access-token", "--password-stdin")
	loginCmd.Stdin = strings.NewReader(opts.Token)

	if output, err := loginCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("docker login failed: %w: %s", err, output)
	}

	// Push image
	slog.Info("pushing image", "image", opts.ImageName)
	pushCmd := exec.CommandContext(ctx, "docker", "push", opts.ImageName)
	pushOutput, err := pushCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker push failed: %w: %s", err, pushOutput)
	}

	// Extract digest from push output
	digest := extractDigest(string(pushOutput))

	return &BuildResult{
		ImageRef: opts.ImageName,
		Digest:   digest,
	}, nil
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
