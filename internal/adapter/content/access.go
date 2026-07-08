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
// A node with access: password and no password_hash inherits the hash and scope of the
// nearest ancestor that has a password_hash. access: none and access: public stop inheritance.
func resolveAccessInfo(docPath string, pageFile string, fm map[string]any) domain.AccessInfo {
	// 1. Frontmatter override.
	if access := getString(fm, "access", ""); access != "" {
		if access == "none" {
			return domain.AccessInfo{Access: "public"}
		}
		if hash := getString(fm, "password_hash", ""); hash != "" {
			rel, _ := filepath.Rel(docPath, pageFile)
			rel = filepath.ToSlash(rel)
			scope := strings.TrimSuffix(rel, ".md")
			if scope == "index" {
				scope = ""
			}
			return domain.AccessInfo{
				Access:       "password",
				PasswordHash: hash,
				Scope:        scope,
			}
		}
		// Frontmatter asks for password but has no hash: inherit from an ancestor.
		rel, _ := filepath.Rel(docPath, pageFile)
		rel = filepath.ToSlash(rel)
		originalScope := strings.TrimSuffix(rel, ".md")
		if originalScope == "index" {
			originalScope = ""
		}
		return inheritPasswordHash(docPath, filepath.Dir(pageFile), originalScope)
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
				if cfg.Access == "public" {
					return domain.AccessInfo{Access: "public"}
				}
				first = domain.AccessInfo{
					Access:       cfg.Access,
					PasswordHash: cfg.PasswordHash,
					Scope:        scope,
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

// inheritPasswordHash walks up from childDir to the doc root looking for a password_hash.
// If a public ancestor is found first, it stops and returns the original scope with no hash.
// If no ancestor has a hash, it returns the original scope with no hash.
func inheritPasswordHash(docPath string, childDir string, originalScope string) domain.AccessInfo {
	docPath = filepath.Clean(docPath)
	dir := childDir

	for {
		cfg, found, err := loadAccessConfig(dir)
		if err == nil && found {
			if cfg.Access == "public" {
				return domain.AccessInfo{Access: "password", Scope: originalScope}
			}
			if cfg.PasswordHash != "" {
				rel, _ := filepath.Rel(docPath, dir)
				scope := filepath.ToSlash(rel)
				if scope == "." {
					scope = ""
				}
				return domain.AccessInfo{
					Access:       "password",
					PasswordHash: cfg.PasswordHash,
					Scope:        scope,
				}
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

	return domain.AccessInfo{Access: "password", Scope: originalScope}
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
