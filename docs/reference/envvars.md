# Reference: Environment variables

Status: **stub** (table filled from the static registry; full prose later)

## Authority

| Layer | Role |
| --- | --- |
| [`.appinfo/meta.toml`](../../.appinfo/meta.toml) | Static declarations (RFC 030/031); discoverable without running the binary |
| `internal/envvars` | In-binary registry for CLI help and live dumps (ADR 005) |
| `yerk help envvars` | Full documentation view from the binary |
| `yerk envvars` | Live process values |
| `yerk --help` / `yerk <cmd> --help` | Primary subset only |

Keep `.appinfo` and `internal/envvars` aligned when adding tool-owned knobs
(same **type**, **policy**, **default**, and enum **values**).
CLI help formatters print those fields so operators see the same facts as the
static file. Do not fork a third hand-written dump.

Prefix: `YERK` → names `YERK__*` ([RFC 027](https://github.com/salotz/rfcs/tree/master/rfcs/salotz.027_env-nexps)).
Value types and policies: [RFC 032](https://github.com/salotz/rfcs/tree/master/rfcs/salotz.032_env-value-types).
`policy=required` means must be set (no default). Today all yerk knobs use
`policy=warn` with an explicit default story.

## Tool-owned (`YERK__*`)

| Name | Type | Policy | Default | Description |
| --- | --- | --- | --- | --- |
| `YERK__CONFIG_DIR` | string | warn | `${XDG_CONFIG_HOME}/yerk` (else `~/.config/yerk`) | Directory containing `config.toml` and `catalog.toml` by default |
| `YERK__CONFIG` | string | warn | `${YERK__CONFIG_DIR}/config.toml` | Absolute path to tool config |
| `YERK__CATALOG` | string | warn | `${YERK__CONFIG_DIR}/catalog.toml` | Absolute path to project catalog |
| `YERK__WORKSPACE_ROOT` | string | warn | config `[workspace].root` | Override checkout tree root |
| `YERK__WORKSPACE_STYLE` | enum | warn | config style; else `workspace-dir` | `workspace-dir` \| `project-dir` |

## Platform (external)

| Name | Type | Policy | Default | Description |
| --- | --- | --- | --- | --- |
| `XDG_CONFIG_HOME` | string | warn | platform / XDG default | Base for user config when `YERK__CONFIG_DIR` unset |
| `HOME` | string | warn | platform home | Fallback for user config dir resolution |
| `PATH` | string | warn | process `PATH` | Must include `git` for clone / `--git` / ls-remote |

## PRJX (not in `.appinfo`)

`PRJX_*` / `PRJX__*` are owned by PRJX (RFC 28), not by the yerk product
prefix. They are **not** declared in `.appinfo/meta.toml`.

- Project identity / optional project-local leaves:
  [`.config/_project-meta.toml`](../../.config/_project-meta.toml)
- Root sentinel: [`.prjx-root`](../../.prjx-root)

The in-binary help registry (`internal/envvars`, `yerk help envvars`) may
still *mention* some PRJX names as future discovery surface. That is CLI
documentation, not an application-info claim that yerk defines them.

## See also

- [Configuration reference](./configuration.md)
- [`.appinfo/meta.toml`](../../.appinfo/meta.toml)
- [ADR 003](../../design/decisions/003-config-xdg-and-env.md), [ADR 005](../../design/decisions/005-cli-help-and-envvars.md)
