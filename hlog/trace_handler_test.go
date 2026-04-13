// ABOUTME: Tests for TraceHandler including the critical WithAttrs/WithGroup delegation.
// ABOUTME: Uses OTel trace API to create mock span contexts for trace_id injection.

package hlog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	oteltrace "go.opentelemetry.io/otel/trace"
)

// validSpanContext returns a context with a valid OTel span context containing
// deterministic trace and span IDs.
func validSpanContext() context.Context {
	tid := oteltrace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	sid := oteltrace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	sc := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: oteltrace.FlagsSampled,
		Remote:     true,
	})
	return oteltrace.ContextWithRemoteSpanContext(context.Background(), sc)
}

func parseJSON(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %s\nerror: %v", data, err)
	}
	return m
}

func TestTraceHandler_WithValidContext(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(NewTraceHandler(inner))

	logger.InfoContext(validSpanContext(), "traced")

	m := parseJSON(t, buf.Bytes())
	if m["trace_id"] != "0102030405060708090a0b0c0d0e0f10" {
		t.Errorf("trace_id = %v, want 0102030405060708090a0b0c0d0e0f10", m["trace_id"])
	}
	if m["span_id"] != "1112131415161718" {
		t.Errorf("span_id = %v, want 1112131415161718", m["span_id"])
	}
}

func TestTraceHandler_WithoutContext(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(NewTraceHandler(inner))

	logger.Info("no-trace")

	m := parseJSON(t, buf.Bytes())
	if _, ok := m["trace_id"]; ok {
		t.Error("trace_id should not be present without span context")
	}
	if _, ok := m["span_id"]; ok {
		t.Error("span_id should not be present without span context")
	}
}

func TestTraceHandler_WithEmptyContext(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(NewTraceHandler(inner))

	logger.InfoContext(context.Background(), "empty-ctx")

	m := parseJSON(t, buf.Bytes())
	if _, ok := m["trace_id"]; ok {
		t.Error("trace_id should not be present with empty context")
	}
}

// TestTraceHandler_WithAttrsPreservesInjection tests the "With Trap":
// logger.With("key", "val") must return a logger that still injects trace IDs.
func TestTraceHandler_WithAttrsPreservesInjection(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(NewTraceHandler(inner))

	derived := logger.With("component", "auth")
	derived.InfoContext(validSpanContext(), "with-attrs")

	m := parseJSON(t, buf.Bytes())
	if m["trace_id"] != "0102030405060708090a0b0c0d0e0f10" {
		t.Errorf("trace_id lost after With(): %v", m["trace_id"])
	}
	if m["component"] != "auth" {
		t.Errorf("component = %v, want auth", m["component"])
	}
}

// TestTraceHandler_WithGroupPreservesInjection tests that WithGroup
// preserves trace injection. With slog semantics, record attrs (including
// injected trace_id) are nested inside the group.
func TestTraceHandler_WithGroupPreservesInjection(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(NewTraceHandler(inner))

	grouped := logger.WithGroup("request")
	grouped.InfoContext(validSpanContext(), "with-group", "method", "GET")

	m := parseJSON(t, buf.Bytes())
	group, ok := m["request"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'request' group in output, got: %v", m)
	}
	if group["trace_id"] != "0102030405060708090a0b0c0d0e0f10" {
		t.Errorf("trace_id lost after WithGroup(): %v", group["trace_id"])
	}
	if group["method"] != "GET" {
		t.Errorf("method lost after WithGroup(): %v", group["method"])
	}
}

func TestTraceHandler_Enabled(t *testing.T) {
	inner := slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelWarn})
	h := NewTraceHandler(inner)

	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("info should be disabled at warn level")
	}
	if !h.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("warn should be enabled at warn level")
	}
}
