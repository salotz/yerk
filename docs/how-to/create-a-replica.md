# How to create a session replica

Status: **stub**

## Goal

Spin out a new on-disk replica for a cataloged project with
`yerk replica create` (worktree from main, or clone from remote).

This is **not** the same as first bootstrap from the remote — use
[materialize](./materialize-a-replica.md) for that.

## Prerequisites

- Project already in the catalog (`domain` + `name`, ADR 012)
- `git` on `PATH`
- For **worktree** method (default): main replica already present
  (`yerk materialize <project-id>` first)

## Steps

1. Optional: confirm paths

   ```sh
   yerk path <project-id>              # workspace
   yerk path <project-id> main         # main hub (worktree source)
   yerk path <project-id> <replica>    # destination
   ```

2. Create with default method (`worktree`):

   ```sh
   yerk replica create <project-id> <replica>
   yerk replica create <project-id>/<replica>
   yerk replica create yerk://domain/name/replica
   ```

3. Or force **clone** method (independent remote checkout; main not required).
   Clones remote default HEAD, then ensures a local branch named like the
   replica (creates from HEAD if missing — the branch need not exist on the
   remote):

   ```sh
   yerk replica create <project-id> <replica> --method clone
   ```

4. Catalog default method (optional on the project row):

   ```toml
   replica_method = "worktree"   # or "clone"
   ```

   CLI `--method` overrides catalog.

5. Destination must not already be a usable checkout (error if present).
   Unlike materialize, create does not treat already-present as success.

6. Verify: `yerk replica get <project-id>/<replica>` or `yerk status <id> <replica>`

## See also

- [How to materialize a replica](./materialize-a-replica.md)
- [Commands reference](../reference/commands.md)
- [ADR 016](../../design/decisions/016-replica-create.md)
