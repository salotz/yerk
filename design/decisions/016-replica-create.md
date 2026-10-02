# 016. Replica create (method, prerequisites, idempotency)

## Status

Accepted (2026-10-01)

## Context

Agents and session managers need yerk-mediated **session spin-out**: a new
replica checkout under the project's effective layout without embedding
worktree-vs-clone policy in callers. Top-level **`yerk materialize`** already
bootstraps a replica from the **catalog remote** (git clone under the hood;
no `clone` product alias). That verb is not the same as spinning a second
checkout from an existing **main** hub.

Identifiers (ADR 012), placement (ADR 013), and host paths (ADR 014) are in
place. This ADR defines **`yerk replica create`**: method, flags,
prerequisites, and idempotency.

## Decision

### Command

```text
yerk replica create <project-id> <replica-name>
yerk replica create <project-id>/<replica-name>
yerk replica create yerk://domain/name/replica
```

- Replica distinguisher is **required** (no default-replica expansion).
- Single project only (no `--all` / `--tag` in this phase).
- Mutate placement: may take `--workspace-style` and `--method` (same conflict
  rules as other mutators — ADR 013).

### Methods

| Method | Behavior |
| --- | --- |
| `worktree` | `git worktree add` from the **main** replica checkout into the style path for `<replica-name>` |
| `clone` | `git clone` of the catalog **remote** into the style path (optionally `--branch` = replica name) |

- **Default method:** `worktree` when unset.
- **Catalog** optional `replica_method` (`worktree` \| `clone`) is ambient policy
  for create (same row field as ADR 013).
- **CLI** `--method worktree|clone` is explicit for this invocation.
- Precedence for method: CLI `--method` → catalog `replica_method` → default
  `worktree`. (Host state does not bind method in MVP.)
- Unknown method names → **error**.

### Main replica

- Prose synonym for catalog **`default_replica`** (fallback name `main` when
  unset and no network resolution is needed for create).
- **`worktree` method:** main checkout must be **present** (usable git
  checkout). If missing or invalid → **hard error** with an actionable message
  (bootstrap main via `yerk materialize` first). No auto-materialize of main
  in v1.
- **`clone` method:** does not require main; requires non-empty catalog
  `remote`.

### Branch / distinguisher

- Default: **replica name = git branch name**.
- **worktree:** new branch from main HEAD when missing; attach to existing
  branch when present.
- **clone:** remote rarely has the session branch; clone default HEAD, then
  `checkout -b <replica>` (or checkout existing local branch) so the
  distinguisher matches HEAD. Do not require the branch on the remote.
- Optional `--branch` / `--ref` that differs from the distinguisher is **out of
  MVP**.

### Idempotency / existing path

- If the destination is already a usable checkout (**present**) or otherwise
  non-empty / invalid → **refuse** (error). No silent skip.
- Optional later `--if-absent` may align with materialize's present-ok path;
  not required now.
- Contrast: `yerk materialize` treats **already present** as success (bootstrap
  idempotency). Create is spin-out and stays fail-if-exists for MVP.

### Side effects

- Ensures parent directories for the destination path.
- On success, **bind host project state** on first init under the project
  (same `BindOnInit` path as ensure/materialize — ADR 013).
- Prints the absolute replica path on stdout (machine-friendly, like
  materialize).

### Package boundaries

```text
CLI  →  id resolve + flags
     →  project.Resolver.CreateReplica
     →  placement paths + presence
     →  gitcmd.WorktreeAdd | gitcmd.Clone
```

`git worktree add` and branch ensure live on the git adapter (`internal/gitcmd`)
beside `Clone`.

## Consequences

- Session managers can spin worktrees without knowing layout math.
- Operators still use **`materialize`** for first/main (or any) remote clone.
- Status/get/lookup see new replicas once present on disk.
- Future: bulk create, `--if-absent`, method in host state, branch≠name.

## Related

- [009](./009-workspace-subcommand-and-ensure-scope.md) — ensure vs replica leaf
- [012](./012-identifiers-and-yerk-uri.md) — id forms
- [013](./013-placement-policy-and-host-state.md) — placement + `replica_method`
- [014](./014-host-local-path-model.md) — paths
- [015](./015-get-and-lookup.md) — read surface for new replicas
- [domain-and-near-term.md](../domain-and-near-term.md)
