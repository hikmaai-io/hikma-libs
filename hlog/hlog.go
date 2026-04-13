// ABOUTME: Shared structured logging for HikmaAI Go services.
// ABOUTME: JSON/text slog handler with automatic trace_id/span_id injection from OTel context.

package hlog

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Options configures the logger created by New.
type Options struct {
	// Level sets the minimum log level. Accepted values: "debug", "info",
	// "warn", "error". Defaults to "info" on empty or unknown input.
	Level string

	// Format selects the output format. "text" uses slog.NewTextHandler,
	// anything else (including empty) uses slog.NewJSONHandler.
	Format string

	// Service is added as a default "service" attribute on every log record.
	// Leave empty to omit.
	Service string

	// Writer is the output destination. Defaults to os.Stderr when nil.
	Writer io.Writer
}

// New creates a *slog.Logger with a TraceHandler wrapping the format handler.
// The TraceHandler auto-injects trace_id and span_id from the OTel span
// context when callers use the *Context log methods (InfoContext, etc.).
func New(opts Options) *slog.Logger {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}

	hopts := &slog.HandlerOptions{Level: ParseLevel(opts.Level)}

	var inner slog.Handler
	if opts.Format == "text" {
		inner = slog.NewTextHandler(w, hopts)
	} else {
		inner = slog.NewJSONHandler(w, hopts)
	}

	logger := slog.New(NewTraceHandler(inner))
	if opts.Service != "" {
		logger = logger.With("service", opts.Service)
	}
	return logger
}

// ParseLevel converts a level string to slog.Level.
// Returns slog.LevelInfo on empty or unrecognized input.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
