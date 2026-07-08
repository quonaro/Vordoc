package content

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"vordoc/internal/domain"
)

// GetAssetAccess returns the effective access info for an asset path.
func (p *Provider) GetAssetAccess(ctx context.Context, docName string, assetPath string) (domain.AccessInfo, error) {
	docPath, err := p.docPath(docName)
	if err != nil {
		return domain.AccessInfo{}, err
	}

	fullPath, err := p.GetAssetPath(ctx, docName, assetPath)
	if err != nil {
		return domain.AccessInfo{}, err
	}

	return resolveAccessInfo(docPath, fullPath, nil), nil
}

// GetAssetPath resolves a static asset path inside a documentation directory.
func (p *Provider) GetAssetPath(ctx context.Context, docName string, assetPath string) (string, error) {
	docPath, err := p.docPath(docName)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(docPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: %s", domain.ErrDocNotFound, docName)
		}
		return "", fmt.Errorf("stat doc: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%w: %s is not a directory", domain.ErrDocNotFound, docName)
	}

	safePath := filepath.Clean(filepath.Join("/", filepath.ToSlash(assetPath)))
	fullPath := filepath.Join(docPath, safePath)

	absDoc, err := filepath.Abs(docPath)
	if err != nil {
		return "", fmt.Errorf("resolving doc path: %w", err)
	}
	absAsset, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("resolving asset path: %w", err)
	}
	if !strings.HasPrefix(absAsset, absDoc+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: asset path escapes doc directory", domain.ErrInvalidPath)
	}

	info, err = os.Stat(absAsset)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: %s", domain.ErrAssetNotFound, assetPath)
		}
		return "", fmt.Errorf("stat asset: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("asset path is a directory")
	}

	return absAsset, nil
}
