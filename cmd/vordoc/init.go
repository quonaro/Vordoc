// Package main is the entry point for the Vordoc CLI and server.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"vordoc/shared/config"
	"vordoc/shared/logging"

	"github.com/joho/godotenv"
	"github.com/quonaro/lota/engine"
)

//go:embed init-templates/index.md
var initIndexMD []byte

func initDoc(_ context.Context, nctx engine.NativeContext) error {
	_ = godotenv.Load()

	logger := logging.New("vordoc", config.LogsConfig{Level: "info", Type: "pretty"})
	cfg := config.LoadFromEnv(logger)

	name := nctx.Args["name"]
	if name == "" {
		name = "welcome"
	}

	docDir := filepath.Join(cfg.Content.Root, name)
	if err := os.MkdirAll(docDir, 0o755); err != nil { // #nosec G301 — content directory must be readable
		return fmt.Errorf("creating doc directory: %w", err)
	}

	configPath := filepath.Join(docDir, "config.yaml")
	if err := writeIfNotExists(configPath, []byte(initConfigYAML(name))); err != nil {
		return err
	}

	indexPath := filepath.Join(docDir, "index.md")
	if err := writeIfNotExists(indexPath, initIndexMD); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(nctx.Stdout, "initialized documentation '%s' in %s\n", name, docDir)
	return nil
}

func writeIfNotExists(path string, data []byte) error {
	if _, err := os.Stat(path); err == nil {
		_, _ = fmt.Fprintf(os.Stdout, "skipped %s (already exists)\n", path)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking %s: %w", path, err)
	}

	// #nosec G306 — default templates must be readable
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	_, _ = fmt.Fprintf(os.Stdout, "created %s\n", path)
	return nil
}

func initConfigYAML(_ string) string {
	return "title: \"Добро пожаловать в Vordoc\"\ndescription: \"Обзор возможностей Vordoc\"\n"
}
