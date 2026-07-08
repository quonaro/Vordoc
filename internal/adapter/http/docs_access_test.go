package http

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"vordoc/internal/adapter/content"
	"vordoc/internal/service"
)

func setupProtectedDoc(t *testing.T, handler **DocsHandler) (*httptest.ResponseRecorder, *chi.Mux) {
	root := t.TempDir()

	docRoot := filepath.Join(root, "admin")
	must(t, os.MkdirAll(docRoot, 0o755))
	mustWrite(t, filepath.Join(docRoot, "config.yaml"), "title: Admin Docs\n")
	mustWrite(t, filepath.Join(docRoot, "access.yaml"), "access: password\npassword_hash: "+hash(t, "secret")+"\n")
	mustWrite(t, filepath.Join(docRoot, "index.md"), "---\ntitle: Admin Home\n---\nAdmin home content\n")
	mustWrite(t, filepath.Join(docRoot, "settings.md"), "---\ntitle: Settings\n---\nSettings content\n")

	publicDir := filepath.Join(docRoot, "public")
	must(t, os.MkdirAll(publicDir, 0o755))
	mustWrite(t, filepath.Join(publicDir, "access.yaml"), "access: none\n")
	mustWrite(t, filepath.Join(publicDir, "info.md"), "---\ntitle: Info\n---\nPublic info\n")

	provider := content.NewProvider(root, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	passwordService := service.NewPasswordService()
	*handler = NewDocsHandler(provider, passwordService, "secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/docs", (*handler).ListDocs)
		r.Get("/{doc}/search", (*handler).Search)
		r.Get("/assets/{doc}/*", (*handler).ServeAsset)
		r.Get("/*", (*handler).GetDocOrPage)
		r.Post("/*", (*handler).VerifyPassword)
	})

	return httptest.NewRecorder(), r
}

func TestDocsHandler_DocLevelProtection(t *testing.T) {
	var handler *DocsHandler
	_, r := setupProtectedDoc(t, &handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["password_required"] != true {
		t.Errorf("expected password_required true")
	}
	if body["scope"] != "" {
		t.Errorf("expected root scope, got %q", body["scope"])
	}
}

func TestDocsHandler_DocVerifyUnlocksAll(t *testing.T) {
	var handler *DocsHandler
	_, r := setupProtectedDoc(t, &handler)

	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin", bytes.NewReader([]byte(`{"password":"secret"}`)))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	r.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected 200 verify, got %d: %s", verifyRec.Code, verifyRec.Body.String())
	}

	cookies := verifyRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected cookie after verify")
	}

	metaReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin", nil)
	for _, c := range cookies {
		metaReq.AddCookie(c)
	}
	metaRec := httptest.NewRecorder()
	r.ServeHTTP(metaRec, metaReq)
	if metaRec.Code != http.StatusOK {
		t.Fatalf("expected 200 metadata, got %d: %s", metaRec.Code, metaRec.Body.String())
	}

	pageReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	for _, c := range cookies {
		pageReq.AddCookie(c)
	}
	pageRec := httptest.NewRecorder()
	r.ServeHTTP(pageRec, pageReq)
	if pageRec.Code != http.StatusOK {
		t.Fatalf("expected 200 page, got %d: %s", pageRec.Code, pageRec.Body.String())
	}
}

func TestDocsHandler_PublicOverride(t *testing.T) {
	var handler *DocsHandler
	_, r := setupProtectedDoc(t, &handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/public/info", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 public override, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDocsHandler_ListDocs_HidesProtectedDescription(t *testing.T) {
	var handler *DocsHandler
	_, r := setupProtectedDoc(t, &handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 list, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Docs []docSummary `json:"docs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.Docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(body.Docs))
	}
	if body.Docs[0].Access != "password" {
		t.Errorf("expected access password, got %q", body.Docs[0].Access)
	}
	if body.Docs[0].Description != "" {
		t.Errorf("expected empty description for protected doc, got %q", body.Docs[0].Description)
	}
	if body.Docs[0].Title != "Admin Docs" {
		t.Errorf("expected title Admin Docs, got %q", body.Docs[0].Title)
	}
}

func TestDocsHandler_Search_HidesSnippet(t *testing.T) {
	var handler *DocsHandler
	_, r := setupProtectedDoc(t, &handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/search?q=content", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 search, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Results []searchResult `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	for _, res := range body.Results {
		if res.Snippet != "" {
			t.Errorf("expected empty snippet for %q, got %q", res.Path, res.Snippet)
		}
	}
}

func TestDocsHandler_ServeAsset_RequiresPassword(t *testing.T) {
	var handler *DocsHandler
	_, r := setupProtectedDoc(t, &handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/admin/settings.md", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 protected asset, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDocsHandler_ServeAsset_ProtectedCacheHeaders(t *testing.T) {
	var handler *DocsHandler
	_, r := setupProtectedDoc(t, &handler)

	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin", bytes.NewReader([]byte(`{"password":"secret"}`)))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	r.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected 200 verify, got %d: %s", verifyRec.Code, verifyRec.Body.String())
	}

	cookies := verifyRec.Result().Cookies()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/admin/settings.md", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 asset, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Errorf("expected Cache-Control private, no-store, got %q", got)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hash(t *testing.T, pwd string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(h)
}
