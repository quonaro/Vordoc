package content

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const uiTextFile = "text.json"

//go:embed default_text.json
var defaultTextData []byte

// defaultUIText returns the embedded UI text defaults.
func defaultUIText() (map[string]any, error) {
	var text map[string]any
	if err := json.Unmarshal(defaultTextData, &text); err != nil {
		return nil, fmt.Errorf("parsing default ui text: %w", err)
	}
	return text, nil
}

// mergeMaps recursively merges src into dst. dst is modified in place.
func mergeMaps(dst, src map[string]any) {
	for key, srcVal := range src {
		if srcMap, ok := srcVal.(map[string]any); ok {
			if dstMap, ok := dst[key].(map[string]any); ok {
				mergeMaps(dstMap, srcMap)
				continue
			}
		}
		dst[key] = srcVal
	}
}

// loadUIText reads the UI text configuration from the content directory and
// merges it over the embedded default text. If the file does not exist, the
// defaults are returned unchanged.
func loadUIText(root string) (map[string]any, error) {
	text, err := defaultUIText()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(root, uiTextFile)
	data, err := os.ReadFile(path) // #nosec G304 — path is built internally
	if err != nil {
		if os.IsNotExist(err) {
			return text, nil
		}
		return nil, fmt.Errorf("reading ui text: %w", err)
	}

	var overrides map[string]any
	if err := json.Unmarshal(data, &overrides); err != nil {
		return nil, fmt.Errorf("parsing ui text: %w", err)
	}

	mergeMaps(text, overrides)
	return text, nil
}

// GetUIText returns the UI text configuration for the frontend.
func (p *Provider) GetUIText(_ context.Context) (map[string]any, error) {
	return loadUIText(p.root)
}
