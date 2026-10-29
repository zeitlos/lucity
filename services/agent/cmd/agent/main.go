package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
	_ "time/tzdata"

	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/agent/internal/api"
	"github.com/zeitlos/lucity/services/agent/internal/config"
	"github.com/zeitlos/lucity/services/agent/internal/harness"
	"github.com/zeitlos/lucity/services/agent/internal/runner"
)

type Config struct {
	LogLevel        string        `envconfig:"LOG_LEVEL" default:"info"`
	ListenAddress   string        `envconfig:"LISTEN_ADDRESS" default:":8080"`
	ConfigDir       string        `envconfig:"CONFIG_DIR" default:"/etc/agent"`
	PolicyDir       string        `envconfig:"POLICY_DIR" default:"/etc/harness"`
	DataDir         string        `envconfig:"DATA_DIR" default:"/data"`
	GatewayURL      string        `envconfig:"AGENT_GATEWAY_URL" required:"true"`
	GatewayKey      string        `envconfig:"AGENT_GATEWAY_KEY" required:"true"`
	HarnessCommand  []string      `envconfig:"HARNESS_COMMAND"`
	ApprovalTimeout time.Duration `envconfig:"APPROVAL_TIMEOUT" default:"15m"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(cfg.LogLevel)

	agent, err := config.Load(cfg.ConfigDir, config.Gateway{URL: cfg.GatewayURL, Key: cfg.GatewayKey})
	if err != nil {
		slog.Error("failed to load agent config", "error", err)
		os.Exit(1)
	}

	paths := harness.Paths{
		Home:    filepath.Join(cfg.DataDir, "home"),
		Workdir: filepath.Join(cfg.DataDir, "workspace"),
		Policy:  cfg.PolicyDir,
	}
	for _, dir := range []string{paths.Home, paths.Workdir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			slog.Error("failed to create directory", "dir", dir, "error", err)
			os.Exit(1)
		}
	}

	h, err := harness.New(agent, paths)
	if err != nil {
		slog.Error("failed to set up harness", "error", err)
		os.Exit(1)
	}
	if err := h.Configure(); err != nil {
		slog.Error("failed to write harness config", "harness", agent.Harness, "error", err)
		os.Exit(1)
	}
	if len(os.Args) > 1 && os.Args[1] == "configure" {
		slog.Info("harness configured", "harness", agent.Harness)
		return
	}

	connectors := make([]string, 0, len(agent.Connectors))
	for _, connector := range agent.Connectors {
		connectors = append(connectors, connector.Name)
	}
	r, err := runner.New(string(agent.Harness), h, cfg.HarnessCommand, connectors, paths.Workdir, filepath.Join(cfg.DataDir, "sessions"), cfg.ApprovalTimeout)
	if err != nil {
		slog.Error("failed to open sessions", "error", err)
		os.Exit(1)
	}

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, r, api.New(cfg.ListenAddress, r))
}
