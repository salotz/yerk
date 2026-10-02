# How to materialize a default replica

Status: **stub**

## Goal

Materialize on-disk replica(s) for cataloged project(s) with `yerk materialize`
(after the project workspace exists if needed).

## Prerequisites

- Project already in the catalog with required `domain` + `name` (ADR 012)
- `git` on `PATH`
- Domain roots configured for relative catalog paths (ADR 008; placement
  successor may change path join)

## Steps

1. Confirm resolve path: `yerk path <project-id> <replica>` (workspace:
   `yerk path <project-id>`)
2. Optional: `yerk workspace ensure <project-id>` (or `--all` / `--tag`) for workspace dirs
3. Materialize one project (default replica, or name it):

   ```sh
   yerk materialize <project-id>
   yerk materialize <project-id> main
   yerk materialize <project-id> --replica main
   yerk materialize personal/yerk/main
   yerk materialize yerk://personal/yerk/main
   ```

   Project id forms: unique short name, `domain/name`, or `yerk://…` URI.

4. **Bulk** — pick exactly one selector (not combined with project args):

   ```sh
   yerk materialize --tag devel
   yerk materialize --all
   ```

   - `--tag` must be in the catalog root `tags` list
   - Empty tag match or empty catalog with `--all` → error
   - Optional `--replica <name>` applies the same distinguisher to every
     selected project; otherwise each project uses its own default

5. **Idempotent present path:** if the replica path is already a usable git
   checkout, materialize reports `already present` on stderr, prints the path,
   and exits successfully (no re-clone). A path that exists but is **not** a
   checkout remains an error.

6. Verify with `yerk status` / `yerk status <project-id> <replica>`

## See also

- [How to create a session replica](./create-a-replica.md) (`replica create`)
- [Commands reference](../reference/commands.md)
- [Identifiers (explanation)](../explanation/identifiers.md)
- [Workspace and replicas (explanation)](../explanation/workspace-and-replicas.md)
