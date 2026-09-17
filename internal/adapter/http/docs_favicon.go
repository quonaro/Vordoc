package http

import (
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ServeFavicon serves the site favicon from the content root.
func (h *DocsHandler) ServeFavicon(w http.ResponseWriter, r *http.Request) {
	faviconPath, err := h.contentProvider.GetFaviconPath(r.Context())
	if err != nil {
		h.logger.Error("failed to resolve favicon", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_resolve_favicon")
		return
	}

	if _, err := os.Stat(faviconPath); err != nil { // #nosec G703 — path is validated by contentProvider
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "favicon_not_found")
			return
		}
		h.logger.Error("failed to stat favicon", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_serve_favicon")
		return
	}

	ext := strings.ToLower(filepath.Ext(faviconPath))
	contentType := "application/octet-stream"
	if mt := mime.TypeByExtension(ext); mt != "" {
		contentType = mt
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeFile(w, r, faviconPath) // #nosec G703 — path is validated by contentProvider
}
