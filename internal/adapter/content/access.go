package content

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// groupConfig holds access rules for a directory group.
type groupConfig struct {
	Access       string `yaml:"access"`
	PasswordHash string `yaml:"password_hash"`
}

// loadAccessConfig reads access.yaml from a directory.
// It returns the config, a bool indicating whether the file existed, and any error.
func loadAccessConfig(dir string) (groupConfig, bool, error) {
	path := filepath.Join(dir, "access.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return groupConfig{}, false, nil
		}
		return groupConfig{}, false, fmt.Errorf("reading access.yaml: %w", err)
	}

	var cfg groupConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return groupConfig{}, false, fmt.Errorf("parsing access.yaml: %w", err)
	}
	if cfg.Access == "" {
		cfg.Access = "public"
	}

	return cfg, true, nil
}

// resolveAccess determines the access level and password hash for a page.
// Priority: frontmatter > nearest access.yaml (walking up) > default "public".
func resolveAccess(docPath string, pageFile string, fm map[string]any) (string, string) {
	// 1. Frontmatter override
	if access := getString(fm, "access", ""); access != "" {
		return access, getString(fm, "password_hash", "")
	}

	// 2. Walk up from page directory to doc root
	dir := filepath.Dir(pageFile)
	docPath = filepath.Clean(docPath)
	for {
		cfg, found, err := loadAccessConfig(dir)
		if err == nil && found {
			return cfg.Access, cfg.PasswordHash
		}
		if dir == docPath {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "public", ""
}

// docConfig holds per-doc metadata.
type docConfig struct {
	Title  string        `yaml:"title"`
	Header *headerConfig `yaml:"header"`
}

// loadDocConfig reads config.yaml from a doc directory.
func loadDocConfig(path string) (docConfig, error) {
	var cfg docConfig

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config.yaml: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config.yaml: %w", err)
	}

	return cfg, nil
}
