package content

import (
	"fmt"
	"path/filepath"
	"strings"
)

// safeContentPath joins name onto base and verifies that the result stays
// inside the content root, preventing path traversal.
func (p *Provider) safeContentPath(base, name string) (string, error) {
	path := filepath.Join(base, name)
	rel, err := filepath.Rel(p.root, path)
	if err != nil {
		return "", fmt.Errorf("invalid content path: %w", err)
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("content path escapes content root")
	}

	return filepath.Join(p.root, rel), nil
}
