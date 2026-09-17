package content

import (
	"context"
	"path/filepath"
	"strings"
)

const defaultFaviconFile = "favicon.ico"

// resolveFavicon maps a favicon declared in the content root to the API endpoint
// that serves it. Absolute URLs and absolute paths are passed through unchanged;
// an empty value keeps the built-in frontend default.
func resolveFavicon(raw string) string {
	if raw == "" {
		return "/favicon.ico"
	}
	if strings.HasPrefix(raw, "/") || strings.Contains(raw, "://") {
		return raw
	}
	return "/api/v1/favicon"
}

// GetFaviconPath returns the resolved favicon file declared in the root config.
// The path is relative to the content root; a leading '/' is also treated as
// relative to the content root.
func (p *Provider) GetFaviconPath(_ context.Context) (string, error) {
	cfg, err := loadSiteConfig(p.root)
	if err != nil {
		return "", err
	}

	favicon := cfg.Favicon
	if favicon == "" {
		favicon = defaultFaviconFile
	}

	return p.safeContentPath(p.root, filepath.FromSlash(favicon))
}
