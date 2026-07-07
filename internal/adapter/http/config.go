package http

import (
	"log/slog"
	"net/http"
)

// ConfigHandler exposes application runtime configuration.
type ConfigHandler struct {
	public map[string]any
	logger *slog.Logger
}

// NewConfigHandler creates a new config handler.
func NewConfigHandler(public map[string]any, logger *slog.Logger) *ConfigHandler {
	if public == nil {
		public = map[string]any{}
	}
	return &ConfigHandler{public: public, logger: logger}
}

// GetPublic returns the public runtime config with a 10-minute cache header.
func (h *ConfigHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=600")
	writeJSON(w, http.StatusOK, h.public)
}
