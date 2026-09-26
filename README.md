# yerk

Host multi-project manager for many software checkouts on one machine.

`yerk` registers projects, checks out **replicas** into a workspace layout,
reports status across repos, and (later) stages host-local configuration into
those checkouts. Vocabulary aligns with [PRJX](https://github.com/salotz/rfcs/tree/master/rfcs/salotz.028_prjx)
(project, replica, workspace).

Similar spirit to tools like `gita`, with extra focus on agentic workflows,
domain layouts, and local config staging.

## Status

Early scaffold. MVP direction:

- catalog of projects (name, remote, tags) in `~/.config/yerk/config.toml`
- CLI: clone / push / pull / status
- two workspace styles:
  - `workspace-dir`: `projects/<project>/<replica>`
  - `project-dir`: `projects/<project>__<replica>`

Working today:

```text
yerk version
yerk status
yerk config path | show | example
```

`register`, `clone`, `pull`, and `push` are stubbed.

## Quick start

Host prerequisites (check only, no installs):

```sh
./.bootstrap/host-tool-check
```

Project tools and build:

```sh
mise install
mise run preload
mise run build
mise run test
```

Run the binary (also on mise `_.path` via `.local/bin`):

```sh
mise run build
.local/bin/yerk version
.local/bin/yerk config example
```

Standalone release-style build (no mise required once Go is available):

```sh
go build -o yerk ./cmd/yerk
```

## Configuration

| Item | Value |
| --- | --- |
| Config dir | `$XDG_CONFIG_HOME/yerk` (default `~/.config/yerk`) |
| Config file | `config.toml` |
| Env file override | `YERK__CONFIG=/path/to/file.toml` |
| Env dir override | `YERK__CONFIG_DIR=/path/to/dir` |
| Workspace root | `YERK__WORKSPACE_ROOT` or `[workspace].root` |
| Workspace style | `YERK__WORKSPACE_STYLE` or `[workspace].style` |

PRJX concerns stay on the PRJX side (`.prjx-root`, `PRJX__…`).
`yerk` discovers them; it does not redefine the spec.

## Layout

- `cmd/yerk` — binary entrypoint
- `internal/cli` — cobra commands
- `internal/config` — XDG + TOML config
- `internal/workspace` — replica path styles
- `internal/version` — build identity
- `design/` — goals and decisions
- `contributing/` — bootstrap and day-to-day tooling
- `.agents/` — agent plans and context

## Name

`yerk` is a carefree handle (proper noun). Optional long form in notes: `yerkum`.
One-liner for agents and READMEs: host multi-project manager; speaks PRJX.
