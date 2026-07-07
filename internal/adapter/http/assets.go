package http

import (
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/domain"
)

// ServeAsset serves a static file from a documentation directory.
func (h *DocsHandler) ServeAsset(w http.ResponseWriter, r *http.Request) {
	docName := strings.TrimSpace(chi.URLParam(r, "doc"))
	assetPath := strings.TrimPrefix(chi.URLParam(r, "*"), "/")

	if docName == "" {
		writeError(w, http.StatusBadRequest, "doc_name_required")
		return
	}
	if assetPath == "" {
		writeError(w, http.StatusNotFound, "asset_not_found")
		return
	}

	path, err := h.contentProvider.GetAssetPath(r.Context(), docName, assetPath)
	if err != nil {
		if errors.Is(err, domain.ErrDocNotFound) {
			writeError(w, http.StatusNotFound, "doc_not_found")
			return
		}
		if errors.Is(err, domain.ErrAssetNotFound) {
			writeError(w, http.StatusNotFound, "asset_not_found")
			return
		}
		h.logger.Error("failed to resolve asset",
			slog.String("error", err.Error()),
			slog.String("doc", docName),
			slog.String("asset", assetPath),
		)
		writeError(w, http.StatusInternalServerError, "failed_to_resolve_asset")
		return
	}

	ext := strings.ToLower(filepath.Ext(path))
	contentType := "application/octet-stream"
	if mt := mime.TypeByExtension(ext); mt != "" {
		contentType = mt
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeFile(w, r, path)
}
