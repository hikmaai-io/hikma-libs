// ABOUTME: Log correlation across an authenticated, single HTTP hop.
// ABOUTME: The wire contains a fresh traceparent only: no caller ID, baggage or sampling request.
package hlog

import (
	"context"
	"crypto/rand"
	"net/http"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// StartIsolatedHop sets a new W3C traceparent on an outbound request and returns
// its local logging context. Unlike normal trace propagation it does not copy the
// caller's trace identity, sampling decision or vendor data onto the private hop.
// Use an independently constructed request, not a clone of inbound headers.
func StartIsolatedHop(ctx context.Context, header http.Header) context.Context {
	var id trace.TraceID
	for !id.IsValid() {
		rand.Read(id[:])
	}
	ctx = localHop(ctx, id)
	header.Del("Tracestate")
	header.Del("Baggage")
	propagation.TraceContext{}.Inject(ctx, propagation.HeaderCarrier(header))
	return ctx
}

// ReceiveIsolatedHop joins a single bounded traceparent only after peer
// authentication. Allocate a different local span; do not inherit tracestate,
// baggage or sampling. An older peer, duplicate or malformed header does not
// establish a link. Correlation is neither identity nor an application receipt.
func ReceiveIsolatedHop(ctx context.Context, header http.Header) context.Context {
	values := header.Values("Traceparent")
	if len(values) != 1 || len(values[0]) != 55 || values[0][:3] != "00-" {
		return ctx
	}
	// Extract in an empty context: invalid input must not appear to parse by
	// falling back to a pre-existing local span context.
	peer := propagation.TraceContext{}.Extract(context.Background(), propagation.MapCarrier{"traceparent": values[0]})
	id := trace.SpanContextFromContext(peer).TraceID()
	if !id.IsValid() {
		return ctx
	}
	return localHop(ctx, id)
}

func localHop(parent context.Context, id trace.TraceID) context.Context {
	var span trace.SpanID
	for !span.IsValid() {
		rand.Read(span[:])
	}
	return trace.ContextWithSpanContext(parent, trace.NewSpanContext(trace.SpanContextConfig{TraceID: id, SpanID: span}))
}
