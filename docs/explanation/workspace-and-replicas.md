# Explanation: Workspace and replicas

Status: **stub**

## Intent

Clarify **project workspace** and **replica**, and why materializing layout vs
git are separate steps (`workspace ensure` vs `materialize`).

## Points (to expand)

- Domain is an **id namespace** only (`personal/yerk`, `yerk://…`) — ADR 012 / 014
- **Catalog** = what projects exist (portable); **config** = how they sit on this host
- Project workspace: optional `[domains]` → `<root>/<name>`; optional host `[[projects]]` path override (absolute/`~/` or relative under root)
- Catalog must not carry host `path` (load error if present)
- Project workspace owns replicas; not a git checkout
- **Effective style** merges layers (built-in → host → dir-local → catalog → **host state** → env → CLI) — ADR 013
- Host state: `$XDG_STATE_HOME/yerk/projects/<domain>/<project>/state.json`
- First successful `workspace ensure` or `materialize` binds style
- Style places the replica (`workspace-dir` / `project-dir`; `name-tags` later)
- `yerk workspace ensure` creates workspace dirs only; `yerk materialize` clones replicas

## See also

- [Configuration](../reference/configuration.md)
- [Catalog](../reference/catalog.md)
- [ADR 013](../../design/decisions/013-placement-policy-and-host-state.md)
- [ADR 014](../../design/decisions/014-host-local-path-model.md)
