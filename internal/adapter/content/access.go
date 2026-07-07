package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"vordoc/internal/domain"

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
	if cfg.Access == "" || cfg.Access == "none" {
		cfg.Access = "public"
	}

	return cfg, true, nil
}

// resolveAccess determines the access level and password hash for a page.
// Priority: frontmatter > nearest access.yaml (walking up) > default "public".
func resolveAccess(docPath string, pageFile string, fm map[string]any) (string, string) {
	info := resolveAccessInfo(docPath, pageFile, fm)
	return info.Access, info.PasswordHash
}

// resolveAccessInfo returns the effective access rule, walking up the directory tree.
// A node with access: password and no password_hash inherits the hash of the nearest
// ancestor that has a password_hash. access: none and access: public stop inheritance.
func resolveAccessInfo(docPath string, pageFile string, fm map[string]any) domain.AccessInfo {
	// 1. Frontmatter override: the page itself owns the rule.
	if access := getString(fm, "access", ""); access != "" {
		if access == "none" {
			access = "public"
		}
		info := domain.AccessInfo{
			Access:       access,
			PasswordHash: getString(fm, "password_hash", ""),
		}
		if access == "password" {
			rel, _ := filepath.Rel(docPath, pageFile)
			rel = filepath.ToSlash(rel)
			info.Scope = strings.TrimSuffix(rel, ".md")
			if info.Scope == "index" {
				info.Scope = ""
			}
		}
		return info
	}

	// 2. Walk up from page directory to doc root.
	dir := filepath.Dir(pageFile)
	docPath = filepath.Clean(docPath)

	var first domain.AccessInfo
	firstSet := false

	for {
		cfg, found, err := loadAccessConfig(dir)
		if err == nil && found {
			rel, _ := filepath.Rel(docPath, dir)
			scope := filepath.ToSlash(rel)
			if scope == "." {
				scope = ""
			}

			if !firstSet {
				firstSet = true
				first = domain.AccessInfo{
					Access:       cfg.Access,
					PasswordHash: cfg.PasswordHash,
					Scope:        scope,
				}
				if cfg.Access == "public" {
					return first
				}
				if cfg.PasswordHash != "" {
					return first
				}
			}

			if first.Access == "password" && cfg.PasswordHash != "" {
				return domain.AccessInfo{
					Access:       "password",
					PasswordHash: cfg.PasswordHash,
					Scope:        scope,
				}
			}
			if first.Access == "password" && cfg.Access == "public" {
				return first
			}
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

	if first.Access == "password" {
		return first
	}
	return domain.AccessInfo{Access: "public"}
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
