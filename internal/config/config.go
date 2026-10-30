package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr    string
	LogLevel    slog.Level
	DatabaseURL string
}

func Load() (Config, error) {

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "", "info":
		level = slog.LevelInfo
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return Config{}, fmt.Errorf("invalid log level: %s", os.Getenv("LOG_LEVEL"))
	}

	return Config{
		HTTPAddr:    addr,
		LogLevel:    level,
		DatabaseURL: databaseURL,
	}, nil
}
