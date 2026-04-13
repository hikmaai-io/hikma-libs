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

Single dependency: `go.opentelemetry.io/otel/trace` (API only, no SDK).

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
