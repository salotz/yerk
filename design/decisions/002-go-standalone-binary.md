# 002. Go standalone binary

## Status

Accepted (2026-09-25)

## Context

The operator is not primarily a Go developer, but wants:

- a shippable standalone binary
- mostly AI-agent-driven implementation
- no heavy runtime install for end users of the tool

## Decision

Implement `yerk` in Go with:

- module path `github.com/salotz/yerk`
- entrypoint `cmd/yerk`
- libraries under `internal/…`
- `go build` producing a single static-friendly binary

Tooling pins Go via `mise.toml`. Caches default under `.local/` so clones work
in restricted environments.

## Consequences

- Agents should prefer small packages and tests next to path logic
- Python remains available for *repo* task scripts if needed; product code is Go
- Cross-compile and release packaging can be added later without redesign
