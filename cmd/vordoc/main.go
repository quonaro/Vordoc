package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vordoc/internal/adapter/content"
	httpadapter "vordoc/internal/adapter/http"
	"vordoc/internal/service"
	"vordoc/shared/config"
	"vordoc/shared/logging"

	"golang.org/x/sync/errgroup"
)

func main() {
	logger := logging.New("vordoc", config.LogsConfig{Level: "info", Type: "pretty"})

	cfg := config.LoadFromEnv(logger)

	logger = logging.New("vordoc", cfg.App.Logs)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	contentProvider := content.NewProvider(cfg.Content.Root, logger)
	if err := contentProvider.EnsureDefaults(ctx); err != nil {
		logger.Error("failed to ensure content defaults", slog.String("error", err.Error()))
		os.Exit(1)
	}
	passwordService := service.NewPasswordService()

	handlers := httpadapter.Handlers{
		Docs:   httpadapter.NewDocsHandler(contentProvider, passwordService, cfg.Auth.PageSecret, logger),
		Config: httpadapter.NewConfigHandler(contentProvider, logger),
	}

	server := httpadapter.NewServer(httpadapter.Config{
		Address:           fmt.Sprintf(":%d", cfg.App.HTTPPort),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}, cfg, logger, handlers)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("starting http server", slog.String("address", server.Addr()))
		return server.Start()
	})

	g.Go(func() error {
		<-gCtx.Done()
		logger.Info("shutting down gracefully")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownGracetime)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server shutdown failed: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil && ctx.Err() == nil {
		logger.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("shutdown complete")
}
