# 013. Placement policy and host project state

## Status

Accepted (2026-10-01)

## Context

Host-wide `[workspace].style` in `config.toml` is too coarse. Operators need
tree defaults (e.g. devel vs admin), sticky bindings after a project is
initialized on a host, and clear rules when ambient config drifts from that
binding. Tool-written state must not overwrite hand-edited config under
`$XDG_CONFIG_HOME`.

Identifiers (ADR 012) already separate **identity** (`yerk://domain/name`) from
filesystem paths. Placement must resolve **how** a known project sits on disk
without treating domain as a path root (path model: ADR 014).

## Decision

### Layered effective placement

Effective **workspace style** (and later style parameters / replica method) is
merged from layers. **Nearer / more specific wins** unless host project state
has bound the project (see conflicts).

| Order (low → high) | Source | Class |
| --- | --- | --- |
| 0 | Built-in default (`workspace-dir`) | ambient |
| 1 | Host `config.toml` `[workspace]` | ambient |
| 2 | Dir-local `.local/yerk/config.toml` (walk) | ambient |
| 3 | Catalog project row (`workspace_style`, …) | ambient |
| 4 | `YERK__WORKSPACE_STYLE` | ambient |
| 5 | **Host project state** (`state.json`) | **bound** (wins over ambient + warn) |
| 6 | Explicit CLI flags (e.g. `--workspace-style`) | **explicit** (error if ≠ bound) |

Dir-local files: multiple may contribute. **Closer to the anchor overrides
farther** (near wins) when merging ambient file layers before catalog/state.

### Dir-local discovery

1. **Anchor** = resolved project workspace path when the target is a known
   project; else an absolute path argument; else process **cwd**.
2. Walk **up** from the anchor’s directory toward **user home**, collecting each
   `<dir>/.local/yerk/config.toml` that exists.
3. **Stop at `$HOME`** (do not walk above the process user’s home).
4. Merge ambient style keys with near-wins among dir-locals, then apply host
   config underneath (host config is farther than any dir-local under home).

Dir-local is tree-scoped config, not an XDG tree.

### Host project state (XDG state)

| Tree | Role |
| --- | --- |
| `$XDG_CONFIG_HOME/yerk` | Operator prefs + catalog (unchanged) |
| `$XDG_STATE_HOME/yerk` | Tool-written project bindings / state |
| `$XDG_CACHE_HOME/yerk` | Disposable caches (later) |
| `$XDG_DATA_HOME/yerk` | Optional durable data (later) |

Per-project binding file:

```text
$XDG_STATE_HOME/yerk/projects/<domain>/<project>/state.json
```

- Override root: **`YERK__STATE_DIR`** (replaces `$XDG_STATE_HOME/yerk`).
- Format: JSON (machine-oriented). Schema version via `apiVersion` (`yerk/v1`).
- MVP fields: bound `workspaceStyle`, optional timestamps; extend later without
  silent relocation of replicas.

### Initialize (write binding)

Write or refresh binding on first successful:

- `yerk workspace ensure` for that project, or
- materialize under the project (`yerk materialize`, later `replica create`)

Rules:

- Ensure on an **already-initialized** project: **no-op** for state (success);
  still ensures the workspace directory exists.
- Optional later `yerk project init`: fail-if-exists (explicit); not required now.
- Binding records the **effective style at init** (after ambient merge, before
  treating state as already bound).

### Conflicts

| Situation | Behavior |
| --- | --- |
| Bound state style ≠ ambient (files / env) | **Keep bound style** for path math; emit a **warning** naming conflicting ambient sources. Do not hard-stop read or mutate commands solely for ambient drift. |
| Explicit CLI flag contradicts bound state | **Error** (no silent override). Rebind/migration is a later explicit feature. |
| `YERK__WORKSPACE_STYLE` vs bound state | Same class as files: **warn**, state wins. |

### Catalog optional placement fields

On `[[projects]]` (optional):

- `workspace_style` — string style name (inline table parameters: later style ADR)
- `replica_method` — `worktree` \| `clone` (consumed by replica create; ignored by path math until then)
- `default_replica` — unchanged (main replica distinguisher)

### Main replica

Prose synonym for the catalog **default replica** (hub checkout). Worktree-based
replica create requires main present (hard error if missing); path policy does
not auto-clone main.

### Which commands take placement flags

- **Mutate placement:** `materialize`, `workspace ensure`, later `replica create`
  (style and/or method where relevant).
- **Read-only:** `status`, `path`, later `get` / `lookup` / `config resolve` /
  `context` — **report** effective policy; do not change binding; no style flags
  that rebind.

### Known styles (this ADR)

| Style | Status |
| --- | --- |
| `workspace-dir` | shipped |
| `project-dir` | shipped |
| `name-tags` | name locked; path math in later style ADR / phase |

Unknown style names → **error** at resolve (clean failure until implemented).

## Consequences

- New packages: host state I/O; placement merge + dir-local walk.
- Layout path math consumes **effective** style (and later params), not only
  raw `config.Load()` style.
- Env registry + `.appinfo` gain `YERK__STATE_DIR` and `XDG_STATE_HOME`.
- `config resolve` (later) reuses the same merge pipeline for observability.
- Tests use temp `YERK__STATE_DIR` / config dirs and synthetic dir-local trees.
- No host-private paths in-repo examples (ADR 007).

## Related

- [003](./003-config-xdg-and-env.md) — XDG config / `YERK__`
- [004](./004-config-and-catalog-split.md) — config vs catalog
- [008](./008-domain-roots-and-relative-catalog-paths.md) — superseded path join (see ADR 014)
- [012](./012-identifiers-and-yerk-uri.md) — id segments used in state paths
- [014](./014-host-local-path-model.md) — workspace path without domain-root join
- [domain-and-near-term.md](../domain-and-near-term.md)
