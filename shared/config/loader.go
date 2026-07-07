package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader loads YAML configuration with auto-creation support.
type Loader struct {
	logger *slog.Logger
}

// NewLoader creates a new config loader.
func NewLoader(logger *slog.Logger) *Loader {
	return &Loader{logger: logger}
}

// envVarRegex matches ${VAR} and ${VAR:-default} patterns.
var envVarRegex = regexp.MustCompile(`\$\{([^}]+)}`)

// expandEnv replaces ${VAR} and ${VAR:-default} with environment variable values.
func expandEnv(input string) string {
	return envVarRegex.ReplaceAllStringFunc(input, func(match string) string {
		content := match[2 : len(match)-1] // Remove ${ and }

		// Check for default value syntax: ${VAR:-default}
		if idx := strings.Index(content, ":-"); idx != -1 {
			varName := content[:idx]
			defaultValue := content[idx+2:]
			if val := os.Getenv(varName); val != "" {
				return val
			}
			return defaultValue
		}

		// Simple ${VAR} syntax
		if val := os.Getenv(content); val != "" {
			return val
		}
		return match // Return original if not found
	})
}

// LoadOrCreate loads config from path, creating default if missing.
func (l *Loader) LoadOrCreate(path string, defaults *Config) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		l.logger.Info("config file not found, creating default", slog.String("path", absPath))
		if err := l.createDefault(absPath, defaults); err != nil {
			return fmt.Errorf("creating default config: %w", err)
		}
		l.logger.Info("default config created", slog.String("path", absPath))
	} else if err != nil {
		return fmt.Errorf("checking config file: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	// Expand environment variables before parsing
	expanded := expandEnv(string(data))

	if err := yaml.Unmarshal([]byte(expanded), defaults); err != nil {
		return fmt.Errorf("parsing config YAML: %w", err)
	}

	return nil
}

// createDefault writes default config to file with generated secrets.
func (l *Loader) createDefault(path string, config *Config) error {
	// Auto-generate page secret
	if config.Auth.PageSecret == "CHANGE_ME" {
		secret, err := generateRandomSecret(32)
		if err != nil {
			return fmt.Errorf("generating page secret: %w", err)
		}
		config.Auth.PageSecret = secret
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling default config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing default config: %w", err)
	}

	return nil
}

// generateRandomSecret creates a random hex string.
func generateRandomSecret(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random secret: %w", err)
	}
	return hex.EncodeToString(b), nil
}
