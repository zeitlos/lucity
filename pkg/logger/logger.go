package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

// Setup configures the default slog logger. Output is JSON by default (for SigNoz / structured log ingestion).
// Set LOG_COLOR=true to enable colored output via tint (useful in interactive terminals).
func Setup(level string) {
	if strings.ToLower(os.Getenv("LOG_COLOR")) == "true" {
		slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level: parseLevel(level),
		})))
		return
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLevel(level),
	})))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
