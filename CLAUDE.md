# ABOUTME: HikmaAI shared Go libraries for cross-service concerns.
# ABOUTME: hlog provides structured logging with automatic OTel trace correlation.

# Hikma Libs

Shared Go libraries for HikmaAI services. Each subdirectory is an independent Go module.

## Modules

| Module | Path | Purpose |
|--------|------|---------|
| hlog | `github.com/hikmaai-io/hikma-libs/hlog` | Structured logging with OTel trace_id/span_id injection |

## hlog

`slog.Handler` wrapper that auto-injects `trace_id` and `span_id` from OTel span context. Provides `New()` constructor with JSON/text format, level parsing, and service field.

```go
logger := hlog.New(hlog.Options{
    Level:   "info",
    Format:  "json",
    Service: "my-service",
})
logger.InfoContext(ctx, "request processed", "status", 200)
// Output: {"time":"...","level":"INFO","msg":"request processed","trace_id":"abc...","span_id":"def...","service":"my-service","status":200}
```

Callers must use `*Context` log methods (`InfoContext`, `WarnContext`, etc.) for trace injection. Plain `Info()`/`Warn()` calls work but won't include trace IDs.

### Dependencies

OpenTelemetry trace and W3C propagation APIs only; no SDK or collector is required. For a privileged service hop that must not carry a caller's trace metadata, use `StartIsolatedHop` on a fresh outgoing request and `ReceiveIsolatedHop` **after** authentication. Use the returned contexts with `InfoContext`. A shared trace ID correlates logs; it is never a receipt or grant.

### Consumers

- hikma-mirsad (security gateway)
- hikma-wasit (backend API)

Both use `replace` directives during development:
```
replace github.com/hikmaai-io/hikma-libs/hlog => ../hikma-libs/hlog
```

## Commands

```bash
make check    # fmt + vet + vulncheck + test
make test     # tests with race detector
make lint     # golangci-lint
```

## Conventions

- Go 1.26.2
- Each module has its own `go.mod`
- Minimal dependencies (stdlib + OTel trace API)
- Feature branches -> PRs to dev -> main
- Conventional commits

## Related

- Workspace: `~/Development/hikmaAI/CLAUDE.md`
- ADR-011: Observability alignment (logging format spec)
