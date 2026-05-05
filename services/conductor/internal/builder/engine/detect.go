package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	rplog "github.com/railwayapp/railpack/core/logger"
)

// Detect scans source code at the given path and returns detected services.
// Shared by all engine implementations — detection always runs in-process.
func Detect(ctx context.Context, repoPath string) ([]DetectResult, error) {
	a, err := app.NewApp(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read app: %w", err)
	}

	env := app.NewEnvironment(nil)
	result := core.GenerateBuildPlan(a, env, &core.GenerateBuildPlanOptions{})

	if !result.Success || len(result.DetectedProviders) == 0 {
		// Log railpack errors when detection fails to help diagnose issues
		for _, l := range result.Logs {
			if l.Level == rplog.Error {
				slog.Warn("railpack detection error", "msg", l.Msg)
			}
		}
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
