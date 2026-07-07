package content

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestProvider_SearchPages(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nWelcome to the home page.\n")
	mustWrite(t, filepath.Join(docRoot, "guide.md"), "---\ntitle: Guide\n---\n# Getting started\n\nThis guide explains how to get started with the project.\n")

	guideDir := filepath.Join(docRoot, "guide")
	must(t, os.MkdirAll(guideDir, 0o755))
	mustWrite(t, filepath.Join(guideDir, "index.md"), "---\ntitle: Guide Index\n---\nOverview of the guide.\n")
	mustWrite(t, filepath.Join(guideDir, "advanced.md"), "---\ntitle: Advanced\n---\nAdvanced topics include configuration and deployment.\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	results, err := p.SearchPages(context.Background(), "doc", "advanced")
	must(t, err)

	if len(results) == 0 {
		t.Fatal("expected search results")
	}

	found := false
	for _, r := range results {
		if r.Title == "Advanced" && r.Path == "guide/advanced" {
			found = true
			if r.Snippet == "" {
				t.Errorf("expected non-empty snippet for Advanced page")
			}
		}
	}
	if !found {
		t.Errorf("Advanced page not found in results")
	}
}

func TestProvider_SearchPages_noResults(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome page\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	results, err := p.SearchPages(context.Background(), "doc", "nonexistent")
	must(t, err)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestProvider_SearchPages_emptyQuery(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome page\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	results, err := p.SearchPages(context.Background(), "doc", "")
	must(t, err)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(results))
	}
}

func TestProvider_SearchPages_docNotFound(t *testing.T) {
	root := t.TempDir()
	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	_, err := p.SearchPages(context.Background(), "missing", "query")
	if err == nil {
		t.Fatal("expected error for missing doc")
	}
}

func TestStripMarkdown(t *testing.T) {
	input := "# Title\n\nThis is **bold** and [a link](http://example.com).\n\n`code`\n\n- item one\n- item two\n"
	want := "Title This is bold and a link. code item one item two"
	got := stripMarkdown(input)
	if got != want {
		t.Errorf("stripMarkdown = %q, want %q", got, want)
	}
}

func TestSearchTerms(t *testing.T) {
	terms := searchTerms("Hello, World!")
	if len(terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(terms))
	}
	if terms[0] != "hello" || terms[1] != "world" {
		t.Errorf("terms = %v, want [hello world]", terms)
	}
}

func TestScoreTerms(t *testing.T) {
	if scoreTerms([]string{"hello"}, "Hello World", "", "") == 0 {
		t.Error("expected title match to score")
	}
	if scoreTerms([]string{"hello"}, "", "hello", "") == 0 {
		t.Error("expected path match to score")
	}
	if scoreTerms([]string{"hello"}, "", "", "say hello") == 0 {
		t.Error("expected content match to score")
	}
	if scoreTerms([]string{"hello", "missing"}, "", "", "say hello") != 0 {
		t.Error("expected score 0 when not all terms match")
	}
}

func TestStripMarkdown_keepsCodeBlock(t *testing.T) {
	input := "Some text.\n\n```js\nconsole.log('hello')\n```\n\nMore text."
	got := stripMarkdown(input)
	want := "Some text. console.log('hello') More text."
	if got != want {
		t.Errorf("stripMarkdown = %q, want %q", got, want)
	}
}

func TestProvider_SearchPages_codeBlock(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\n```js\nconsole.log('hello')\n```\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	results, err := p.SearchPages(context.Background(), "doc", "console")
	must(t, err)

	if len(results) == 0 {
		t.Fatal("expected search results for term inside code block")
	}

	if results[0].Title != "Home" {
		t.Errorf("expected Home page, got %q", results[0].Title)
	}
}
