// Command builder is the in-pod build runner that the conductor's
// Build Jobs invoke. It clones the user's source repo, generates a
// railpack plan, executes the plan against BuildKit, and pushes the
// resulting image to the platform's OCI registry.
//
// This binary has no API surface — it reads its inputs from
// environment variables (set by the conductor when it constructs the
// Job spec) and runs to completion.
package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/pkg/logger"
)

type Config struct {
	BuildID      string   `envconfig:"BUILD_ID" required:"true"`
	SourceURL    string   `envconfig:"BUILD_SOURCE_URL" required:"true"`
	GitRef       string   `envconfig:"BUILD_GIT_REF" required:"true"`
	ContextPath  string   `envconfig:"BUILD_CONTEXT_PATH"`
	TargetRefs   []string `envconfig:"BUILD_TARGET_REFS" required:"true"`
	BuildkitAddr string   `envconfig:"BUILDKIT_ADDR" required:"true"`
	GitHubToken  string   `envconfig:"GITHUB_TOKEN"`
}

func main() {
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

	if err := executeBuild(config); err != nil {
		slog.Error("build failed", "error", err)

		os.Exit(1)
	}
}
