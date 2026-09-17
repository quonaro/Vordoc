package http

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/adapter/content"
	"vordoc/internal/service"
)

func TestDocsHandler_ServeFavicon(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte("favicon: \"favicon.ico\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "favicon.ico"), []byte("icon-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	handler := NewDocsHandler(provider, service.NewPasswordService(), "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/favicon", handler.ServeFavicon)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/favicon", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "icon-bytes" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "icon-bytes")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/vnd.microsoft.icon" {
		t.Errorf("Content-Type = %q, want %q", ct, "image/vnd.microsoft.icon")
	}
}

func TestDocsHandler_ServeFavicon_missing(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte("favicon: \"favicon.ico\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	handler := NewDocsHandler(provider, service.NewPasswordService(), "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/favicon", handler.ServeFavicon)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/favicon", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
