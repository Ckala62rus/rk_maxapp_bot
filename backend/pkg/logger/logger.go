// Package logger provides structured logging setup.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Config configures logger behavior.
type Config struct {
	Level string
	Env   string
}

// New creates a slog logger with JSON output and standard fields.
func New(cfg Config) *slog.Logger {
	// Map textual level to slog level.
	level := parseLevel(cfg.Level)
	// JSON handler is better for Loki parsing.
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: strings.EqualFold(cfg.Level, "debug"),
	})

	log := slog.New(handler)

	// Standard fields for observability.
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "maxapp"
	}
	appVersion := os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = "0.1.0"
	}

	return log.With(
		"app_name", appName,
		"app_version", appVersion,
		"env", cfg.Env,
	)
}

// parseLevel converts config string to slog.Level.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
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
