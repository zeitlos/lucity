package main

import (
	"log/slog"
	"os"

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
	BuildkitAddr string   `envconfig:"BUILDKIT_ADDR" required:"true"`
	GitHubToken  string   `envconfig:"GITHUB_TOKEN"`

	BuildkitTLSCACert string `envconfig:"BUILDKIT_TLS_CA"`
	BuildkitTLSCert   string `envconfig:"BUILDKIT_TLS_CERT"`
	BuildkitTLSKey    string `envconfig:"BUILDKIT_TLS_KEY"`
	BuildkitServer    string `envconfig:"BUILDKIT_SERVER_NAME"`
}

const buildVarsFile = "/etc/lucity/.env"

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

		os.Exit(1)
	}
}
