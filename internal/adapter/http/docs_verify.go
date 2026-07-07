package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/domain"
)

// verifyInput holds the password verification request.
type verifyInput struct {
	Password string `json:"password"`
}

// verifyOutput holds the password verification response.
type verifyOutput struct {
	Success bool   `json:"success"`
	Scope   string `json:"scope,omitempty"`
}

// VerifyPassword checks a page password and sets an access cookie on success.
func (h *DocsHandler) VerifyPassword(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "invalid_path")
		return
	}

	docName := parts[0]
	pagePath := ""
	if len(parts) > 1 {
		pagePath = strings.Join(parts[1:], "/")
	}

	var input verifyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if pagePath == "" {
		h.verifyDoc(w, r, docName, input.Password)
		return
	}

	h.verifyPage(w, r, docName, pagePath, input.Password)
}

func (h *DocsHandler) verifyDoc(w http.ResponseWriter, r *http.Request, docName, password string) {
	summary, err := h.contentProvider.GetDocSummary(r.Context(), docName)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrDocNotFound.Error()) {
			writeError(w, http.StatusNotFound, "doc_not_found")
			return
		}
		h.logger.Error("failed to get doc for verify", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_verify")
		return
	}

	if summary.Access != "password" {
		writeJSON(w, http.StatusOK, verifyOutput{Success: true})
		return
	}

	if summary.PasswordHash == "" {
		writeError(w, http.StatusInternalServerError, "password_hash_not_configured")
		return
	}

	if !h.passwordService.Verify(password, summary.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid_password")
		return
	}

	h.setAccessCookie(w, r, docName, summary.Scope)
	writeJSON(w, http.StatusOK, verifyOutput{Success: true, Scope: summary.Scope})
}

func (h *DocsHandler) verifyPage(w http.ResponseWriter, r *http.Request, docName, pagePath, password string) {
	page, err := h.contentProvider.GetPage(r.Context(), docName, pagePath)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrPageNotFound.Error()) {
			writeError(w, http.StatusNotFound, "page_not_found")
			return
		}
		h.logger.Error("failed to get page for verify", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_verify")
		return
	}

	if page.Access != "password" {
		writeJSON(w, http.StatusOK, verifyOutput{Success: true})
		return
	}

	if page.PasswordHash == "" {
		writeError(w, http.StatusInternalServerError, "password_hash_not_configured")
		return
	}

	if !h.passwordService.Verify(password, page.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid_password")
		return
	}

	h.setAccessCookie(w, r, docName, page.AccessScope)
	writeJSON(w, http.StatusOK, verifyOutput{Success: true, Scope: page.AccessScope})
}
