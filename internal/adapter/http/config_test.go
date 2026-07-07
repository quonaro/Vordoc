package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/adapter/content"
)

func TestConfigHandler_GetText(t *testing.T) {
	root := t.TempDir()
	textPath := filepath.Join(root, "text.json")
	if err := os.WriteFile(textPath, []byte(`{"app":{"title":"Test Docs"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	handler := NewConfigHandler(provider, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/text", handler.GetText)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/text", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	app, ok := body["app"].(map[string]any)
	if !ok {
		t.Fatalf("expected app object, got %T", body["app"])
	}
	if app["title"] != "Test Docs" {
		t.Errorf("expected title 'Test Docs', got %v", app["title"])
	}
	if app["logoAlt"] != "logo" {
		t.Errorf("expected logoAlt default to be preserved, got %v", app["logoAlt"])
	}
}

func TestConfigHandler_GetText_mergeDefaults(t *testing.T) {
	root := t.TempDir()
	textPath := filepath.Join(root, "text.json")
	if err := os.WriteFile(textPath, []byte(`{"search":{"placeholder":"Find..."}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	handler := NewConfigHandler(provider, slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/text", handler.GetText)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/text", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	search, ok := body["search"].(map[string]any)
	if !ok {
		t.Fatalf("expected search object, got %T", body["search"])
	}
	if search["placeholder"] != "Find..." {
		t.Errorf("expected placeholder override, got %v", search["placeholder"])
	}
	if search["searching"] != "Searching..." {
		t.Errorf("expected searching default to be preserved, got %v", search["searching"])
	}
}
