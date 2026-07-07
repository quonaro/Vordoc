package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"vordoc/internal/domain"
	"vordoc/internal/service"
)

// DocsHandler handles documentation read endpoints.
type DocsHandler struct {
	contentProvider domain.ContentProvider
	passwordService *service.PasswordService
	cookieSecret    []byte
	logger          *slog.Logger
}

// NewDocsHandler constructs a docs handler.
func NewDocsHandler(contentProvider domain.ContentProvider, passwordService *service.PasswordService, cookieSecret string, logger *slog.Logger) *DocsHandler {
	return &DocsHandler{
		contentProvider: contentProvider,
		passwordService: passwordService,
		cookieSecret:    []byte(cookieSecret),
		logger:          logger,
	}
}

// docSummary is a lightweight list entry for a documentation.
type docSummary struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Access      string `json:"access,omitempty"`
}

// searchResult is a single search hit exposed to the frontend.
type searchResult struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	Snippet string `json:"snippet,omitempty"`
}

// passwordRequiredResponse is returned when a protected resource is accessed.
type passwordRequiredResponse struct {
	Error            string `json:"error"`
	PasswordRequired bool   `json:"password_required"`
	Scope            string `json:"scope"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func writePasswordRequired(w http.ResponseWriter, scope string) {
	writeJSON(w, http.StatusForbidden, passwordRequiredResponse{
		Error:            "password_required",
		PasswordRequired: true,
		Scope:            scope,
	})
}

func splitDocPath(path string) (doc string, page string, ok bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", "", false
	}
	return parts[0], strings.Join(parts[1:], "/"), true
}
