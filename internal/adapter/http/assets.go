// Package http provides HTTP handlers and server wiring for Vordoc.
package http

import (
	"embed"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// AssetHandler serves the embedded frontend SPA.
type AssetHandler struct {
	fs     fs.FS
	logger *slog.Logger
}

// NewAssetHandler creates a handler that serves the embedded frontend build.
func NewAssetHandler(logger *slog.Logger) *AssetHandler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		logger.Error("failed to open embedded frontend dist", slog.String("error", err.Error()))
		panic(err)
	}
	return &AssetHandler{fs: sub, logger: logger}
}

// Serve serves a static asset or falls back to index.html for SPA routing.
func (h *AssetHandler) Serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	p := path.Clean(r.URL.Path)
	if p == "/" || p == "." || p == "" {
		h.serveIndex(w, r)
		return
	}
	p = strings.TrimPrefix(p, "/")

	file, err := h.fs.Open(p)
	if err != nil {
		h.serveIndex(w, r)
		return
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		h.serveIndex(w, r)
		return
	}

	if stat.IsDir() {
		h.servePath(w, r, path.Join(p, "index.html"))
		return
	}

	h.servePath(w, r, p)
}

func (h *AssetHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	h.servePath(w, r, "index.html")
}

func (h *AssetHandler) servePath(w http.ResponseWriter, r *http.Request, filePath string) {
	data, err := fs.ReadFile(h.fs, filePath)
	if err != nil {
		if filePath != "index.html" {
			h.serveIndex(w, r)
			return
		}
		h.logger.Error("failed to read embedded index.html", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_serve_asset")
		return
	}

	h.setCacheHeaders(w, filePath)

	contentType := "application/octet-stream"
	if ext := path.Ext(filePath); ext != "" {
		if mt := mime.TypeByExtension(ext); mt != "" {
			contentType = mt
		}
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		// #nosec G705 — data is read from the embedded filesystem
		if _, err := w.Write(data); err != nil {
			h.logger.Error("failed to write response", slog.String("error", err.Error()))
		}
	}
}

func (h *AssetHandler) setCacheHeaders(w http.ResponseWriter, filePath string) {
	if filePath == "index.html" {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		return
	}
	if strings.HasPrefix(filePath, "_nuxt/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
}
