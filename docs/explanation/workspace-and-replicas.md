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
- First successful `workspace ensure`, `materialize`, or `replica create` binds style
- Style places the replica (`workspace-dir` / `project-dir`; `name-tags` later)
- `yerk workspace ensure` creates workspace dirs only
- `yerk materialize` clones replicas from the **remote** (bootstrap; already-present ok)
- `yerk replica create` spins a session replica: **worktree** from main (default) or **clone** method (ADR 016); refuse if dest present; worktree hard-errors if main missing

### Styles (path math)

| Style | Main / default replica | Other replicas |
| --- | --- | --- |
| `workspace-dir` | `<workspace>/<replica>` | same pattern |
| `project-dir` | `<parent>/<name>__<replica>` (siblings of workspace) | same |
| `name-tags` | `<workspace>/<name>` (bare project dir) | `<workspace>/<name>__R` |

`name-tags` optional params (catalog or host `[[projects]]` inline table):

```toml
workspace_style = { style = "name-tags", main_dir = "~/.bimker", replica_dir = "~/tree/…/bimker" }
```

See [ADR 018](../../design/decisions/018-name-tags-workspace-style.md).

## See also

- [Configuration](../reference/configuration.md)
- [Catalog](../reference/catalog.md)
- [ADR 013](../../design/decisions/013-placement-policy-and-host-state.md)
- [ADR 014](../../design/decisions/014-host-local-path-model.md)
