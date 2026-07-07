package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"vordoc/shared/config"
)

// Server wraps the HTTP API server.
type Server struct {
	server *http.Server
	logger *slog.Logger
}

// Config defines HTTP server settings.
type Config struct {
	Address           string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
}

// Handlers groups all HTTP handlers the server wires up.
type Handlers struct {
	Docs   *DocsHandler
	Config *ConfigHandler
}

// NewServer builds HTTP server with injected handlers.
func NewServer(cfg Config, appCfg config.Config, logger *slog.Logger, handlers Handlers) *Server {
	r := chi.NewRouter()

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Logging
	r.Use(newLoggingMiddleware(logger))

	// API routes: versioned content API under /api/v1, public config under /api.
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/config", handlers.Config.GetConfig)
		r.Get("/docs", handlers.Docs.ListDocs)
		r.Get("/logo", handlers.Docs.ServeLogo)
		r.Get("/*", handlers.Docs.GetDocOrPage)
		r.Post("/*", handlers.Docs.VerifyPassword)
	})
	r.Get("/api/config", handlers.Config.GetConfig)

	// Static themes with correct MIME types
	themesPath := appCfg.Content.ThemesDir
	if themesPath == "" {
		themesPath = "./themes"
	}
	r.Mount("/themes/", http.StripPrefix("/themes/", newCSSFileServer(http.Dir(themesPath))))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			logger.Error("failed to write health response", slog.String("error", err.Error()))
		}
	})

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           r,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	return &Server{server: srv, logger: logger}
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.server.Addr
}

// Start launches the HTTP server.
func (s *Server) Start() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("starting http server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down http server: %w", err)
	}
	s.logger.Info("http server stopped")
	return nil
}

func newLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info("request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.String("duration", time.Since(start).String()),
			)
		})
	}
}

// newCSSFileServer wraps http.FileServer to set correct Content-Type for CSS files.
func newCSSFileServer(fs http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 4 && path[len(path)-4:] == ".css" {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		}
		http.FileServer(fs).ServeHTTP(w, r)
	})
}
