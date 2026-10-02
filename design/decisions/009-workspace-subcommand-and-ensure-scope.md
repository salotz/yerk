# 009. Workspace subcommand and ensure scope

## Status

Accepted (2026-09-28)

## Context

Near-term verbs distinguished **materialize workspace** (layout dirs) from
**materialize replica** (`clone`). The CLI still exposed a top-level
`yerk ensure` that created the **default replica** directory leaf
(`…/yerk/main`), not the project workspace (`…/yerk`). That mixed placement
style into “ensure workspace” and surprised operators who expected only the
catalog workspace path.

Path resolution already treats bare project name as the workspace
(`yerk path <project>` → workspace; `yerk path <project> <replica>` →
checkout). Ensure should match that boundary: layout for the workspace
folder, not a checkout leaf. Replica materialization remains `clone`
(and any future replica-oriented commands).

Grouping under a `workspace` parent also leaves room for further workspace
ops without growing the root command list.

## Decision

### CLI nesting

| Command | Role |
| --- | --- |
| `yerk workspace ensure <project>…` | Create named **project workspace** directories (`mkdir -p`) |
| `yerk workspace ensure --all` | Same for every catalog project (opt-in bulk) |
| (removed) `yerk ensure` | No top-level alias; use the nested command |

Bare `yerk workspace ensure` (no names, no `--all`) is an **error**. Bulk
materialize is never the default. `--all` and project names are mutually
exclusive.

Future workspace-only operations go under `yerk workspace …`.

### Ensure creates the workspace only

- Resolve host placement → absolute **project workspace** (optional domain
  root + name, or host `[[projects]]` path; ADR 014).
- `mkdir -p` that directory.
- **Do not** create the replica distinguisher leaf (`…/main`).
- **Do not** run git, resolve default branch, or take `--network`.

Replica checkout paths are created by `yerk materialize` (parents via
`EnsureParents`, leaf by `git clone`) or left missing until then.

### Internal helpers

- `workspace.EnsureDir` — project workspace (and general mkdir -p).
- `workspace.EnsureParents` — parents of a replica path before clone.
- `workspace.EnsureReplicaDir` — retained for intentional replica-leaf mkdir;
  not used by `workspace ensure`.

## Consequences

- Docs, domain language, envvars command paths, and agent surface lists use
  `yerk workspace ensure` / `--all` / `--tag`.
- Status / materialize still target default or named **replicas**; ensure is
  orthogonal and does not imply a replica distinguisher.
- Empty catalog with `--all` is an error; bare ensure without selection is
  always an error.

## Related

- [domain-and-near-term.md](../domain-and-near-term.md) — verbs table
- [014](./014-host-local-path-model.md) — workspace path math
- [004](./004-config-and-catalog-split.md) — config vs catalog split
- [013](./013-placement-policy-and-host-state.md) — style bind on ensure
