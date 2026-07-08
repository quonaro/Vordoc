package content

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestProvider_GetDocSummary_docAccess(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Home\naccess: password\npassword_hash: "+hash(t, "secret")+"\n---\nHome\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	summary, err := p.GetDocSummary(context.Background(), "doc")
	must(t, err)

	if summary.Access != "password" {
		t.Errorf("expected access password, got %q", summary.Access)
	}
	if summary.PasswordHash == "" {
		t.Errorf("expected password hash")
	}
	if summary.Scope != "" {
		t.Errorf("expected root scope, got %q", summary.Scope)
	}
	if summary.Title != "Test Doc" {
		t.Errorf("expected title Test Doc, got %q", summary.Title)
	}
}

func TestProvider_GetPage_inheritedScope(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\naccess: password\npassword_hash: "+hash(t, "secret")+"\n")
	mustWrite(t, filepath.Join(docRoot, "guide.md"), "---\ntitle: Guide\n---\nGuide\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	page, err := p.GetPage(context.Background(), "doc", "guide")
	must(t, err)

	if page.Access != "password" {
		t.Errorf("expected access password, got %q", page.Access)
	}
	if page.AccessScope != "" {
		t.Errorf("expected inherited root scope, got %q", page.AccessScope)
	}
	if page.PasswordHash == "" {
		t.Errorf("expected inherited password hash")
	}
}

func TestProvider_GetPage_folderScope(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	adminDir := filepath.Join(docRoot, "admin")
	must(t, os.MkdirAll(adminDir, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\n")
	mustWrite(t, filepath.Join(adminDir, "config.yaml"), "access: password\npassword_hash: "+hash(t, "admin")+"\n")
	mustWrite(t, filepath.Join(adminDir, "settings.md"), "---\ntitle: Settings\n---\nSettings\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	page, err := p.GetPage(context.Background(), "doc", "admin/settings")
	must(t, err)

	if page.Access != "password" {
		t.Errorf("expected access password, got %q", page.Access)
	}
	if page.AccessScope != "admin" {
		t.Errorf("expected scope admin, got %q", page.AccessScope)
	}
}

func TestProvider_GetPage_publicOverride(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	adminDir := filepath.Join(docRoot, "admin")
	publicDir := filepath.Join(adminDir, "public")
	must(t, os.MkdirAll(publicDir, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\n")
	mustWrite(t, filepath.Join(adminDir, "config.yaml"), "access: password\npassword_hash: "+hash(t, "admin")+"\n")
	mustWrite(t, filepath.Join(publicDir, "config.yaml"), "access: none\n")
	mustWrite(t, filepath.Join(publicDir, "info.md"), "---\ntitle: Info\n---\nInfo\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	page, err := p.GetPage(context.Background(), "doc", "admin/public/info")
	must(t, err)

	if page.Access != "public" {
		t.Errorf("expected access public, got %q", page.Access)
	}
	if page.AccessScope != "" {
		t.Errorf("expected empty scope, got %q", page.AccessScope)
	}
}

func TestProvider_GetPage_inheritHash(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	adminDir := filepath.Join(docRoot, "admin")
	must(t, os.MkdirAll(adminDir, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\naccess: password\npassword_hash: "+hash(t, "shared")+"\n")
	mustWrite(t, filepath.Join(adminDir, "config.yaml"), "access: password\n")
	mustWrite(t, filepath.Join(adminDir, "settings.md"), "---\ntitle: Settings\n---\nSettings\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	page, err := p.GetPage(context.Background(), "doc", "admin/settings")
	must(t, err)

	if page.Access != "password" {
		t.Errorf("expected access password, got %q", page.Access)
	}
	if page.AccessScope != "" {
		t.Errorf("expected inherited root scope, got %q", page.AccessScope)
	}
	if ok := bcrypt.CompareHashAndPassword([]byte(page.PasswordHash), []byte("shared")); ok != nil {
		t.Errorf("expected inherited hash to match password shared")
	}
}

func TestProvider_GetAssetAccess(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	adminDir := filepath.Join(docRoot, "admin")
	must(t, os.MkdirAll(adminDir, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\naccess: password\npassword_hash: "+hash(t, "secret")+"\n")
	mustWrite(t, filepath.Join(adminDir, "image.png"), "png")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	info, err := p.GetAssetAccess(context.Background(), "doc", "admin/image.png")
	must(t, err)

	if info.Access != "password" {
		t.Errorf("expected access password, got %q", info.Access)
	}
	if info.Scope != "" {
		t.Errorf("expected root scope, got %q", info.Scope)
	}
}

func TestProvider_GetAssetAccess_rootAssetProtected(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\naccess: password\npassword_hash: "+hash(t, "secret")+"\n")
	mustWrite(t, filepath.Join(docRoot, "logo.png"), "png")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	info, err := p.GetAssetAccess(context.Background(), "doc", "logo.png")
	must(t, err)

	if info.Access != "password" {
		t.Errorf("expected access password, got %q", info.Access)
	}
	if info.Scope != "" {
		t.Errorf("expected root scope, got %q", info.Scope)
	}
}

func TestProvider_GetAssetAccess_publicSubdirAsset(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	publicDir := filepath.Join(docRoot, "public")
	must(t, os.MkdirAll(publicDir, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\naccess: password\npassword_hash: "+hash(t, "secret")+"\n")
	mustWrite(t, filepath.Join(publicDir, "config.yaml"), "access: none\n")
	mustWrite(t, filepath.Join(publicDir, "info.md"), "info\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	info, err := p.GetAssetAccess(context.Background(), "doc", "public/info.md")
	must(t, err)

	if info.Access != "public" {
		t.Errorf("expected access public, got %q", info.Access)
	}
}

func TestProvider_GetPage_publicFrontmatterWithHash(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Test Doc\n")
	mustWrite(t, filepath.Join(docRoot, "guide.md"), "---\ntitle: Guide\naccess: public\npassword_hash: "+hash(t, "secret")+"\n---\nGuide\n")

	p := NewProvider(root, slog.New(slog.NewTextHandler(io.Discard, nil)))

	page, err := p.GetPage(context.Background(), "doc", "guide")
	must(t, err)

	if page.Access != "public" {
		t.Errorf("expected access public, got %q", page.Access)
	}
	if page.PasswordHash != "" {
		t.Errorf("expected empty hash for public page, got %q", page.PasswordHash)
	}
}

func hash(t *testing.T, pwd string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	must(t, err)
	return string(h)
}
