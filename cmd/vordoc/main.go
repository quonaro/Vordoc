// Package main is the entry point for the Vordoc CLI and server.
package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"vordoc/internal/adapter/content"
	httpadapter "vordoc/internal/adapter/http"
	"vordoc/internal/service"
	"vordoc/shared/config"
	"vordoc/shared/logging"

	"github.com/joho/godotenv"
	"github.com/quonaro/lota/engine"
	"golang.org/x/sync/errgroup"
)

//go:embed cli.yml
var cliYAML []byte

var version = "dev"

func main() {
	builder := engine.NewBuilder("vordoc", cliYAML)
	builder.RegisterNative("run", runServer)
	builder.RegisterNative("pass", hashPassword)
	builder.RegisterNative("version", showVersion)
	builder.RegisterNative("init", initDoc)

	app, err := builder.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	args, showVersion, showHelp := parseGlobalFlags(os.Args[1:])
	if showHelp || len(args) == 0 {
		app.PrintHelp()
		return
	}
	if showVersion {
		_, _ = fmt.Fprintf(os.Stdout, "vordoc version %s\n", currentVersion())
		return
	}

	if err := app.Run(context.Background(), args); err != nil {
		var groupErr *engine.GroupError
		if errors.As(err, &groupErr) {
			app.PrintGroupHelp(groupErr.Groups)
			return
		}
		fmt.Fprintf(os.Stderr, "run: %v\n", err)
		os.Exit(1)
	}
}

func parseGlobalFlags(args []string) (remaining []string, showVersion, showHelp bool) {
	for _, a := range args {
		switch a {
		case "-v", "--version", "-V":
			showVersion = true
		case "-h", "--help":
			showHelp = true
		default:
			remaining = append(remaining, a)
		}
	}
	return
}

func currentVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				return s.Value[:7]
			}
		}
	}
	return "unknown"
}

func runServer(ctx context.Context, _ engine.NativeContext) error {
	_ = godotenv.Load()

	logger := logging.New("vordoc", config.LogsConfig{Level: "info", Type: "pretty"})
	cfg := config.LoadFromEnv(logger)
	logger = logging.New("vordoc", cfg.App.Logs)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	contentProvider := content.NewProvider(cfg.Content.Root, logger)
	if err := contentProvider.EnsureDefaults(ctx); err != nil {
		logger.Error("failed to ensure content defaults", slog.String("error", err.Error()))
		return err
	}
	passwordService := service.NewPasswordService()

	handlers := httpadapter.Handlers{
		Docs:   httpadapter.NewDocsHandler(contentProvider, passwordService, cfg.Auth.PageSecret, logger),
		Config: httpadapter.NewConfigHandler(contentProvider, logger),
		Assets: httpadapter.NewAssetHandler(logger),
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
		return err
	}

	logger.Info("shutdown complete")
	return nil
}

func hashPassword(_ context.Context, nctx engine.NativeContext) error {
	password := nctx.Args["password"]
	if password == "" {
		return errors.New("password argument is required")
	}

	svc := service.NewPasswordService()
	hash, err := svc.Hash(password)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	_, _ = fmt.Fprintln(nctx.Stdout, hash)
	return nil
}

func showVersion(_ context.Context, nctx engine.NativeContext) error {
	v := currentVersion()
	if v == "unknown" {
		_, _ = fmt.Fprintln(nctx.Stdout, "vordoc version unknown")
		return nil
	}
	_, _ = fmt.Fprintf(nctx.Stdout, "vordoc version %s\n", v)
	return nil
}
