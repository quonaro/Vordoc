package content

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestProvider_GetFaviconPath_relative_to_content_root(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "config.yaml"), "favicon: \"favicon.ico\"\n")
	mustWrite(t, filepath.Join(root, "favicon.ico"), "icon")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	path, err := p.GetFaviconPath(context.Background())
	must(t, err)

	want := filepath.Join(root, "favicon.ico")
	if path != want {
		t.Errorf("GetFaviconPath = %q, want %q", path, want)
	}
}

func TestProvider_GetFaviconPath_leading_slash_is_content_root(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "config.yaml"), "favicon: \"/images/icon.ico\"\n")
	must(t, os.MkdirAll(filepath.Join(root, "images"), 0o755))
	mustWrite(t, filepath.Join(root, "images", "icon.ico"), "icon")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	path, err := p.GetFaviconPath(context.Background())
	must(t, err)

	want := filepath.Join(root, "images", "icon.ico")
	if path != want {
		t.Errorf("GetFaviconPath = %q, want %q", path, want)
	}
}

func TestProvider_GetFaviconPath_defaults_when_unset(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "config.yaml"), "root:\n  title: Test\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	path, err := p.GetFaviconPath(context.Background())
	must(t, err)

	want := filepath.Join(root, "favicon.ico")
	if path != want {
		t.Errorf("GetFaviconPath = %q, want %q", path, want)
	}
}

func TestProvider_GetFaviconPath_rejects_traversal(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "config.yaml"), "favicon: \"../secret.ico\"\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if _, err := p.GetFaviconPath(context.Background()); err == nil {
		t.Error("expected error for favicon path escaping the content root")
	}
}
