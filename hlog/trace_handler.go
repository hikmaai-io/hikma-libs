// ABOUTME: slog.Handler wrapper that injects trace_id and span_id from OTel span context.
// ABOUTME: Delegates to any inner handler; WithAttrs/WithGroup re-wrap to preserve injection.

package hlog

import (
	"context"
	"log/slog"

	oteltrace "go.opentelemetry.io/otel/trace"
)

// TraceHandler wraps a slog.Handler and prepends trace_id and span_id
// attributes when a valid OTel span context is present in the context.
// Callers must use the *Context log methods (InfoContext, WarnContext, etc.)
// for trace injection to activate.
type TraceHandler struct {
	inner slog.Handler
}

// NewTraceHandler creates a TraceHandler wrapping inner.
func NewTraceHandler(inner slog.Handler) *TraceHandler {
	return &TraceHandler{inner: inner}
}

// Enabled delegates to the inner handler.
func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle injects trace_id and span_id if the context contains a valid
// OTel span context, then delegates to the inner handler.
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	sc := oteltrace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.inner.Handle(ctx, r)
}

// WithAttrs returns a new TraceHandler wrapping the inner handler's WithAttrs
// result. This preserves trace injection on derived loggers created via
// logger.With("key", "val").
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup returns a new TraceHandler wrapping the inner handler's WithGroup
// result. This preserves trace injection on grouped loggers.
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{inner: h.inner.WithGroup(name)}
}
