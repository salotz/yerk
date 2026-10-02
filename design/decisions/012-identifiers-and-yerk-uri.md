# 012. Identifiers and `yerk://` URI currency

## Status

Accepted (2026-09-30)

## Context

Commands historically addressed catalog rows by **short project name** only
(`yerk status yerk`, `yerk path yerk main`). That collides when two domains
share a name, ties identity to filesystem layout, and gives agents no stable
handle independent of workspace style.

The product needs:

- Layout-independent **project** and **replica** identifiers
- One canonical URI form for machine output and internal APIs
- Human/CLI bare shortcuts that **expand** to that URI
- Clear ambiguity errors (never silent pick)

ADR 011 already stamps `apiVersion: yerk/v1` on resources; resources should
carry the canonical URI once identity is defined.

## Decision

### Two spellings only

| Form | Example | Role |
| --- | --- | --- |
| Bare identifier | `personal/wumpus`, `personal/wumpus/main` | CLI ergonomics |
| Canonical URI | `yerk://personal/wumpus`, `yerk://personal/wumpus/main` | Internal currency + machine output |

- No third form such as `yerk:…` (opaque, no `//`).
- Query strings and fragments are **rejected** in v1.
- Everything internal expands bare → canonical URI.

### Grammar

```text
domain      = 1*segment
project     = 1*segment
replica     = 1*segment          ; optional third path segment
segment     = 1*( ALPHA / DIGIT / "-" / "_" / "." )
bare        = domain "/" project [ "/" replica ]
uri         = "yerk://" bare
```

- Segments are non-empty; no empty `//` path parts.
- Leading/trailing `/` rejected on bare form.
- Case-sensitive match against catalog (no case folding).
- Forbidden in segments (v1): whitespace, `?`, `#`, `@`, `\`, and path
  separators other than the structural `/` between domain/project/replica.

### Catalog identity

- Keep catalog fields **`name` + `domain`** (not a single `id` field).
- Derived bare project id: `domain/name`.
- Derived project URI: `yerk://domain/name`.
- Replica URI: `yerk://domain/name/replica`.
- **`domain` is required** on every project row. Empty domain → load/validate
  error. No implicit default domain.

### Short names

- A single path segment (`wumpus`, or `wumpus/main`) is a **short** form.
- Expand against the catalog: unique project name → that row’s domain.
- If two or more projects share the name → **error** listing candidates
  (`personal/wumpus`, `work/wumpus`). No silent default domain.

### Project vs replica

- Two-segment id/URI (`personal/wumpus`, `yerk://personal/wumpus`) means the
  **project** resource only.
- Replica requires an explicit third segment or a separate replica argument.
- Commands must **not** auto-expand a project id to the default/main replica
  when resolving identity (path/status may still choose a default replica for
  *probes* after the project is known — that is operation policy, not id
  expansion).

### CLI acceptance

Commands accept, where a project (and optional replica) is selected:

1. Canonical URI or bare id (including unique short forms)
2. Legacy positionals: `<project-name-or-id> [replica]` when still useful

Shared parse/expand lives in **`internal/id`**. Catalog resolve helpers live
beside load/validate (`internal/config`) and call into `id`.

### API resources

- `Project`, `ProjectStatus`, and `ReplicaStatus` carry a `uri` field set to
  the canonical `yerk://…` form for that resource.
- `apiVersion` remains `yerk/v1` (ADR 011). Casual renames of JSON keys are
  avoided; breaking shape changes bump version or release notes.

### Package boundaries

```text
user string  →  id.Parse / Expand  →  id.Ref
                     ↓
config.Catalog.Resolve  →  config.Project + id.Ref
                     ↓
api resources (uri set) / layout / git
```

Filesystem paths are **not** identifiers. Reverse path → id is a later
lookup feature (not defined here).

## Consequences

- Catalog validation tightens: non-empty `domain` on every project.
- Name-only `Catalog.Find` remains a low-level helper; CLI should prefer
  resolve/expand so `domain/name` and URIs work.
- Docs must teach id vs on-disk path; host placement is ADR 014 (optional
  domain roots / host `[[projects]]`), not catalog path.
- OS-level `yerk://` URL handler is **out of scope**.
- Host project state paths under XDG state should nest as
  `projects/<domain>/<project>/` using the same segment strings (placement
  ADR).

## Related

- [011](./011-api-resources.md) — resource shapes / `yerk/v1`
- [004](./004-config-and-catalog-split.md) — catalog file identity fields
- [008](./008-domain-roots-and-relative-catalog-paths.md) — prior path join
- [014](./014-host-local-path-model.md) — host placement; domain remains the **id** namespace
- [domain-and-near-term.md](../domain-and-near-term.md)
