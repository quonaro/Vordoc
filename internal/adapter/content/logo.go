package content

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

// LogoInfo describes a logo file to be served.
type LogoInfo struct {
	Path     string
	MIMEType string
}

// GetLogoPath returns the resolved logo file for the root site or a documentation.
// If the requested logo is missing, it falls back to the root logo.
func (p *Provider) GetLogoPath(_ context.Context, doc string) (string, error) {
	rootLogo, err := p.rootLogoPath()
	if err != nil {
		return "", err
	}
	if doc == "" {
		return rootLogo, nil
	}

	docLogo, err := p.docLogoPath(doc)
	if err != nil {
		return rootLogo, nil
	}
	if docLogo == "" {
		return rootLogo, nil
	}

	return docLogo, nil
}

// GetLogoInfo returns the logo file path and MIME type for a doc or the root site.
func (p *Provider) GetLogoInfo(ctx context.Context, doc string) (LogoInfo, error) {
	path, err := p.GetLogoPath(ctx, doc)
	if err != nil {
		return LogoInfo{}, err
	}
	return LogoInfo{Path: path, MIMEType: logoMimeType(path)}, nil
}

// rootLogoPath resolves the logo file declared in the root site config.
func (p *Provider) rootLogoPath() (string, error) {
	cfg, err := loadSiteConfig(p.root)
	if err != nil {
		return "", err
	}

	logo := defaultLogoFile
	if cfg.Header != nil && cfg.Header.Logo != nil && cfg.Header.Logo.Path != "" {
		logo = cfg.Header.Logo.Path
	}

	return p.resolveLogoPath("", logo)
}

// docLogoPath resolves the logo file declared in a doc config.
// An empty string with a nil error means the doc has a header section but no logo.
func (p *Provider) docLogoPath(doc string) (string, error) {
	cfg, err := loadDocConfig(filepath.Join(p.root, doc, "config.yaml"))
	if err != nil {
		return "", err
	}
	if cfg.Header == nil {
		return "", nil
	}
	if cfg.Header.Logo == nil || cfg.Header.Logo.Path == "" {
		return "", nil
	}
	return p.resolveLogoPath(doc, cfg.Header.Logo.Path)
}

// resolveLogoPath cleans a logo path relative to the content root and prevents traversal.
// A leading '/' in filename is treated as relative to the content root, regardless of doc.
func (p *Provider) resolveLogoPath(doc, filename string) (string, error) {
	base := p.root
	if doc != "" && !strings.HasPrefix(filename, "/") {
		base = filepath.Join(p.root, doc)
	}
	if filename == "" {
		filename = defaultLogoFile
	}

	path := filepath.Join(base, filename)
	rel, err := filepath.Rel(p.root, path)
	if err != nil {
		return "", fmt.Errorf("invalid logo path: %w", err)
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("logo path escapes content root")
	}

	return filepath.Join(p.root, rel), nil
}

// logoMimeType returns the MIME type for a logo file based on its extension.
func logoMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".svg" {
		return "image/svg+xml"
	}
	mt := mime.TypeByExtension(ext)
	if mt == "" {
		return "application/octet-stream"
	}
	return mt
}
