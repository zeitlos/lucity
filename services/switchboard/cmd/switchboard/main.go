package main

import (
	"log/slog"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/switchboard/bridge"
	"github.com/zeitlos/lucity/services/switchboard/telegram"
)

type Config struct {
	LogLevel        string        `envconfig:"LOG_LEVEL" default:"info"`
	TelegramToken   string        `envconfig:"TELEGRAM_BOT_TOKEN" required:"true"`
	TelegramAPIURL  string        `envconfig:"TELEGRAM_API_URL" default:"https://api.telegram.org"`
	ChatID          int64         `envconfig:"TELEGRAM_CHAT_ID" required:"true"`
	AlertsTopicID   int64         `envconfig:"TELEGRAM_ALERTS_TOPIC_ID"`
	AllowedUserIDs  []int64       `envconfig:"TELEGRAM_ALLOWED_USER_IDS" required:"true"`
	AgentCommand    []string      `envconfig:"AGENT_COMMAND" default:"claude-agent-acp"`
	AgentWorkdir    string        `envconfig:"AGENT_WORKDIR" default:"/workspace"`
	MCPServersPath  string        `envconfig:"MCP_SERVERS_PATH" default:"/etc/switchboard/mcp-servers.json"`
	StatePath       string        `envconfig:"STATE_PATH" default:"/data/threads.json"`
	ApprovalTimeout time.Duration `envconfig:"APPROVAL_TIMEOUT" default:"15m"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	mcpServers, err := bridge.LoadMCPServers(config.MCPServersPath)
	if err != nil {
		slog.Error("failed to load MCP servers", "error", err)
		os.Exit(1)
	}

	threads, err := bridge.OpenThreads(config.StatePath)
	if err != nil {
		slog.Error("failed to open thread state", "error", err)
		os.Exit(1)
	}

	server := bridge.New(
		telegram.New(config.TelegramToken, telegram.WithAPIURL(config.TelegramAPIURL)),
		bridge.Forum{ChatID: config.ChatID, AlertsTopicID: config.AlertsTopicID},
		threads,
		bridge.Agent{Command: config.AgentCommand, Workdir: config.AgentWorkdir, MCPServers: mcpServers},
		bridge.WithAllowedUsers(config.AllowedUserIDs...),
		bridge.WithApprovalTimeout(config.ApprovalTimeout),
	)

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, server)
}
