# Reference: Configuration

Status: **stub**

## Layout

| Item | Default / notes |
| --- | --- |
| Config dir | `$XDG_CONFIG_HOME/yerk` (usually `~/.config/yerk`) |
| Tool config | `config.toml` — ambient style, optional **`[domains]`**, optional **`[[projects]]`** |
| Catalog | `catalog.toml` — portable project registry (see [catalog](./catalog.md)) |
| State dir | `$XDG_STATE_HOME/yerk` — project bindings |
| Application info | [`.appinfo/meta.toml`](../../.appinfo/meta.toml) |

## Overrides

| Variable | Effect |
| --- | --- |
| `YERK__CONFIG_DIR` | Directory containing both default config files |
| `YERK__CONFIG` | Explicit tool config file path |
| `YERK__CATALOG` | Explicit catalog file path |
| `YERK__STATE_DIR` | Explicit state root |
| `YERK__WORKSPACE_STYLE` | Ambient workspace style overlay |

## `config.toml`

- `[workspace].style` — ambient host default (`workspace-dir` \| `project-dir`)
- Effective style is **merged** (ADR 013)
- **`[domains]`** — optional map domain name → host-absolute root (`~/…` ok)
- **`[[projects]]`** — optional host rows matched by `name` + `domain`

```toml
[workspace]
style = "workspace-dir"

[domains]
personal = "~/tree/personal/devel"

# Only exceptions need rows:
[[projects]]
name = "bimker"
domain = "personal"
path = "~/.bimker"
# workspace_style = "project-dir"
```

### Workspace path resolution (ADR 014)

| Condition | Workspace |
| --- | --- |
| Host row `path` absolute or `~/…` | that path |
| Host row `path` relative | `<domains[domain]>/<path>` |
| No path (no row or empty) | `<domains[domain]>/<name>` |
| Otherwise | error — set domain root or full path |

Most catalog projects need **no** host row when a domain root is set.

| Field | Required | Meaning |
| --- | --- | --- |
| `[domains].*` | no | Default base for projects in that domain |
| `[[projects]].name` | yes (if row present) | Short name (with domain → bare id) |
| `[[projects]].domain` | yes (if row present) | Id namespace |
| `[[projects]].path` | no | Override or relative under domain root |
| `[[projects]].workspace_style` | no | Host-row ambient style override |

## Host project state

```text
$XDG_STATE_HOME/yerk/projects/<domain>/<project>/state.json
```

## See also

- [Catalog reference](./catalog.md)
- [ADR 013](../../design/decisions/013-placement-policy-and-host-state.md), [ADR 014](../../design/decisions/014-host-local-path-model.md)
