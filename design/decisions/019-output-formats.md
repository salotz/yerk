# 019. Shared `--output` formats

## Status

Accepted (2026-10-02)

## Context

ADR 011 introduced model-driven `internal/api` resources so printers could
share one shape. ADR 015 / 017 shipped **`--output json`** on get/lookup/context
(and config resolve) with human text as the default. Plan polish asked for a
broader surface: **`json | yaml | table`**.

Operators and agents need:

- Stable machine documents without scraping tables
- YAML where tooling prefers it (same field names as JSON)
- An explicit **table** value that means “human layout” (not a second machine schema)

Status remained table-only; agents still scraped or ran get repeatedly.

## Decision

### Flag

Shared flag name: **`--output`**.

| Value | Meaning |
| --- | --- | 
| omit / empty | Command default human layout (key/value or table) |
| `table` | Same as human default (explicit) |
| `json` | One indented JSON document (`apiVersion` / `kind` on resources) |
| `yaml` | One YAML document; field names match `json` tags (ADR 011) |

Unsupported values error. No TTY auto-json.

### Commands (this phase)

| Command | Structured payload |
| --- | --- |
| `get` / `lookup` / noun get\|lookup | `ProjectInfo` or `ReplicaInfo` |
| `context` / `context dir` | `ToolContext` / `DirContext` |
| `config resolve` | `ConfigResolve` |
| `status` | single project → `ProjectStatus`; single replica → `ReplicaStatus`; multi → JSON/YAML **array** of `ProjectStatus` |

Human status still prints preamble lines (`workspace.style=…`) only for non-structured output.

### Implementation notes

- Validate once via `validateOutputFlag`; encode via `writeStructured`.
- YAML via `gopkg.in/yaml.v3` using the same structs (json names).
- No separate “table schema”; `table` does not invent columns beyond the human printer.

### Out of scope

- JSON Schema / OpenAPI generation
- Streaming / NDJSON lists
- Changing default human status columns

## Consequences

- Agents can pick json or yaml uniformly on read commands and status.
- ADR 015/017 wording that limited values to `json` is extended by this ADR.
- Adding structured output to a new command should reuse the same flag vocabulary.

## Related

- [011](./011-api-resources.md) — resource types
- [015](./015-get-and-lookup.md) — get/lookup output
- [017](./017-agent-context-dumps.md) — context dumps
