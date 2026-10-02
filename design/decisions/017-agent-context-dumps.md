# 017. Agent context dumps (`yerk context`)

## Status

Accepted (2026-10-02)

## Context

AI agents and session managers need a **single structured dump** of yerk
vocabulary, host paths, and “where am I?” without scraping help text or
chaining many CLI calls. Phase 3–4 already provide `get` / `lookup` and
`config resolve`. This ADR defines the **compose** surface.

## Decision

### Commands

```text
yerk context                 # tool-wide static + host paths
yerk context dir [path]      # directory → project/replica context
```

- Default path for `dir` is process **cwd**.
- Shared `--output json` (same flag as get/lookup). Human text is default.
- No TTY auto-json required.

### Payloads (`internal/api`, `apiVersion: yerk/v1`)

| Kind | Role |
| --- | --- |
| `ToolContext` | Vocabulary, command map, XDG/env pointers, how-to hints |
| `DirContext` | Lookup of path + project/replica info + short placement + optional status overall |

**JSON key stability (v0):** best-effort under `apiVersion: yerk/v1`. Casual
renames are forbidden; breaking changes bump `apiVersion` or release notes
(no iron-clad pre-1.0 guarantee).

### Composition (no duplicated business logic)

```text
context
  → static help sections + config.Dir / state.Dir / catalog path
  → version identity

context dir
  → abs path
  → project.LookupPath (universal: prefer replica)
  → ProjectInfo | ReplicaInfo (already on lookup)
  → ConfigResolve for the matched project (compact: effective style + bound + warnings)
  → ProjectStatus overall + replica names when cheap (presence-oriented; change optional)
```

CLI `RunE` stays thin: load → build → print.

### Human output

- **Tool:** labeled sections (commands, paths, env primary knobs, vocabulary).
- **Dir:** key/value header + nested project/replica block; pointer to
  `yerk status <id>` for full change tables.

### Non-goals

- Full markdown docs dump
- Mutating placement or state
- Parallel probes
- OS `yerk://` handler

## Consequences

- Agents can bootstrap with one call.
- `get` / `lookup` / `config resolve` remain the precise single-purpose tools.
- Future: optional `--with-change` on dir context; yaml output with Phase 8.

## Related

- [011](./011-api-resources.md) — resource kinds
- [015](./015-get-and-lookup.md) — get/lookup composition
- [013](./013-placement-policy-and-host-state.md) — placement
- [domain-and-near-term.md](../domain-and-near-term.md)
