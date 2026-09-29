# How to clone a default replica

Status: **stub**

## Goal

Materialize an on-disk replica for a cataloged project with `yerk clone`
(after the project workspace exists if needed).

## Prerequisites

- Project already in the catalog
- `git` on `PATH`
- Domain roots configured for relative catalog paths (ADR 008)

## Steps (to write)

1. Confirm resolve path: `yerk path <project> <replica>` (workspace: `yerk path <project>`)
2. Optional: `yerk workspace ensure <project>` (or `--all`) for workspace dirs
3. `yerk clone <project> [replica]`
4. Verify with `yerk status` / `yerk status --git`

## See also

- [Commands reference](../reference/commands.md)
- [Workspace and replicas (explanation)](../explanation/workspace-and-replicas.md)
