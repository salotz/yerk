# Reference: Configuration

Status: **stub**

## Layout

| Item | Default / notes |
| --- | --- |
| Config dir | `$XDG_CONFIG_HOME/yerk` (usually `~/.config/yerk`) |
| Tool config | `config.toml` — workspace style, domain roots, future tool knobs |
| Catalog | `catalog.toml` — project registry (see [catalog](./catalog.md)) |
| Application info | [`.appinfo/meta.toml`](../../.appinfo/meta.toml) — products + env registry (RFC 030/031); not runtime config |

## Overrides

| Variable | Effect |
| --- | --- |
| `YERK__CONFIG_DIR` | Directory containing both default files |
| `YERK__CONFIG` | Explicit tool config file path |
| `YERK__CATALOG` | Explicit catalog file path |
| `YERK__WORKSPACE_STYLE` | Workspace style override (`workspace-dir` \| `project-dir`) |

Full env documentation: `yerk help envvars`. Live values: `yerk envvars`.

## `config.toml`

- `[workspace].style` — how replicas sit under each resolved project workspace:
  - `workspace-dir`: `<workspace>/<replica>`
  - `project-dir`: `<dir(workspace)>/<name>__<replica>`
- `[domains]` — map domain name → host-absolute root (leading `~/` ok).
  Relative catalog paths join `<domains[domain]>/<path>` ([ADR 008](../../design/decisions/008-domain-roots-and-relative-catalog-paths.md)).
- Defaults when the file is missing (`style = workspace-dir`, no domains)

## Examples

Portable starters: [`examples/config.toml`](../../examples/config.toml),
[`examples/catalog.toml`](../../examples/catalog.toml). See [ADR 007](../../design/decisions/007-examples-and-host-local-data.md).
Edit domain roots for **your** host; keep catalog paths relative when possible.

## See also

- [Catalog reference](./catalog.md)
- [Environment variables](./envvars.md)
- [How to use example config files](../how-to/use-example-config.md)
- [ADR 003](../../design/decisions/003-config-xdg-and-env.md), [ADR 004](../../design/decisions/004-config-and-catalog-split.md), [ADR 007](../../design/decisions/007-examples-and-host-local-data.md), [ADR 008](../../design/decisions/008-domain-roots-and-relative-catalog-paths.md)
