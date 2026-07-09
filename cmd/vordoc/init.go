// Package main is the entry point for the Vordoc CLI and server.
package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"vordoc/internal/adapter/content"
	"vordoc/internal/service"
	"vordoc/shared/config"
	"vordoc/shared/logging"

	"github.com/joho/godotenv"
	"github.com/quonaro/lota/engine"
)

//go:embed all:init-templates
var initTemplatesFS embed.FS

const (
	memberPassword = "member"
	adminPassword  = "admin"
)

func initDoc(ctx context.Context, nctx engine.NativeContext) error {
	_ = godotenv.Load()

	logger := logging.New("vordoc", config.LogsConfig{Level: "info", Type: "pretty"})
	cfg := config.LoadFromEnv(logger)

	name := nctx.Args["name"]
	if name == "" {
		name = "welcome"
	}

	memberHash, err := generateDemoHash(memberPassword)
	if err != nil {
		return fmt.Errorf("hashing member password: %w", err)
	}
	adminHash, err := generateDemoHash(adminPassword)
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}

	replacements := map[string]string{
		"__MEMBER_HASH__": memberHash,
		"__ADMIN_HASH__":  adminHash,
	}

	// #nosec G301 — content directory must be readable
	if err := os.MkdirAll(cfg.Content.Root, 0o755); err != nil {
		return fmt.Errorf("creating content root: %w", err)
	}

	if err := copyTemplateDir("init-templates/content", cfg.Content.Root, name, replacements); err != nil {
		return fmt.Errorf("copying templates: %w", err)
	}

	contentProvider := content.NewProvider(cfg.Content.Root, logger)
	if err := contentProvider.EnsureDefaults(ctx); err != nil {
		return fmt.Errorf("ensuring content defaults: %w", err)
	}

	_, _ = fmt.Fprintf(nctx.Stdout, "initialized documentation '%s' in %s\n", name, filepath.Join(cfg.Content.Root, name))
	return nil
}

func generateDemoHash(password string) (string, error) {
	svc := service.NewPasswordService()
	hash, err := svc.Hash(password)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func copyTemplateDir(srcRel, dstDir, name string, replacements map[string]string) error {
	entries, err := initTemplatesFS.ReadDir(srcRel)
	if err != nil {
		return fmt.Errorf("reading template dir %s: %w", srcRel, err)
	}

	for _, entry := range entries {
		srcPath := path.Join(srcRel, entry.Name())
		dstName := entry.Name()
		if srcRel == "init-templates/content" && entry.Name() == "welcome" {
			dstName = name
		}
		dstPath := filepath.Join(dstDir, dstName)

		if entry.IsDir() {
			if _, err := os.Stat(dstPath); err == nil {
				_, _ = fmt.Fprintf(os.Stdout, "warning: %s already exists, skipping\n", dstPath)
				continue
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("checking directory %s: %w", dstPath, err)
			}
			// #nosec G301 — content directory must be readable
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dstPath, err)
			}
			if err := copyTemplateDir(srcPath, dstPath, name, replacements); err != nil {
				return err
			}
			continue
		}

		data, err := initTemplatesFS.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("reading template file %s: %w", srcPath, err)
		}

		content := string(data)
		for old, newVal := range replacements {
			content = strings.ReplaceAll(content, old, newVal)
		}

		if err := writeIfNotExists(dstPath, []byte(content)); err != nil {
			return err
		}
	}

	return nil
}

func writeIfNotExists(path string, data []byte) error {
	if _, err := os.Stat(path); err == nil {
		_, _ = fmt.Fprintf(os.Stdout, "warning: %s already exists, skipping\n", path)
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
