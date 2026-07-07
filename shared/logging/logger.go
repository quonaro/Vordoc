package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"vordoc/shared/config"
)

// New creates a process logger with the given service name and config.
func New(name string, cfg config.LogsConfig) *slog.Logger {
	level := parseLevel(cfg.Level)
	logType := strings.ToLower(strings.TrimSpace(cfg.Type))

	serviceName := strings.TrimSpace(name)
	if serviceName == "" {
		serviceName = "vordoc"
	}

	var handler slog.Handler
	switch logType {
	case "pretty":
		handler = &minimalHandler{
			w:     os.Stdout,
			level: level,
		}
	case "text", "console":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	default:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	return slog.New(handler).With(
		slog.String("service", serviceName),
		slog.Int("pid", os.Getpid()),
	)
}

func parseLevel(raw string) slog.Level {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return slog.LevelInfo
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(candidate))); err == nil {
		return level
	}

	switch strings.ToLower(candidate) {
	case "warning":
		return slog.LevelWarn
	case "fatal":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// minimalHandler implements a minimal log format: [LEVEL] message attrs...
type minimalHandler struct {
	w     io.Writer
	level slog.Level
}

func (h *minimalHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *minimalHandler) Handle(_ context.Context, r slog.Record) error {
	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key != slog.TimeKey && a.Key != slog.LevelKey && a.Key != slog.MessageKey {
			attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
		}
		return true
	})

	levelStr := r.Level.String()
	if len(levelStr) > 4 {
		levelStr = levelStr[:4]
	}

	output := fmt.Sprintf("[%s] %s", levelStr, r.Message)
	if len(attrs) > 0 {
		output += " | " + strings.Join(attrs, " | ")
	}
	output += "\n"

	_, err := h.w.Write([]byte(output))
	return err
}

func (h *minimalHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *minimalHandler) WithGroup(_ string) slog.Handler {
	return h
}
