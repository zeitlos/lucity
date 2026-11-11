package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/zeitlos/lucity/pkg/logger"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	BuildID      string   `envconfig:"BUILD_ID" required:"true"`
	SourceURL    string   `envconfig:"BUILD_SOURCE_URL" required:"true"`
	GitRef       string   `envconfig:"BUILD_GIT_REF" required:"true"`
	ContextPath  string   `envconfig:"BUILD_CONTEXT_PATH"`
	TargetRefs   []string `envconfig:"BUILD_TARGET_REFS" required:"true"`
	CacheKey     string   `envconfig:"BUILD_CACHE_KEY" required:"true"`
	BuildkitAddr string   `envconfig:"BUILDKIT_ADDR" required:"true"`
	GitHubToken  string   `envconfig:"GITHUB_TOKEN"`

	BuildkitTLSCACert string `envconfig:"BUILDKIT_TLS_CA"`
	BuildkitTLSCert   string `envconfig:"BUILDKIT_TLS_CERT"`
	BuildkitTLSKey    string `envconfig:"BUILDKIT_TLS_KEY"`
	BuildkitServer    string `envconfig:"BUILDKIT_SERVER_NAME"`
}

const buildVarsFile = "/etc/lucity/.env"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "run-scan" {
		runScan()
		return
	}

	logger.Setup("info")

	var config Config

	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("build runner starting",
		"build_id", config.BuildID,
		"source_url", config.SourceURL,
		"target_refs", config.TargetRefs,
	)

	var buildVars map[string]string

	if _, err := os.Stat(buildVarsFile); err == nil {
		buildVars, err = godotenv.Read("/etc/lucity/.env")

		if err != nil {
			slog.Error("failed to load build vars file", "error", err, "file", buildVarsFile)
			os.Exit(1)
		}

		slog.Info("build vars loaded from file", "file", buildVarsFile, "varsCount", len(buildVars))
	}

	if err := executeBuild(config, buildVars); err != nil {
		slog.Error("build failed", "error", err)

		if hint := failureHint(err); hint != "" {
			fmt.Println(hint)
		}

		os.Exit(1)
	}
}

func failureHint(err error) string {
	message := err.Error()

	switch {
	case strings.Contains(message, "exit code: 137"):
		return "A build step was killed (exit code 137), which almost always means it ran out of memory. Lower the step's memory use, for example NODE_OPTIONS=--max-old-space-size=2048 for a Node build, and deploy again."
	case strings.Contains(message, "buildkit solve failed") && (strings.Contains(message, "EOF") || strings.Contains(message, "code = Unavailable") || strings.Contains(message, "connection reset")):
		return "The build daemon stopped during this build. That usually means a build on it ran out of memory, this one or another running at the same time. Deploy again. If it fails the same way, the build needs less memory, for example NODE_OPTIONS=--max-old-space-size=2048 for a Node build."
	case strings.Contains(message, "buildkit not ready") || strings.Contains(message, "git checkout timed out"):
		return "This failed on the platform's side, before your code was built. Deploy again."
	case strings.Contains(message, "clone failed"):
		return "Fetching the source failed before your code was built. The git output above says why: a network error is on the platform's side, so deploy again; an unknown commit or a repository the GitHub App can no longer read is not."
	}

	return ""
}
