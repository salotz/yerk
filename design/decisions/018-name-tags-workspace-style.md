# 018. `name-tags` workspace style (+ inline params)

## Status

Accepted (2026-10-02)

## Context

Operators need layouts where the **main** checkout is a bare project directory
(`…/projects/wumpus`) and **session** checkouts sit beside it with a tag
(`…/projects/wumpus__feat`). Some projects split main into a special home
(`~/.bimker`) while other replicas live under a devel tree.

Phase 2 locked the name `name-tags` and path math; path math was deferred until
this phase. Placement already merges a **style name** string; this ADR ships
the math and **inline parameter tables**.

## Decision

### Style name

`name-tags` (hyphenated; same as placement vocabulary).

### Path math (default, no params)

Let **container** = host project workspace path (ADR 014:
`[domains]` default `<root>/<name>` or `[[projects]]` path). That container is
what `yerk path <project>`, `workspace ensure`, and project status PATH report.

| Role | Path |
| --- | --- |
| Project workspace | **container** (parent of bare main + tagged siblings) |
| Main replica | `container/<project-name>` |
| Other replica `R` | `container/<project-name>__R` |

- Separator is **`__`** (not configurable in this phase).
- Main is **not** `container/main`. The main distinguisher is still the catalog
  `default_replica` (fallback `main`) for **ids** and status; only the **path
  leaf** is the bare project name.
- `yerk path <id> <default>` → bare main dir; `path <id> feat` → tagged sibling.

### Inline parameters (optional)

TOML **string or table** for `workspace_style` on catalog rows and host
`[[projects]]` (and host `[workspace].style` remains a **string** only for the
global ambient default):

```toml
workspace_style = "name-tags"

workspace_style = { style = "name-tags", main_dir = "~/.bimker", replica_dir = "~/tree/personal/devel/bimker" }
```

| Key | Meaning |
| --- | --- |
| `style` | Base style name (required in table form) |
| `main_dir` | Absolute or `~/…` override for the **main** checkout only |
| `replica_dir` | Absolute or `~/…` **container** for tagged non-main replicas (and default container for tags when set) |

When `main_dir` is set, main path = expanded `main_dir` (not
`container/<name>`). When `replica_dir` is set, tagged replicas use
`replica_dir/<name>__R`. Project workspace path remains the host workspace
path (ADR 014); if only params supply dirs, host path should still be set to
the logical container (usually `replica_dir` or parent of main) so ensure/status
have a stable workspace root.

**No** named derivation registry (`[workspace.styles.*]`).

### Placement merge

- Effective payload includes **style name** + optional **params** (main_dir,
  replica_dir) from the winning ambient layer that supplied them.
- Host **state** continues to bind **style name only** (ADR 013). Params remain
  ambient (catalog / host row / dir-local when present). Bound style name still
  wins over ambient name with the usual warning.
- CLI `--workspace-style` remains a **plain style name** (no inline table on
  CLI in this phase).

### Live replica discovery

- Main path present → include default/main distinguisher.
- Scan tag container for `<name>__*` live checkouts → other distinguishers.
- Lookup path walk matches bare main or tagged sibling roots.

### Other styles

`workspace-dir` and `project-dir` unchanged. Params on those styles are
**ignored** (or error if set — prefer **error** for unknown keys / params on
styles that do not accept them).

## Consequences

- `internal/workspace` implements name-tags + Layout params.
- Config/catalog decode flexible `workspace_style` values.
- Docs + examples show name-tags; dogfood bimker-like rows use params.
- Tests use temp dirs only (ADR 007).

## Related

- [013](./013-placement-policy-and-host-state.md) — placement layers
- [014](./014-host-local-path-model.md) — host workspace path
- [016](./016-replica-create.md) — create under any style path
- [domain-and-near-term.md](../domain-and-near-term.md)
