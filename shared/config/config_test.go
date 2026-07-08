package config

import (
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestLoadFromEnv_uses_defaults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := LoadFromEnv(logger)

	if cfg.App.HTTPPort != 12300 {
		t.Errorf("HTTPPort = %d, want 12300", cfg.App.HTTPPort)
	}
	if cfg.Content.Root != "./content" {
		t.Errorf("Content.Root = %q, want ./content", cfg.Content.Root)
	}
	if cfg.App.Logs.Level != "info" {
		t.Errorf("Logs.Level = %q, want info", cfg.App.Logs.Level)
	}
	if cfg.App.Logs.Type != "pretty" {
		t.Errorf("Logs.Type = %q, want pretty", cfg.App.Logs.Type)
	}
	if cfg.App.ShutdownGracetime != 10*time.Second {
		t.Errorf("ShutdownGracetime = %v, want 10s", cfg.App.ShutdownGracetime)
	}
	if cfg.Auth.PageSecret == "" {
		t.Error("PageSecret should be generated when not set")
	}
}

func TestLoadFromEnv_reads_overrides(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Setenv("VORDOC_PORT", "8080")
	t.Setenv("VORDOC_CONTENT", "./docs")
	t.Setenv("VORDOC_LOG_LEVEL", "debug")
	t.Setenv("VORDOC_LOG_TYPE", "json")
	t.Setenv("VORDOC_SHUTDOWN_GRACE", "5s")
	t.Setenv("VORDOC_PAGE_SECRET", "my-secret")

	cfg := LoadFromEnv(logger)

	if cfg.App.HTTPPort != 8080 {
		t.Errorf("HTTPPort = %d, want 8080", cfg.App.HTTPPort)
	}
	if cfg.Content.Root != "./docs" {
		t.Errorf("Content.Root = %q, want ./docs", cfg.Content.Root)
	}
	if cfg.App.Logs.Level != "debug" {
		t.Errorf("Logs.Level = %q, want debug", cfg.App.Logs.Level)
	}
	if cfg.App.Logs.Type != "json" {
		t.Errorf("Logs.Type = %q, want json", cfg.App.Logs.Type)
	}
	if cfg.App.ShutdownGracetime != 5*time.Second {
		t.Errorf("ShutdownGracetime = %v, want 5s", cfg.App.ShutdownGracetime)
	}
	if cfg.Auth.PageSecret != "my-secret" {
		t.Errorf("PageSecret = %q, want my-secret", cfg.Auth.PageSecret)
	}
}

func TestLoadFromEnv_ignores_invalid_values(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Setenv("VORDOC_PORT", "not-a-number")
	t.Setenv("VORDOC_SHUTDOWN_GRACE", "not-a-duration")

	cfg := LoadFromEnv(logger)

	if cfg.App.HTTPPort != 12300 {
		t.Errorf("HTTPPort = %d, want default 12300", cfg.App.HTTPPort)
	}
	if cfg.App.ShutdownGracetime != 10*time.Second {
		t.Errorf("ShutdownGracetime = %v, want default 10s", cfg.App.ShutdownGracetime)
	}
}

func TestLoadFromEnv_generates_different_secrets(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Ensure no secret is inherited from the environment.
	t.Setenv("VORDOC_PAGE_SECRET", "")

	cfg1 := LoadFromEnv(logger)
	cfg2 := LoadFromEnv(logger)

	if cfg1.Auth.PageSecret == "" {
		t.Error("first generated secret is empty")
	}
	if cfg1.Auth.PageSecret == cfg2.Auth.PageSecret {
		t.Error("generated secrets should be random")
	}
}
