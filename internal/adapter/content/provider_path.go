package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"vordoc/internal/domain"
)

// docPath returns the filesystem path for a documentation directory,
// rejecting names that would escape the content root.
func (p *Provider) docPath(name string) (string, error) {
	docPath := filepath.Join(p.root, name)
	rel, err := filepath.Rel(p.root, docPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%w: %s", domain.ErrDocNotFound, name)
	}
	return docPath, nil
}

// pagePathInsideDoc reports whether the resolved page file stays inside docPath.
func pagePathInsideDoc(docPath, pageFile string) bool {
	rel, err := filepath.Rel(docPath, pageFile)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

// resolvePageFile resolves the filesystem path for a page inside a documentation.
// It tries, in order: <pagePath>.md, <pagePath>/index.md, <pagePath>/info.md.
// Paths that escape the documentation directory are rejected.
// It returns the resolved page file path and the documentation directory path.
func (p *Provider) resolvePageFile(docName, pagePath string) (string, string, error) {
	docPath, err := p.docPath(docName)
	if err != nil {
		return "", "", fmt.Errorf("%w: %s/%s", domain.ErrPageNotFound, docName, pagePath)
	}

	candidates := []string{
		filepath.Join(docPath, pagePath+".md"),
		filepath.Join(docPath, pagePath, "index.md"),
		filepath.Join(docPath, pagePath, "info.md"),
	}

	for _, candidate := range candidates {
		if !pagePathInsideDoc(docPath, candidate) {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, docPath, nil
		}
	}

	return "", "", fmt.Errorf("%w: %s/%s", domain.ErrPageNotFound, docName, pagePath)
}
