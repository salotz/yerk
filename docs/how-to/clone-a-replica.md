# How to clone a default replica

Status: **stub**

## Goal

Materialize an on-disk replica for a cataloged project with `yerk clone`
(and ensure workspace parents if needed).

## Prerequisites

- Project already in the catalog
- `git` on `PATH`
- Workspace root configured

## Steps (to write)

1. Confirm resolve path: `yerk path <project>`
2. Optional: `yerk ensure <project>` for layout parents
3. `yerk clone <project> [replica]`
4. Verify with `yerk status` / `yerk status --git`

## See also

- [Commands reference](../reference/commands.md)
- [Workspace and replicas (explanation)](../explanation/workspace-and-replicas.md)
