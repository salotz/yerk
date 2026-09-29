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
| `status` | List projects + presence (+ optional change) | `--git`, `--tag` (select by declared catalog tag), `--network` |
| `path` / `resolve` | Print workspace path, or replica path when distinguisher given | `path <proj>` → workspace; `path <proj> <replica>` → checkout |
| `workspace ensure` | Create project workspace dirs only | names required, or `--all`; no replica leaf; no git |
| `clone` | Materialize replica via git | progress names branch/ref |
| `config` | Tool config path / show | samples in `examples/` |
| `catalog` | Catalog path / show (table + declared tags) | samples in `examples/`; ADR 010 |
| `envvars` | Live env values for this process | docs: `yerk help envvars` |
| `version` | Print version identity | |

## See also

- [How to check status](../how-to/check-status.md)
- [How to clone a replica](../how-to/clone-a-replica.md)
- [Catalog reference](./catalog.md)
- [ADR 005](../../design/decisions/005-cli-help-and-envvars.md)
- [ADR 009](../../design/decisions/009-workspace-subcommand-and-ensure-scope.md)
- [ADR 010](../../design/decisions/010-catalog-tag-vocabulary.md)
