package http

import (
	"log/slog"
	"net/http"
)

// ListDocs returns all available documentations.
func (h *DocsHandler) ListDocs(w http.ResponseWriter, r *http.Request) {
	names, err := h.contentProvider.ListDocs(r.Context())
	if err != nil {
		h.logger.Error("failed to list docs", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_list_docs")
		return
	}

	docs := make([]docSummary, 0, len(names))
	for _, name := range names {
		summary := docSummary{Name: name, Title: name, Access: "public"}
		if s, err := h.contentProvider.GetDocSummary(r.Context(), name); err == nil {
			summary.Title = s.Title
			summary.Access = s.Access
			if s.Access != "password" || h.hasValidCookie(r, name, s.Scope) {
				summary.Description = s.Description
			}
		}
		docs = append(docs, summary)
	}

	writeJSON(w, http.StatusOK, map[string]any{"docs": docs})
}
