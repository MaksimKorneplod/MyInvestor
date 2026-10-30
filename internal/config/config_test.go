package config

import (
	"log/slog"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/dbname")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if got.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want %q", got.HTTPAddr, ":8080")
	}
	if got.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want %v", got.LogLevel, slog.LevelInfo)
	}
}

func TestLoadEnvironmentValues(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/dbname")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if got.HTTPAddr != ":9090" {
		t.Errorf("HTTPAddr = %q, want %q", got.HTTPAddr, ":9090")
	}
	if got.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want %v", got.LogLevel, slog.LevelDebug)
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("LOG_LEVEL", "verbose")
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want an error for invalid LOG_LEVEL")
	}
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want an error for missing DATABASE_URL")
	}
}
