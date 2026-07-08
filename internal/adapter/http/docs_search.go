package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/domain"
)

// globalSearchResult is the public representation of a doc group.
type globalSearchResult struct {
	Name        string             `json:"name"`
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	Access      string             `json:"access,omitempty"`
	Pages       []globalSearchPage `json:"pages,omitempty"`
}

type globalSearchPage struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	Snippet string `json:"snippet,omitempty"`
}

// GlobalSearch searches across all documentations.
func (h *DocsHandler) GlobalSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	if query == "" {
		writeJSON(w, http.StatusOK, map[string]any{"results": []globalSearchResult{}})
		return
	}

	results, err := h.contentProvider.SearchAllDocs(r.Context(), query)
	if err != nil {
		h.logger.Error("failed to search all docs", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_search")
		return
	}

	out := make([]globalSearchResult, 0, len(results))
	for _, doc := range results {
		group := globalSearchResult{
			Name:        doc.Name,
			Title:       doc.Title,
			Description: doc.Description,
			Access:      doc.Access,
		}

		isProtected := doc.Access == "password"
		if !isProtected {
			group.Pages = make([]globalSearchPage, 0, len(doc.Pages))
			for _, page := range doc.Pages {
				group.Pages = append(group.Pages, globalSearchPage{
					Title:   page.Title,
					Path:    page.Path,
					Snippet: page.Snippet,
				})
			}
		}

		out = append(out, group)
	}

	writeJSON(w, http.StatusOK, map[string]any{"results": out})
}

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
