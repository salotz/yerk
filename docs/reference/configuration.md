# Reference: Configuration

Status: **stub**

## Layout

| Item | Default / notes |
| --- | --- |
| Config dir | `$XDG_CONFIG_HOME/yerk` (usually `~/.config/yerk`) |
| Tool config | `config.toml` — workspace root/style, future tool knobs |
| Catalog | `catalog.toml` — project registry (see [catalog](./catalog.md)) |
| Application info | [`.appinfo/meta.toml`](../../.appinfo/meta.toml) — products + env registry (RFC 030/031); not runtime config |

## Overrides

| Variable | Effect |
| --- | --- |
| `YERK__CONFIG_DIR` | Directory containing both default files |
| `YERK__CONFIG` | Explicit tool config file path |
| `YERK__CATALOG` | Explicit catalog file path |
| `YERK__WORKSPACE_ROOT` | Workspace root override |
| `YERK__WORKSPACE_STYLE` | Workspace style override |

Full env documentation: `yerk help envvars`. Live values: `yerk envvars`.

## `config.toml` (to flesh out)

- `[workspace].root`
- `[workspace].style`
- Defaults when the file is missing

## See also

- [Catalog reference](./catalog.md)
- [Environment variables](./envvars.md)
- [ADR 003](../../design/decisions/003-config-xdg-and-env.md), [ADR 004](../../design/decisions/004-config-and-catalog-split.md)
