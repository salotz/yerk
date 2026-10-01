# Reference: Commands

Status: **stub**

## Authority

For flags and per-command help text, prefer:

```sh
yerk --help
yerk <command> --help
```

This page will summarize the stable command set and point at help, not fork
a second full flag dump.

## Command set (near-term)

| Command | Role | Notes |
| --- | --- | --- |
| `status` | Project or replica presence + change | `status [project-id [replica]]`; id forms; `--tag`; change on by default; `--presence-only`; `--network` |
| `path` / `resolve` | Print workspace path, or replica path when distinguisher given | `path <id>` → workspace; `path <id> <replica>` or `id/replica` → checkout |
| `workspace ensure` | Create project workspace dirs only | project ids required, or `--all`; no replica leaf; no git |
| `materialize` | Materialize replica(s) from remote via git | `materialize <id> [replica]` \| `--all` \| `--tag`; optional `--replica`; already-present → ok; no `clone` alias |
| `config` | Tool config path / show | samples in `examples/`; later `config resolve` |
| `catalog` | Catalog path / show (table + declared tags) | samples in `examples/`; ADR 010 |
| `envvars` | Live env values for this process | docs: `yerk help envvars` |
| `version` | Print version identity | |

## See also

- [How to check status](../how-to/check-status.md)
- [How to materialize a replica](../how-to/materialize-a-replica.md)
- [Identifiers](../explanation/identifiers.md)
- [Catalog reference](./catalog.md)
- [ADR 005](../../design/decisions/005-cli-help-and-envvars.md)
- [ADR 009](../../design/decisions/009-workspace-subcommand-and-ensure-scope.md)
- [ADR 010](../../design/decisions/010-catalog-tag-vocabulary.md)
- [ADR 011](../../design/decisions/011-api-resources.md)
- [ADR 012](../../design/decisions/012-identifiers-and-yerk-uri.md)
