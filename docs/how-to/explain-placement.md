# How to explain placement (“why is my style X?”)

Status: **draft**

## Command

```sh
yerk config resolve <project-id>
yerk config resolve personal/yerk --output json
```

Shows the **ordered contribution stack** for placement-related keys (v1:
workspace style) and the **effective** snapshot for one catalog project.

## How to read the dump

1. **files** — config/catalog/dir-local paths considered, plus **state.json only
   when a binding exists** (missing state is a contribution note, not a fake path).
2. **contributions** — layers from low → high precedence among ambient sources;
   bound **state** overrides ambient (with a warning if they disagree); explicit
   CLI would appear last (not set on a plain resolve).
3. **effectiveStyle** / **bound** — what path math actually uses.

Typical layers:

| Layer | Source |
| --- | --- |
| `built-in` | default `workspace-dir` |
| `host-config` | `$XDG_CONFIG_HOME/yerk/config.toml` `[workspace].style` |
| `dir-local` | `<dir>/.local/yerk/config.toml` walking up from the project workspace |
| `catalog` | optional `workspace_style` on the catalog row |
| `env` | `YERK__WORKSPACE_STYLE` |
| `host-project` | optional host `[[projects]]` `workspace_style` |
| `state` | `$XDG_STATE_HOME/yerk/projects/<domain>/<name>/state.json` |
| `cli` | `--workspace-style` (mutate commands; usually unset here) |

If ambient changed and you want the binding to match, see
[Refresh host project state](./update-project-state.md) (`yerk state update`).

## See also

- [Configuration reference](../reference/configuration.md)
- [ADR 013](../../design/decisions/013-placement-policy-and-host-state.md)
- [Get project info](./get-project-info.md)
- [Refresh host project state](./update-project-state.md)
