# 015. Project/replica get and path lookup

## Status

Accepted (2026-10-01)

## Context

Agents and session managers need a uniform **read** of catalog identity, host
paths, presence, and effective placement — not only `status` tables or the last
path segment of cwd. ADR 011 introduced API resources; ADR 012 fixed identity;
ADR 013/014 supply effective style and workspace paths. This ADR defines the
CLI surface and info payloads for forward **get** (id → resource) and reverse
**lookup** (path → id + resource).

## Decision

### Command tree

Ship **both** universal verbs and noun-scoped verbs:

| Command | Input | Result |
| --- | --- | --- |
| `yerk get <id>` | bare id / short / `yerk://…` | project **or** replica info by id shape |
| `yerk lookup <path>` | filesystem path | project or replica info by path match |
| `yerk project get <id>` | project id only | project info |
| `yerk project lookup <path>` | path | project info (path may be under a replica) |
| `yerk replica get <id>` | replica id, or project id + replica arg | replica info |
| `yerk replica lookup <path>` | path | replica info (must be under a replica root) |

- `get` with a **project** id (two segments / project URI) → **project** resource
  only. No auto-expand to default/main replica (ADR 012).
- `get` with a **replica** id (third segment) → **replica** resource.
- `project get` rejects a replica segment (use `replica get` or universal `get`).
- `replica get` requires an explicit replica (id third segment or second arg).

### Lookup semantics

1. Resolve the argument to an **absolute** path (expand `~` via existing helpers
   when needed; relative paths are relative to process cwd).
2. For each catalog project, compute **workspace** path (ADR 014 + placement)
   and, when the path lies under that workspace, a **candidate replica**
   distinguisher (first path segment under the workspace for shipped styles).
3. A path **matches a replica** when it equals the computed replica checkout
   path or is a subdirectory of it (walk-up / any depth under the root).
4. A path **matches a project** when it equals the workspace path or is under
   it (including under a replica).
5. If multiple projects match, the **longest** matched root wins (most specific).
6. `replica lookup` errors when the path is only under a workspace but not under
   any computed replica root.
7. `lookup` (universal) prefers replica info when a replica root matches;
   otherwise project info.

Unknown path (no catalog workspace contains it) → error.

### Payloads (`internal/api`)

New kinds (still `apiVersion: yerk/v1`):

| Kind | Role |
| --- | --- |
| `ProjectInfo` | Catalog identity + workspace path/presence + effective placement + URI |
| `ReplicaInfo` | Catalog identity + replica path/presence (+ optional light probe fields) + placement + URI |

Include at least: `uri`, `name`/`project`, `domain`, `remote`, `tags`,
`defaultReplica` (project), workspace or replica `path` + `presence`, and a
`placement` object (`style`, `bound`, optional `sources` / `warnings`).

Do **not** require a full git change probe for get/lookup by default (keep
reads cheap). Presence classification is enough; change/branch may be omitted
or left default.

### Output

- **Human-friendly** default: stable key/value (or short labeled) text on
  stdout.
- **`--output json`**: single JSON document of the info resource (indent OK).
- Shared flag name `--output`; values `json` | `yaml` | `table` (human default
  when omitted). See [019](./019-output-formats.md).

TTY auto-json is **not** required in this phase.

### Package boundaries

```text
CLI  →  config load + id resolve / path abs
     →  project.Resolver Get* / Lookup*
     →  api.ProjectInfo | api.ReplicaInfo
     →  print human | json
```

Business logic stays out of cobra `RunE` beyond flags, selection, and print.

## Consequences

- Session managers can take `yerk lookup <path>` as a stable id string, then
  `yerk get <uri> --output json` (or lookup `--output json`) without scraping
  tables.
- `status` remains the multi-row / change-oriented view; get is one-resource
  detail with placement; lookup defaults to id only.
- Phase 6 `context dir` should compose these helpers rather than reimplement
  walk-up.
- JSON field names on info kinds should not churn casually (same discipline as
  ADR 011); a later stability bump can move `apiVersion` if needed.

## Related

- [011](./011-api-resources.md) — API resource package
- [012](./012-identifiers-and-yerk-uri.md) — id / URI forms
- [013](./013-placement-policy-and-host-state.md) — effective placement
- [014](./014-host-local-path-model.md) — workspace path model
- [domain-and-near-term.md](../domain-and-near-term.md) — get verb
