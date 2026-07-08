package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/adapter/content"
	"vordoc/internal/service"
)

func TestDocsHandler_Search(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	if err := os.MkdirAll(docRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docRoot, "config.yaml"), []byte("title: Test Doc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docRoot, "index.md"), []byte("---\ntitle: Home\n---\nHome page\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docRoot, "guide.md"), []byte("---\ntitle: Guide\n---\n# Searchable guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	passwordService := service.NewPasswordService()
	handler := NewDocsHandler(provider, passwordService, "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/{doc}/search", handler.Search)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/doc/search?q=searchable", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Results []struct {
			Title   string `json:"title"`
			Path    string `json:"path"`
			Snippet string `json:"snippet"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(body.Results))
	}
	if body.Results[0].Title != "Guide" {
		t.Errorf("expected title Guide, got %q", body.Results[0].Title)
	}
	if body.Results[0].Path != "guide" {
		t.Errorf("expected path guide, got %q", body.Results[0].Path)
	}
	if body.Results[0].Snippet == "" {
		t.Errorf("expected non-empty snippet")
	}
}

func TestDocsHandler_Search_docNotFound(t *testing.T) {
	root := t.TempDir()
	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	passwordService := service.NewPasswordService()
	handler := NewDocsHandler(provider, passwordService, "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/{doc}/search", handler.Search)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missing/search?q=test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestDocsHandler_Search_missingDocParam(t *testing.T) {
	root := t.TempDir()
	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	passwordService := service.NewPasswordService()
	handler := NewDocsHandler(provider, passwordService, "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/{doc}/search", handler.Search)

	req := httptest.NewRequest(http.MethodGet, "/api/v1//search?q=test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDocsHandler_Search_emptyQuery(t *testing.T) {
	root := t.TempDir()
	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	passwordService := service.NewPasswordService()
	handler := NewDocsHandler(provider, passwordService, "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/{doc}/search", handler.Search)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/doc/search?q=", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		Results []any `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(body.Results))
	}
}

func TestDocsHandler_ServeAsset(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	if err := os.MkdirAll(docRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docRoot, "config.yaml"), []byte("title: Test Doc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	imagesDir := filepath.Join(docRoot, "images")
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "logo.svg"), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	passwordService := service.NewPasswordService()
	handler := NewDocsHandler(provider, passwordService, "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/assets/{doc}/*", handler.ServeAsset)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/doc/images/logo.svg", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "image/svg+xml" {
		t.Errorf("expected Content-Type image/svg+xml, got %q", got)
	}

	wantDisposition := `inline; filename="logo.svg"`
	if got := rec.Header().Get("Content-Disposition"); got != wantDisposition {
		t.Errorf("expected Content-Disposition %q, got %q", wantDisposition, got)
	}

	if body := rec.Body.String(); !strings.Contains(body, "<title>logo.svg</title>") {
		t.Errorf("expected body to contain injected title, got %q", body)
	}
}

func TestDocsHandler_ServeAsset_SVG_KeepsExistingTitle(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	if err := os.MkdirAll(docRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docRoot, "config.yaml"), []byte("title: Test Doc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	imagesDir := filepath.Join(docRoot, "images")
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imagesDir, "logo.svg"), []byte("<svg><title>Custom title</title></svg>"), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	passwordService := service.NewPasswordService()
	handler := NewDocsHandler(provider, passwordService, "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/assets/{doc}/*", handler.ServeAsset)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/doc/images/logo.svg", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "<title>Custom title</title>") {
		t.Errorf("expected existing title to be preserved, got %q", body)
	}
	if strings.Contains(body, "<title>logo.svg</title>") {
		t.Errorf("expected existing title not to be overwritten")
	}
}

func TestDocsHandler_GetDocOrPage_rejectsPathTraversal(t *testing.T) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "doc")
	otherRoot := filepath.Join(root, "other")
	if err := os.MkdirAll(docRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(otherRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docRoot, "config.yaml"), []byte("title: Test Doc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docRoot, "index.md"), []byte("---\ntitle: Home\n---\nHome\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(otherRoot, "secret.md"), []byte("---\ntitle: Secret\n---\nSecret content\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	passwordService := service.NewPasswordService()
	handler := NewDocsHandler(provider, passwordService, "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/*", handler.GetDocOrPage)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/doc/../other/secret", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for path traversal, got %d: %s", rec.Code, rec.Body.String())
	}
}
