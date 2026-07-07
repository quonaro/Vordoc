package content

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"vordoc/internal/domain"

	"gopkg.in/yaml.v3"
)

const defaultLogoFile = "logotype.svg"

// headerConfig mirrors domain.HeaderConfig for YAML parsing.
type headerConfig struct {
	Enable bool   `yaml:"enable"`
	Title  string `yaml:"title"`
	Logo   string `yaml:"logo"`
}

// siteConfig holds the root content configuration.
type siteConfig struct {
	EnableRootPage bool          `yaml:"enable_root_page"`
	Header         *headerConfig `yaml:"header"`
}

// loadSiteConfig reads the root config.yaml from the content directory.
// If the file does not exist, a default configuration is returned.
func loadSiteConfig(root string) (siteConfig, error) {
	var cfg siteConfig

	path := filepath.Join(root, "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultSiteConfig(), nil
		}
		return cfg, fmt.Errorf("reading site config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing site config: %w", err)
	}

	return cfg, nil
}

// defaultSiteConfig returns the root configuration used when config.yaml is missing.
func defaultSiteConfig() siteConfig {
	return siteConfig{
		EnableRootPage: true,
		Header: &headerConfig{
			Enable: true,
			Title:  "Vordoc",
			Logo:   defaultLogoFile,
		},
	}
}

// GetRootConfig returns the root content configuration.
func (p *Provider) GetRootConfig(_ context.Context) (domain.RootConfig, error) {
	cfg, err := loadSiteConfig(p.root)
	if err != nil {
		return domain.RootConfig{}, err
	}
	return resolveRootConfig(cfg), nil
}

// resolveDocHeader returns the effective header for a documentation.
// A doc-level header replaces the root header entirely when present.
func (p *Provider) resolveDocHeader(name string, cfg docConfig) *domain.HeaderConfig {
	rootCfg, _ := loadSiteConfig(p.root)

	var h headerConfig
	if cfg.Header != nil {
		h = *cfg.Header
	} else if rootCfg.Header != nil {
		h = *rootCfg.Header
	}

	if h.Title == "" {
		h.Title = "Vordoc"
	}
	if h.Logo == "" {
		h.Logo = defaultLogoFile
	}

	return &domain.HeaderConfig{
		Enable: h.Enable,
		Title:  h.Title,
		Logo:   fmt.Sprintf("/api/v1/logo?doc=%s", name),
	}
}

// resolveRootConfig applies defaults and exposes the logo through the API endpoint.
func resolveRootConfig(cfg siteConfig) domain.RootConfig {
	h := cfg.Header
	if h == nil {
		h = &headerConfig{Enable: true}
	}
	if h.Title == "" {
		h.Title = "Vordoc"
	}
	if h.Logo == "" {
		h.Logo = defaultLogoFile
	}

	return domain.RootConfig{
		EnableRootPage: cfg.EnableRootPage,
		Header: &domain.HeaderConfig{
			Enable: h.Enable,
			Title:  h.Title,
			Logo:   "/api/v1/logo",
		},
	}
}
