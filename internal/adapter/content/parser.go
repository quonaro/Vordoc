package content

import (
	"bytes"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// parseFrontmatter extracts YAML frontmatter and body from markdown content.
func parseFrontmatter(data []byte) (frontmatter map[string]any, body string, err error) {
	data = bytes.TrimSpace(data)

	if !bytes.HasPrefix(data, []byte("---")) {
		return map[string]any{}, string(data), nil
	}

	parts := bytes.SplitN(data, []byte("---"), 3)
	if len(parts) < 3 {
		return map[string]any{}, string(data), nil
	}

	fm := make(map[string]any)
	if err := yaml.Unmarshal(parts[1], &fm); err != nil {
		return nil, "", fmt.Errorf("parsing frontmatter: %w", err)
	}

	body = string(bytes.TrimSpace(parts[2]))
	return fm, body, nil
}

// getString extracts a string value from frontmatter with a fallback.
func getString(m map[string]any, key, fallback string) string {
	v, ok := m[key]
	if !ok {
		return fallback
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return fallback
}

// getInt extracts an int value from frontmatter with a fallback.
func getInt(m map[string]any, key string, fallback int) int {
	v, ok := m[key]
	if !ok {
		return fallback
	}

	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		if i, err := strconv.Atoi(n); err == nil {
			return i
		}
	}
	return fallback
}

// getBool extracts a bool value from frontmatter with a fallback.
func getBool(m map[string]any, key string, fallback bool) bool {
	v, ok := m[key]
	if !ok {
		return fallback
	}
	b, ok := v.(bool)
	if ok {
		return b
	}
	return fallback
}
