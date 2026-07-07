package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/domain"
)

// Search searches within the current documentation for the given query.
func (h *DocsHandler) Search(w http.ResponseWriter, r *http.Request) {
	docName := strings.TrimSpace(chi.URLParam(r, "doc"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	if docName == "" {
		writeError(w, http.StatusBadRequest, "doc_name_required")
		return
	}
	if query == "" {
		writeJSON(w, http.StatusOK, map[string]any{"results": []searchResult{}})
		return
	}

	results, err := h.contentProvider.SearchPages(r.Context(), docName, query)
	if err != nil {
		if strings.Contains(err.Error(), domain.ErrDocNotFound.Error()) {
			writeError(w, http.StatusNotFound, "doc_not_found")
			return
		}
		h.logger.Error("failed to search pages", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_search")
		return
	}

	out := make([]searchResult, 0, len(results))
	for _, res := range results {
		if res.Access == "password" && !h.hasValidCookie(r, docName, res.AccessScope) {
			res.Snippet = ""
		}
		out = append(out, searchResult{
			Title:   res.Title,
			Path:    res.Path,
			Snippet: res.Snippet,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"results": out})
}
