package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/domain"
	"vordoc/internal/service"
)

// DocsHandler handles documentation read endpoints.
type DocsHandler struct {
	contentProvider domain.ContentProvider
	passwordService *service.PasswordService
	cookieSecret    []byte
	logger          *slog.Logger
}

// NewDocsHandler constructs a docs handler.
func NewDocsHandler(contentProvider domain.ContentProvider, passwordService *service.PasswordService, cookieSecret string, logger *slog.Logger) *DocsHandler {
	return &DocsHandler{
		contentProvider: contentProvider,
		passwordService: passwordService,
		cookieSecret:    []byte(cookieSecret),
		logger:          logger,
	}
}

// docSummary is a lightweight list entry for a documentation.
type docSummary struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Access      string `json:"access,omitempty"`
}

// searchResult is a single search hit exposed to the frontend.
type searchResult struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	Snippet string `json:"snippet,omitempty"`
}

// ServeLogo serves the logo image for the root site or a documentation.
func (h *DocsHandler) ServeLogo(w http.ResponseWriter, r *http.Request) {
	doc := strings.TrimSpace(r.URL.Query().Get("doc"))

	path, err := h.contentProvider.GetLogoPath(r.Context(), doc)
	if err != nil {
		h.logger.Error("failed to resolve logo", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to resolve logo")
		return
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "logo not found")
			return
		}
		h.logger.Error("failed to stat logo", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to serve logo")
		return
	}

	ext := strings.ToLower(filepath.Ext(path))
	contentType := "application/octet-stream"
	if ext == ".svg" {
		contentType = "image/svg+xml"
	} else if mt := mime.TypeByExtension(ext); mt != "" {
		contentType = mt
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeFile(w, r, path)
}

// Search searches within the current documentation for the given query.
func (h *DocsHandler) Search(w http.ResponseWriter, r *http.Request) {
	docName := strings.TrimSpace(chi.URLParam(r, "doc"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	if docName == "" {
		writeError(w, http.StatusBadRequest, "doc name is required")
		return
	}
	if query == "" {
		writeJSON(w, http.StatusOK, map[string]any{"results": []searchResult{}})
		return
	}

	results, err := h.contentProvider.SearchPages(r.Context(), docName, query)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrDocNotFound.Error()) {
			writeError(w, http.StatusNotFound, "documentation not found")
			return
		}
		h.logger.Error("failed to search pages", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to search")
		return
	}

	out := make([]searchResult, 0, len(results))
	for _, res := range results {
		if res.Access == "password" {
			if !h.hasValidCookie(r, docName, res.Path) {
				// Hide snippet from protected pages without access.
				res.Snippet = ""
			}
		}
		out = append(out, searchResult{
			Title:   res.Title,
			Path:    res.Path,
			Snippet: res.Snippet,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"results": out})
}

// ListDocs returns all available documentations.
func (h *DocsHandler) ListDocs(w http.ResponseWriter, r *http.Request) {
	names, err := h.contentProvider.ListDocs(r.Context())
	if err != nil {
		h.logger.Error("failed to list docs", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to list docs")
		return
	}

	docs := make([]docSummary, 0, len(names))
	for _, name := range names {
		summary := docSummary{Name: name, Title: name, Access: "public"}
		if doc, err := h.contentProvider.GetDoc(r.Context(), name); err == nil {
			summary.Title = doc.Title
			summary.Description = doc.Description
			summary.Access = doc.Access
		}
		docs = append(docs, summary)
	}

	writeJSON(w, http.StatusOK, map[string]any{"docs": docs})
}

// GetDocOrPage handles both doc metadata and page requests.
func (h *DocsHandler) GetDocOrPage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	docName := parts[0]

	if len(parts) == 1 {
		doc, err := h.contentProvider.GetDoc(r.Context(), docName)
		if err != nil {
			if strings.Contains(err.Error(), domain.ErrDocNotFound.Error()) {
				writeError(w, http.StatusNotFound, "documentation not found")
				return
			}
			h.logger.Error("failed to get doc", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "failed to get doc")
			return
		}

		// Don't expose password-protected index content without a valid cookie.
		if doc.IndexPage != nil {
			if doc.IndexPage.Access == "password" && !h.hasValidCookie(r, docName, doc.IndexPage.Path) {
				doc.IndexPage = nil
			} else {
				doc.IndexPage.Access = ""
			}
		}

		writeJSON(w, http.StatusOK, doc)
		return
	}

	pagePath := strings.Join(parts[1:], "/")

	page, err := h.contentProvider.GetPage(r.Context(), docName, pagePath)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrPageNotFound.Error()) {
			writeError(w, http.StatusNotFound, "page not found")
			return
		}
		h.logger.Error("failed to get page", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to get page")
		return
	}

	if page.Access == "password" {
		if !h.hasValidCookie(r, docName, pagePath) {
			writeJSON(w, http.StatusForbidden, map[string]any{
				"error":             "password required",
				"password_required": true,
			})
			return
		}
	}

	page.Access = ""
	writeJSON(w, http.StatusOK, page)
}

// verifyInput holds the password verification request.
type verifyInput struct {
	Password string `json:"password"`
}

// verifyOutput holds the password verification response.
type verifyOutput struct {
	Success bool `json:"success"`
}

// VerifyPassword checks a page password and sets an access cookie on success.
func (h *DocsHandler) VerifyPassword(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	docName := parts[0]
	pagePath := strings.Join(parts[1:], "/")

	var input verifyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get the page to check its access level
	page, err := h.contentProvider.GetPage(r.Context(), docName, pagePath)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrPageNotFound.Error()) {
			writeError(w, http.StatusNotFound, "page not found")
			return
		}
		h.logger.Error("failed to get page for verify", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to verify")
		return
	}

	if page.Access != "password" {
		writeJSON(w, http.StatusOK, verifyOutput{Success: true})
		return
	}

	if page.PasswordHash == "" {
		writeError(w, http.StatusInternalServerError, "password hash not configured")
		return
	}

	if !h.passwordService.Verify(input.Password, page.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid password")
		return
	}

	h.setAccessCookie(w, docName, pagePath)
	writeJSON(w, http.StatusOK, verifyOutput{Success: true})
}

// cookieValue represents the signed cookie payload.
type cookieValue struct {
	Doc  string `json:"doc"`
	Page string `json:"page"`
	Exp  int64  `json:"exp"`
}

func (h *DocsHandler) hasValidCookie(r *http.Request, doc, page string) bool {
	cookie, err := r.Cookie("vordoc_access")
	if err != nil {
		return false
	}

	// Cookie value: base64(sig + "." + json)
	val, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return false
	}

	parts := strings.SplitN(string(val), ".", 2)
	if len(parts) != 2 {
		return false
	}

	sigBytes, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	expectedSig := h.sign(parts[1])
	if !hmac.Equal(sigBytes, expectedSig) {
		return false
	}

	var cv cookieValue
	if err := json.Unmarshal([]byte(parts[1]), &cv); err != nil {
		return false
	}

	if cv.Exp < time.Now().Unix() {
		return false
	}

	return cv.Doc == doc && cv.Page == page
}

func (h *DocsHandler) setAccessCookie(w http.ResponseWriter, doc, page string) {
	cv := cookieValue{
		Doc:  doc,
		Page: page,
		Exp:  time.Now().Add(24 * time.Hour).Unix(),
	}

	data, _ := json.Marshal(cv)
	sig := base64.URLEncoding.EncodeToString(h.sign(string(data)))
	raw := fmt.Sprintf("%s.%s", sig, string(data))
	value := base64.URLEncoding.EncodeToString([]byte(raw))

	http.SetCookie(w, &http.Cookie{
		Name:     "vordoc_access",
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
}

func (h *DocsHandler) sign(data string) []byte {
	mac := hmac.New(sha256.New, h.cookieSecret)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}
