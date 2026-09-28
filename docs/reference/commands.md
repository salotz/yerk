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
| `status` | List projects + presence (+ optional change) | `--git`, `--tag` |
| `path` / `resolve` | Print replica path | |
| `ensure` | Materialize workspace parents | |
| `clone` | Materialize replica via git | |
| `envvars` | Live env values for this process | docs: `yerk help envvars` |
| `register` | Catalog write | Stub / later |
| `pull` / `push` | Sync | Stub / later |

## See also

- [How to check status](../how-to/check-status.md)
- [How to clone a replica](../how-to/clone-a-replica.md)
- [ADR 005](../../design/decisions/005-cli-help-and-envvars.md)
