package content

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"vordoc/internal/domain"
)

func TestProvider_scanDocPages_detects_directory_index(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome page\n")

	// Group with an index and child pages.
	guideDir := filepath.Join(docRoot, "guide")
	must(t, os.MkdirAll(guideDir, 0o755))
	mustWrite(t, filepath.Join(guideDir, "index.md"), "---\ntitle: Guide\n---\nGuide index\n")
	mustWrite(t, filepath.Join(guideDir, "intro.md"), "---\ntitle: Introduction\n---\nIntro\n")

	// Group with only an index and no child pages.
	emptyDir := filepath.Join(docRoot, "empty")
	must(t, os.MkdirAll(emptyDir, 0o755))
	mustWrite(t, filepath.Join(emptyDir, "index.md"), "---\ntitle: Empty\n---\nEmpty index\n")

	// Group with children but no index.
	noIndexDir := filepath.Join(docRoot, "noindex")
	must(t, os.MkdirAll(noIndexDir, 0o755))
	mustWrite(t, filepath.Join(noIndexDir, "page.md"), "---\ntitle: Page\n---\nPage\n")

	// Group with no children and no index should be skipped.
	skipDir := filepath.Join(docRoot, "skip")
	must(t, os.MkdirAll(skipDir, 0o755))

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	nodes, err := p.scanDocPages(docRoot)
	must(t, err)

	guide := findNode(nodes, "guide")
	if guide == nil {
		t.Fatal("guide node not found")
	}
	if !guide.HasIndex {
		t.Errorf("guide.HasIndex = false, want true")
	}
	if guide.Title != "Guide" {
		t.Errorf("guide.Title = %q, want %q", guide.Title, "Guide")
	}
	if len(guide.Children) != 1 {
		t.Errorf("guide.Children len = %d, want 1", len(guide.Children))
	}

	empty := findNode(nodes, "empty")
	if empty == nil {
		t.Fatal("empty node not found")
	}
	if !empty.HasIndex {
		t.Errorf("empty.HasIndex = false, want true")
	}
	if len(empty.Children) != 0 {
		t.Errorf("empty.Children len = %d, want 0", len(empty.Children))
	}

	noIndex := findNode(nodes, "noindex")
	if noIndex == nil {
		t.Fatal("noindex node not found")
	}
	if noIndex.HasIndex {
		t.Errorf("noindex.HasIndex = true, want false")
	}
	if len(noIndex.Children) != 1 {
		t.Errorf("noindex.Children len = %d, want 1", len(noIndex.Children))
	}

	if findNode(nodes, "skip") != nil {
		t.Error("skip node should have been omitted")
	}
}

func findNode(nodes []domain.PageNode, path string) *domain.PageNode {
	for i := range nodes {
		if nodes[i].Path == path {
			return &nodes[i]
		}
	}
	return nil
}

func TestProvider_GetDoc_embeds_index_page(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome content\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	doc, err := p.GetDoc(context.Background(), "doc")
	must(t, err)

	if doc.IndexPage == nil {
		t.Fatal("doc.IndexPage is nil")
	}
	if doc.IndexPage.Title != "Home" {
		t.Errorf("IndexPage.Title = %q, want %q", doc.IndexPage.Title, "Home")
	}
	if doc.IndexPage.Content != "Home content" {
		t.Errorf("IndexPage.Content = %q, want %q", doc.IndexPage.Content, "Home content")
	}
}

func TestProvider_GetPage_sets_filePath(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome\n")

	guideDir := filepath.Join(docRoot, "guide")
	must(t, os.MkdirAll(guideDir, 0o755))
	mustWrite(t, filepath.Join(guideDir, "index.md"), "---\ntitle: Guide\n---\nGuide\n")
	mustWrite(t, filepath.Join(guideDir, "intro.md"), "---\ntitle: Intro\n---\nIntro\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	cases := []struct {
		pagePath string
		want     string
	}{
		{pagePath: "", want: "index.md"},
		{pagePath: "guide", want: "guide/index.md"},
		{pagePath: "guide/intro", want: "guide/intro.md"},
	}

	for _, c := range cases {
		page, err := p.GetPage(context.Background(), "doc", c.pagePath)
		must(t, err)
		if page.FilePath != c.want {
			t.Errorf("FilePath for %q = %q, want %q", c.pagePath, page.FilePath, c.want)
		}
	}
}

func TestProvider_GetDoc_description_from_index(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\ndescription: From index\n---\nHome content\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	doc, err := p.GetDoc(context.Background(), "doc")
	must(t, err)

	if doc.Description != "From index" {
		t.Errorf("doc.Description = %q, want %q", doc.Description, "From index")
	}
	if doc.IndexPage == nil || doc.IndexPage.Description != "From index" {
		t.Errorf("index page description missing or incorrect")
	}
}

func TestProvider_scanDocPages_sorts_by_order(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome\n")
	mustWrite(t, filepath.Join(docRoot, "z-last.md"), "---\ntitle: Z Last\norder: 1\n---\nZ\n")
	mustWrite(t, filepath.Join(docRoot, "a-first.md"), "---\ntitle: A First\norder: 0\n---\nA\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	nodes, err := p.scanDocPages(docRoot)
	must(t, err)

	if len(nodes) != 2 {
		t.Fatalf("nodes len = %d, want 2", len(nodes))
	}
	if nodes[0].Path != "a-first" {
		t.Errorf("nodes[0].Path = %q, want a-first", nodes[0].Path)
	}
	if nodes[1].Path != "z-last" {
		t.Errorf("nodes[1].Path = %q, want z-last", nodes[1].Path)
	}
}

func TestProvider_scanDocPages_hides_show_false(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\n---\nHome\n")
	mustWrite(t, filepath.Join(docRoot, "visible.md"), "---\ntitle: Visible\n---\nVisible\n")
	mustWrite(t, filepath.Join(docRoot, "hidden.md"), "---\ntitle: Hidden\nshow: false\n---\nHidden\n")

	hiddenDir := filepath.Join(docRoot, "hidden-dir")
	must(t, os.MkdirAll(hiddenDir, 0o755))
	mustWrite(t, filepath.Join(hiddenDir, "index.md"), "---\ntitle: Hidden Dir\nshow: false\n---\nHidden dir\n")

	visibleDir := filepath.Join(docRoot, "visible-dir")
	must(t, os.MkdirAll(visibleDir, 0o755))
	mustWrite(t, filepath.Join(visibleDir, "index.md"), "---\ntitle: Visible Dir\n---\nVisible dir\n")
	mustWrite(t, filepath.Join(visibleDir, "child.md"), "---\ntitle: Child\nshow: false\n---\nChild\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))
	nodes, err := p.scanDocPages(docRoot)
	must(t, err)

	if findNode(nodes, "hidden") != nil {
		t.Error("hidden page should be omitted")
	}
	if findNode(nodes, "hidden-dir") != nil {
		t.Error("hidden directory should be omitted")
	}

	visibleDirNode := findNode(nodes, "visible-dir")
	if visibleDirNode == nil {
		t.Fatal("visible-dir node not found")
	}
	if findNode(visibleDirNode.Children, "child") != nil {
		t.Error("hidden child page should be omitted")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}
