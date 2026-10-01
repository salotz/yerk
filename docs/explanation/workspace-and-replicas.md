# Explanation: Workspace and replicas

Status: **stub**

## Intent

Clarify **domain root**, **project workspace**, and **replica**, and why
materializing layout vs git are separate steps
(`workspace ensure` vs `materialize`).

## Points (to expand)

- Domain root (`[domains.personal]`) is host-absolute; catalog-relative paths join it
- Project workspace owns replicas; not a git checkout
- Style places the replica under the workspace (`workspace-dir` / `project-dir`)
- Replica distinguisher (often default branch short name)
- Absolute catalog `path` is a host escape hatch
- `yerk workspace ensure <project-id>…` / `--all` creates workspace dirs only (not `…/main`)
- `yerk materialize` creates the git checkout under that layout (product name; git still clones)
- Multi-replica / worktree-oriented styles (future detail)

## See also

- [Concepts](./concepts.md)
- [Identifiers](./identifiers.md)
- [How to materialize a replica](../how-to/materialize-a-replica.md)
- [Configuration reference](../reference/configuration.md)
- [ADR 008](../../design/decisions/008-domain-roots-and-relative-catalog-paths.md)
- [ADR 009](../../design/decisions/009-workspace-subcommand-and-ensure-scope.md)
- [ADR 012](../../design/decisions/012-identifiers-and-yerk-uri.md)
