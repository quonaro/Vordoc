package content

import (
	"fmt"
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
