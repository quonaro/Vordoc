package content

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestProvider_GetDoc_header_defaults_to_enabled(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\nheader:\n  title: Doc Header\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	doc, err := p.GetDoc(context.Background(), "doc")
	must(t, err)

	if doc.Header == nil {
		t.Fatal("doc.Header is nil")
	}
	if !doc.Header.Enable {
		t.Errorf("doc.Header.Enable = false, want true when header is present without explicit enable")
	}
	if doc.Header.Title != "Doc Header" {
		t.Errorf("doc.Header.Title = %q, want %q", doc.Header.Title, "Doc Header")
	}
}

func TestProvider_GetDoc_header_explicit_disable(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\nheader:\n  enable: false\n  title: Doc Header\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	doc, err := p.GetDoc(context.Background(), "doc")
	must(t, err)

	if doc.Header == nil {
		t.Fatal("doc.Header is nil")
	}
	if doc.Header.Enable {
		t.Errorf("doc.Header.Enable = true, want false when header.enable is explicitly false")
	}
	if doc.Header.Title != "Doc Header" {
		t.Errorf("doc.Header.Title = %q, want %q", doc.Header.Title, "Doc Header")
	}
}

func TestProvider_GetLogoPath_leading_slash_is_content_root(t *testing.T) {
	root := t.TempDir()

	must(t, os.WriteFile(filepath.Join(root, "2.svg"), []byte("<svg></svg>"), 0o644))

	docRoot := filepath.Join(root, "welcome")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\nheader:\n  title: Welcome\n  logo:\n    path: /2.svg\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	path, err := p.GetLogoPath(context.Background(), "welcome")
	must(t, err)

	want := filepath.Join(root, "2.svg")
	if path != want {
		t.Errorf("GetLogoPath = %q, want %q", path, want)
	}
}
