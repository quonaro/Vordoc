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

const defaultLogoSize = 40
const defaultFontSize = 24
const defaultFontName = "FabergeDigital.otf"

// defaultHeader returns the built-in header defaults.
func defaultHeader() headerConfig {
	return headerConfig{
		Enable:   true,
		Elements: []string{"logo", "search", "theme-switch"},
		Title:    "Vordoc",
		Logo: &logoConfig{
			Path: defaultLogoFile,
			Size: defaultLogoSize,
		},
		Font: &fontConfig{
			Name: defaultFontName,
			Size: defaultFontSize,
		},
	}
}

// defaultTheme returns the built-in theme defaults.
func defaultTheme() themeConfig {
	return themeConfig{
		Default:     "system",
		AccentColor: "#3b82f6",
	}
}

// headerConfig mirrors domain.HeaderConfig for YAML parsing.
// Elements controls which header elements are rendered and in what order.
type headerConfig struct {
	Enable   bool        `yaml:"enable"`
	Elements []string    `yaml:"elements"`
	Title    string      `yaml:"title"`
	Logo     *logoConfig `yaml:"logo"`
	Font     *fontConfig `yaml:"font"`
}

// logoConfig mirrors domain.LogoConfig for YAML parsing.
type logoConfig struct {
	Path string `yaml:"path"`
	Size int    `yaml:"size"`
}

// fontConfig mirrors domain.FontConfig for YAML parsing.
type fontConfig struct {
	Name string `yaml:"name"`
	Size int    `yaml:"size"`
}

// themeConfig mirrors domain.ThemeConfig for YAML parsing.
type themeConfig struct {
	Default     string `yaml:"default"`
	AccentColor string `yaml:"accent-color"`
}

// rootPageConfig mirrors domain.RootPageConfig for YAML parsing.
type rootPageConfig struct {
	Enable bool   `yaml:"enable"`
	Title  string `yaml:"title"`
}

// siteConfig holds the root content configuration.
type siteConfig struct {
	Root    *rootPageConfig `yaml:"root"`
	Favicon string          `yaml:"favicon"`
	Header  *headerConfig   `yaml:"header"`
	Theme   *themeConfig    `yaml:"theme"`
}

// loadSiteConfig reads the root config.yaml from the content directory.
// If the file does not exist, a default configuration is returned.
func loadSiteConfig(root string) (siteConfig, error) {
	var cfg siteConfig

	path := filepath.Join(root, "config.yaml")
	data, err := os.ReadFile(path) // #nosec G304 — path is built internally
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

// fillHeaderDefaults copies the provided header and fills any unset fields with
// built-in defaults. The source value is never mutated.
func fillHeaderDefaults(src headerConfig) headerConfig {
	d := defaultHeader()
	if src.Elements == nil {
		src.Elements = d.Elements
	}
	if src.Title == "" {
		src.Title = d.Title
	}
	if src.Logo == nil {
		src.Logo = d.Logo
	} else {
		if src.Logo.Path == "" {
			src.Logo.Path = d.Logo.Path
		}
		if src.Logo.Size == 0 {
			src.Logo.Size = d.Logo.Size
		}
	}
	if src.Font == nil {
		src.Font = d.Font
	} else {
		if src.Font.Name == "" {
			src.Font.Name = d.Font.Name
		}
		if src.Font.Size == 0 {
			src.Font.Size = d.Font.Size
		}
	}
	return src
}

// defaultRootPage returns the built-in root page defaults.
func defaultRootPage() rootPageConfig {
	return rootPageConfig{
		Enable: true,
		Title:  "Vordoc",
	}
}

// defaultSiteConfig returns the root configuration used when config.yaml is missing.
func defaultSiteConfig() siteConfig {
	d := defaultTheme()
	h := defaultHeader()
	return siteConfig{
		Root:   &rootPageConfig{Enable: true, Title: "Vordoc"},
		Header: &h,
		Theme:  &themeConfig{Default: d.Default, AccentColor: d.AccentColor},
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
	rootHeader := resolveRootConfig(rootCfg).Header

	if cfg.Header == nil && rootHeader != nil {
		return &domain.HeaderConfig{
			Enable:   rootHeader.Enable,
			Elements: rootHeader.Elements,
			Title:    rootHeader.Title,
			Logo: &domain.LogoConfig{
				Path: fmt.Sprintf("/api/v1/logo?doc=%s", name),
				Size: rootHeader.Logo.Size,
			},
			Font: rootHeader.Font,
		}
	}

	h := fillHeaderDefaults(*cfg.Header)
	return &domain.HeaderConfig{
		Enable:   h.Enable,
		Elements: h.Elements,
		Title:    h.Title,
		Logo: &domain.LogoConfig{
			Path: fmt.Sprintf("/api/v1/logo?doc=%s", name),
			Size: h.Logo.Size,
		},
		Font: &domain.FontConfig{
			Name: h.Font.Name,
			Size: h.Font.Size,
		},
	}
}

// fillThemeDefaults copies the provided theme and fills any unset fields with
// built-in defaults. The source value is never mutated.
func fillThemeDefaults(src themeConfig) themeConfig {
	d := defaultTheme()
	if src.Default == "" {
		src.Default = d.Default
	}
	if src.AccentColor == "" {
		src.AccentColor = d.AccentColor
	}
	return src
}

// fillRootPageDefaults copies the provided root page config and fills any unset
// fields with built-in defaults. The source value is never mutated.
func fillRootPageDefaults(src rootPageConfig) rootPageConfig {
	d := defaultRootPage()
	if !src.Enable && src.Title == "" {
		src.Enable = d.Enable
	}
	if src.Title == "" {
		src.Title = d.Title
	}
	return src
}

// resolveRootConfig applies defaults and exposes the logo through the API endpoint.
func resolveRootConfig(cfg siteConfig) domain.RootConfig {
	h := defaultHeader()
	if cfg.Header != nil {
		h = fillHeaderDefaults(*cfg.Header)
	}

	t := defaultTheme()
	if cfg.Theme != nil {
		t = fillThemeDefaults(*cfg.Theme)
	}

	r := defaultRootPage()
	if cfg.Root != nil {
		r = fillRootPageDefaults(*cfg.Root)
	}

	favicon := cfg.Favicon
	if favicon == "" {
		favicon = "/favicon.ico"
	}

	return domain.RootConfig{
		Root: domain.RootPageConfig{
			Enable: r.Enable,
			Title:  r.Title,
		},
		Favicon: favicon,
		Header: &domain.HeaderConfig{
			Enable:   h.Enable,
			Elements: h.Elements,
			Title:    h.Title,
			Logo: &domain.LogoConfig{
				Path: "/api/v1/logo",
				Size: h.Logo.Size,
			},
			Font: &domain.FontConfig{
				Name: h.Font.Name,
				Size: h.Font.Size,
			},
		},
		Theme: &domain.ThemeConfig{
			Default:     t.Default,
			AccentColor: t.AccentColor,
		},
	}
}
