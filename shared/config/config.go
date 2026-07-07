package config

import "time"

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
