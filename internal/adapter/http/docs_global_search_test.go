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
	"vordoc/internal/service"
)

func TestDocsHandler_GlobalSearch(t *testing.T) {
	root := t.TempDir()

	docA := filepath.Join(root, "doc-a")
	if err := os.MkdirAll(docA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docA, "config.yaml"), []byte("title: Alpha\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docA, "index.md"), []byte("---\ntitle: Alpha Home\n---\nAlpha page\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	docB := filepath.Join(root, "doc-b")
	if err := os.MkdirAll(docB, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docB, "config.yaml"), []byte("title: Beta\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docB, "index.md"), []byte("---\ntitle: Beta Home\n---\nBeta page content\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	passwordService := service.NewPasswordService()
	handler := NewDocsHandler(provider, passwordService, "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Get("/api/v1/search", handler.GlobalSearch)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=beta", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Results []struct {
			Name   string `json:"name"`
			Title  string `json:"title"`
			Access string `json:"access"`
			Pages  []struct {
				Title   string `json:"title"`
				Path    string `json:"path"`
				Snippet string `json:"snippet"`
			} `json:"pages"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Results) != 1 {
		t.Fatalf("expected 1 doc result, got %d", len(body.Results))
	}

	var betaFound bool
	for _, doc := range body.Results {
		if doc.Name == "doc-b" {
			betaFound = true
			if doc.Title != "Beta" {
				t.Errorf("expected title Beta, got %q", doc.Title)
			}
			if len(doc.Pages) == 0 {
				t.Error("expected beta pages in result")
			}
		}
	}
	if !betaFound {
		t.Error("expected doc-b in results")
	}
}
