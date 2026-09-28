# Explanation: Workspace and replicas

Status: **stub**

## Intent

Clarify that a **workspace** is layout policy plus root, while a **replica**
is a concrete checkout—and why materializing them are separate steps
(`ensure` vs `clone`).

## Points (to expand)

- Workspace root and style decide where paths resolve
- Replica distinguisher (often default branch short name)
- Optional per-project `path` override
- Ensure creates parents; clone creates the git checkout
- Multi-replica / worktree-oriented styles (future detail)

## See also

- [Concepts](./concepts.md)
- [How to clone a replica](../how-to/clone-a-replica.md)
- [Configuration reference](../reference/configuration.md)
