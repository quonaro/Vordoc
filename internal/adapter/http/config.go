package http

import (
	"log/slog"
	"net/http"

	"vordoc/internal/domain"
)

// ConfigHandler exposes the root content configuration.
type ConfigHandler struct {
	contentProvider domain.ContentProvider
	logger          *slog.Logger
}

// NewConfigHandler creates a new config handler.
func NewConfigHandler(contentProvider domain.ContentProvider, logger *slog.Logger) *ConfigHandler {
	return &ConfigHandler{contentProvider: contentProvider, logger: logger}
}

// GetConfig returns the root content configuration.
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.contentProvider.GetRootConfig(r.Context())
	if err != nil {
		h.logger.Error("failed to load root config", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_load_config")
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

// GetText returns the UI text configuration.
func (h *ConfigHandler) GetText(w http.ResponseWriter, r *http.Request) {
	text, err := h.contentProvider.GetUIText(r.Context())
	if err != nil {
		h.logger.Error("failed to load ui text", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed_to_load_config")
		return
	}
	writeJSON(w, http.StatusOK, text)
}
