package hlog

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestIsolatedAuthenticatedHop(t *testing.T) {
	ctx, cancel := context.WithCancel(validSpanContext())
	defer cancel()
	h := http.Header{"Tracestate": {"test=private"}, "Baggage": {"private=value"}, "Traceparent": {"invalid", "ambiguous"}}
	send := StartIsolatedHop(ctx, h)
	a := trace.SpanContextFromContext(send)
	if !a.IsValid() || a.TraceID() == trace.SpanContextFromContext(ctx).TraceID() || a.IsSampled() || len(h) != 1 || len(h.Values("Traceparent")) != 1 || h.Get("Traceparent") != "00-"+a.TraceID().String()+"-"+a.SpanID().String()+"-00" {
		t.Fatal("did not start a fresh bounded hop")
	}
	if trace.SpanContextFromContext(StartIsolatedHop(ctx, make(http.Header))).TraceID() == a.TraceID() {
		t.Fatal("trace ID reused between writes")
	}
	h.Set("Tracestate", "test=private")
	h.Set("Baggage", "private=value")
	h.Set("Traceparent", strings.TrimSuffix(h.Get("Traceparent"), "00")+"01")
	recv := ReceiveIsolatedHop(ctx, h)
	b := trace.SpanContextFromContext(recv)
	if b.TraceID() != a.TraceID() || !b.IsValid() || a.SpanID() == b.SpanID() || b.TraceState().Len() != 0 || b.IsSampled() || b.IsRemote() {
		t.Fatal("peer context was not isolated")
	}
	var out bytes.Buffer
	New(Options{Writer: &out}).InfoContext(recv, "received")
	record := parseJSON(t, out.Bytes())
	if record["trace_id"] != a.TraceID().String() || record["span_id"] != b.SpanID().String() || strings.Contains(out.String(), "private") {
		t.Fatal("incorrect log context")
	}
	cancel()
	if send.Err() != context.Canceled || recv.Err() != context.Canceled {
		t.Fatal("lost parent cancellation")
	}
}

func TestIsolatedHopIgnoresInvalidOrAbsentLink(t *testing.T) {
	h := make(http.Header)
	StartIsolatedHop(t.Context(), h)
	for _, value := range [][]string{nil, {"private"}, {strings.Repeat("x", 512)}, {strings.Repeat("0", 55)}, {"00-" + strings.Repeat("0", 32) + "-1234567890abcdef-00"}, {h.Get("Traceparent"), h.Get("Traceparent")}} {
		header := http.Header{"Traceparent": value}
		ctx := validSpanContext()
		if ReceiveIsolatedHop(ctx, header) != ctx {
			t.Fatal("malformed input changed context")
		}
	}
}
