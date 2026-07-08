// Package config loads and provides Vordoc application configuration.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config holds unified configuration for Vordoc.
type Config struct {
	App     AppConfig     `yaml:"app"`
	Content ContentConfig `yaml:"content"`
	Auth    AuthConfig    `yaml:"auth"`
}

// AppConfig holds application-wide settings.
type AppConfig struct {
	HTTPPort          int           `yaml:"http_port"`
	ShutdownGracetime time.Duration `yaml:"shutdown_gracetime"`
	Logs              LogsConfig    `yaml:"logs"`
}

// ContentConfig holds filesystem content settings.
type ContentConfig struct {
	Root string `yaml:"root"`
}

// AuthConfig holds authentication settings for page-level access.
type AuthConfig struct {
	PageSecret string `yaml:"page_secret"`
}

// LogsConfig holds logging configuration.
type LogsConfig struct {
	Level string `yaml:"level"`
	Type  string `yaml:"type"` // pretty, json, text
}

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		App: AppConfig{
			HTTPPort:          12300,
			ShutdownGracetime: 10 * time.Second,
			Logs: LogsConfig{
				Level: "info",
				Type:  "pretty",
			},
		},
		Content: ContentConfig{
			Root: "./content",
		},
		Auth: AuthConfig{
			PageSecret: "CHANGE_ME",
		},
	}
}

// LoadFromEnv loads configuration from environment variables with defaults.
// If VORDOC_PAGE_SECRET is not set, a random secret is generated and a warning is logged.
func LoadFromEnv(logger *slog.Logger) Config {
	cfg := DefaultConfig()

	if port := os.Getenv("VORDOC_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.App.HTTPPort = p
		}
	}

	if root := os.Getenv("VORDOC_CONTENT"); root != "" {
		cfg.Content.Root = root
	}

	if level := os.Getenv("VORDOC_LOG_LEVEL"); level != "" {
		cfg.App.Logs.Level = level
	}

	if typ := os.Getenv("VORDOC_LOG_TYPE"); typ != "" {
		cfg.App.Logs.Type = typ
	}

	if grace := os.Getenv("VORDOC_SHUTDOWN_GRACE"); grace != "" {
		if d, err := time.ParseDuration(grace); err == nil {
			cfg.App.ShutdownGracetime = d
		}
	}

	if secret := os.Getenv("VORDOC_PAGE_SECRET"); secret != "" {
		cfg.Auth.PageSecret = secret
	} else {
		secret, err := generateRandomSecret(32)
		if err != nil {
			logger.Warn("failed to generate page secret", slog.String("error", err.Error()))
		} else {
			cfg.Auth.PageSecret = secret
			logger.Warn("VORDOC_PAGE_SECRET is not set; a random secret was generated. Set VORDOC_PAGE_SECRET explicitly for persistent cookies across restarts.")
		}
	}

	return cfg
}

// generateRandomSecret creates a random hex string.
func generateRandomSecret(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random secret: %w", err)
	}
	return hex.EncodeToString(b), nil
}
