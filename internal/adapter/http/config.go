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
		writeError(w, http.StatusInternalServerError, "failed to load root config")
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}
