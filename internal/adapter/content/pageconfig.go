package content

import (
	"os"
	"path/filepath"
)

// loadPageConfigEntry finds the nearest pages: entry for pageFile relative to
// docPath. It walks up from the page's directory to the doc root, loading
// config.yaml at each level and checking if the page's relative path (relative
// to that config.yaml's directory) is present in the pages: map.
// Returns the pageConfig and true if found, or a zero value and false otherwise.
func loadPageConfigEntry(docPath, pageFile string) (pageConfig, bool) {
	docPath = filepath.Clean(docPath)
	dir := filepath.Dir(pageFile)

	for {
		cfgPath := filepath.Join(dir, "config.yaml")
		if _, err := os.Stat(cfgPath); err == nil {
			cfg, err := loadDocConfig(cfgPath)
			if err == nil && cfg.Pages != nil {
				rel, _ := filepath.Rel(dir, pageFile)
				rel = filepath.ToSlash(rel)
				if pc, ok := cfg.Pages[rel]; ok {
					return pc, true
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
	return pageConfig{}, false
}

// mergePageConfig overlays non-zero pageConfig values onto a frontmatter map.
// Values from pageConfig take priority; frontmatter fills the gaps. The
// original map is not mutated.
func mergePageConfig(fm map[string]any, pc pageConfig) map[string]any {
	merged := make(map[string]any, len(fm)+6)
	for k, v := range fm {
		merged[k] = v
	}
	if pc.Title != "" {
		merged["title"] = pc.Title
	}
	if pc.Icon != "" {
		merged["icon"] = pc.Icon
	}
	if pc.Description != "" {
		merged["description"] = pc.Description
	}
	if pc.Order != 0 {
		merged["order"] = pc.Order
	}
	if pc.Show != nil {
		merged["show"] = *pc.Show
	}
	if pc.Access != "" {
		merged["access"] = pc.Access
	}
	if pc.PasswordHash != "" {
		merged["password_hash"] = pc.PasswordHash
	}
	return merged
}
