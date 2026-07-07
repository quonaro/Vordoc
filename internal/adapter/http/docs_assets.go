package http

import (
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"vordoc/internal/domain"
)

// ServeLogo serves the logo image for the root site or a documentation.
func (h *DocsHandler) ServeLogo(w http.ResponseWriter, r *http.Request) {
	doc := strings.TrimSpace(r.URL.Query().Get("doc"))

	if doc != "" {
		summary, err := h.contentProvider.GetDocSummary(r.Context(), doc)
		if err != nil {
			if strings.Contains(err.Error(), domain.ErrDocNotFound.Error()) {
				writeError(w, http.StatusNotFound, "doc_not_found")
				return
			}
			h.logger.Error("failed to resolve logo", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "failed_to_resolve_logo")
			return
		}
		if summary.Access == "password" && !h.hasValidCookie(r, doc, summary.Scope) {
			writePasswordRequired(w, summary.Scope)
			return
		}
	}

	logoPath, err := h.contentProvider.GetLogoPath(r.Context(), doc)
	if err != nil {
		h.logger.Error("failed to resolve logo", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_resolve_logo")
		return
	}

	if _, err := os.Stat(logoPath); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "logo_not_found")
			return
		}
		h.logger.Error("failed to stat logo", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_serve_logo")
		return
	}

	ext := strings.ToLower(filepath.Ext(logoPath))
	contentType := "application/octet-stream"
	if ext == ".svg" {
		contentType = "image/svg+xml"
	} else if mt := mime.TypeByExtension(ext); mt != "" {
		contentType = mt
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeFile(w, r, logoPath)
}

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

	filePath, err := h.contentProvider.GetAssetPath(r.Context(), docName, assetPath)
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

	access, err := h.contentProvider.GetAssetAccess(r.Context(), docName, assetPath)
	if err != nil {
		if errors.Is(err, domain.ErrDocNotFound) {
			writeError(w, http.StatusNotFound, "doc_not_found")
			return
		}
		if errors.Is(err, domain.ErrAssetNotFound) {
			writeError(w, http.StatusNotFound, "asset_not_found")
			return
		}
		h.logger.Error("failed to resolve asset access",
			slog.String("error", err.Error()),
			slog.String("doc", docName),
			slog.String("asset", assetPath),
		)
		writeError(w, http.StatusInternalServerError, "failed_to_resolve_asset")
		return
	}

	if access.Access == "password" && !h.hasValidCookie(r, docName, access.Scope) {
		writePasswordRequired(w, access.Scope)
		return
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	contentType := "application/octet-stream"
	if mt := mime.TypeByExtension(ext); mt != "" {
		contentType = mt
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", path.Base(assetPath)))
	http.ServeFile(w, r, filePath)
}
