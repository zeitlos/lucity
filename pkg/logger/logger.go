package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

func Setup(level string) {
	enableColor := strings.ToLower(os.Getenv("LOG_COLOR")) == "true"
	handler := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLevel(level),
	}))

	if enableColor {
		handler = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level: parseLevel(level),
		}))

		return
	}

	slog.SetDefault(handler)
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
