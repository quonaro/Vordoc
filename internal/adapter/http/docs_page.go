package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/domain"
)

// GetDocOrPage handles both doc metadata and page requests.
func (h *DocsHandler) GetDocOrPage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "invalid_path")
		return
	}

	docName := parts[0]

	if len(parts) == 1 {
		h.serveDoc(w, r, docName)
		return
	}

	h.servePage(w, r, docName, strings.Join(parts[1:], "/"))
}

func (h *DocsHandler) serveDoc(w http.ResponseWriter, r *http.Request, docName string) {
	summary, err := h.contentProvider.GetDocSummary(r.Context(), docName)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrDocNotFound.Error()) {
			writeError(w, http.StatusNotFound, "doc_not_found")
			return
		}
		h.logger.Error("failed to get doc summary", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_get_doc")
		return
	}

	if summary.Access == "password" && !h.hasValidCookie(r, docName, summary.Scope) {
		writePasswordRequired(w, summary.Scope)
		return
	}

	doc, err := h.contentProvider.GetDoc(r.Context(), docName)
	if err != nil {
		h.logger.Error("failed to get doc", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_get_doc")
		return
	}

	if doc.IndexPage != nil {
		if doc.IndexPage.Access == "password" && !h.hasValidCookie(r, docName, doc.IndexPage.AccessScope) {
			doc.IndexPage = nil
		} else {
			doc.IndexPage.Access = ""
		}
	}

	writeJSON(w, http.StatusOK, doc)
}

func (h *DocsHandler) servePage(w http.ResponseWriter, r *http.Request, docName, pagePath string) {
	page, err := h.contentProvider.GetPage(r.Context(), docName, pagePath)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrPageNotFound.Error()) {
			writeError(w, http.StatusNotFound, "page_not_found")
			return
		}
		h.logger.Error("failed to get page", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_get_page")
		return
	}

	if page.Access == "password" && !h.hasValidCookie(r, docName, page.AccessScope) {
		writePasswordRequired(w, page.AccessScope)
		return
	}

	page.Access = ""
	writeJSON(w, http.StatusOK, page)
}
