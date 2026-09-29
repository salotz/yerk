# How to clone a default replica

Status: **stub**

## Goal

Materialize on-disk replica(s) for cataloged project(s) with `yerk clone`
(after the project workspace exists if needed).

## Prerequisites

- Project already in the catalog
- `git` on `PATH`
- Domain roots configured for relative catalog paths (ADR 008)

## Steps

1. Confirm resolve path: `yerk path <project> <replica>` (workspace: `yerk path <project>`)
2. Optional: `yerk workspace ensure <project>` (or `--all`) for workspace dirs
3. Clone one project (default replica, or name it):

   ```sh
   yerk clone <project>
   yerk clone <project> main
   yerk clone <project> --replica main
   ```

4. **Bulk** — pick exactly one selector (not combined with project args):

   ```sh
   yerk clone --tag devel
   yerk clone --all
   ```

   - `--tag` must be in the catalog root `tags` list
   - Empty tag match or empty catalog with `--all` → error
   - Optional `--replica <name>` applies the same distinguisher to every
     selected project; otherwise each project uses its own default

5. **Idempotent present path:** if the replica path is already a usable git
   checkout, clone reports `already present` on stderr, prints the path, and
   exits successfully (no re-clone). A path that exists but is **not** a
   checkout remains an error.

6. Verify with `yerk status` / `yerk status <project> <replica>`

## See also

- [Commands reference](../reference/commands.md)
- [Workspace and replicas (explanation)](../explanation/workspace-and-replicas.md)
