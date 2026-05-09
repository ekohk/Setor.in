// Package logger configures slog for structured JSON logging.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a slog.Logger writing JSON to stdout with the given level.
// Sets the returned logger as the default.
func New(level string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(level),
		AddSource: false,
	}
	h := slog.NewJSONHandler(os.Stdout, opts)
	l := slog.New(h)
	slog.SetDefault(l)
	return l
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
