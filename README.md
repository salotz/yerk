# yerk

Project manager for many projects on your host.

`yerk` is meant to be used both by humans, agents, and harnesses for better management of project-based agent sessions.

`yerk` keeps a **catalog** of projects, checks out **replicas** into a
**workspace** layout, resolves paths by name, reports **presence** and
**change** status, and allows for bulk actions on projects.

`yerk` supports `git` primarily as a first class mechanism for project version controlling,
but is not limited to `git`.
Practically when applied to git repos a **replica** encapsulates both clones and worktrees.

`yerk` supports different styles of **workspaces** for projects (i.e. the directories your cloned repos and worktrees live in), with a focus on clear patterns and standards for working on multiple changes simultaneously for humans and agents (e.g. git worktrees).

`yerk` is designed to work across polyrepos and within monorepos, or any combination of them.
In `yerk` you can declare groupings of repos to be managed together for polyrepos and you can work with sub-projects within monorepos as well.

`yerk` is written in Go to provide portable single file binaries.


## Declarations

Static application info and environment registry
([RFC 030](https://github.com/salotz/rfcs/tree/master/rfcs/salotz.030_application-info),
[RFC 031](https://github.com/salotz/rfcs/tree/master/rfcs/salotz.031_application-env)):

| Path | Role |
| --- | --- |
| [`.appinfo/meta.toml`](./.appinfo/meta.toml) | Product metadata + env var declarations (`YERK__*`, plus platform names yerk actually reads) |
| [`.config/_project-meta.toml`](./.config/_project-meta.toml) | PRJX project id (`salotz.yerk`); not tool env |
| [`.prjx-root`](./.prjx-root) | PRJX project-root sentinel |

CLI help still uses the in-binary registry (`internal/envvars`, ADR 005):
`yerk help envvars` / `yerk envvars`. Keep `.appinfo` aligned when adding knobs.

## Quick start

Build, then point at a throwaway config dir seeded from portable examples
([ADR 007](./design/decisions/007-examples-and-host-local-data.md)):

```sh
mise run build

mkdir -p /tmp/yerk-dev
cp examples/config.toml /tmp/yerk-dev/config.toml
cp examples/catalog.toml /tmp/yerk-dev/catalog.toml
# edit [domains] roots in config.toml (default <root>/<name>); catalog has no path (ADR 014)

export YERK__CONFIG_DIR=/tmp/yerk-dev
.local/bin/yerk catalog show
.local/bin/yerk status
```

Or install into real XDG:

```sh
mkdir -p ~/.config/yerk
cp examples/config.toml ~/.config/yerk/config.toml
cp examples/catalog.toml ~/.config/yerk/catalog.toml
# edit [domains] for this host; optional [[projects]] path only for exceptions (ADR 014)
yerk status
```

Samples are files under `examples/` only (no `config example` / `catalog example` CLI).

## Configuration

### Files and paths

| Item | Value |
| --- | --- |
| Config dir | `$XDG_CONFIG_HOME/yerk` (default `~/.config/yerk`) |
| Tool config | `config.toml` — style, optional `[domains]`, optional host `[[projects]]` (ADR 014) |
| Catalog | `catalog.toml` — portable `[[projects]]` registry (no host paths) |
| Application info | [`.appinfo/meta.toml`](./.appinfo/meta.toml) — products + env registry (RFC 030/031) |

### Environment variables

Declared in `.appinfo/meta.toml` (`[env]`, prefix `YERK`).
Primary tool knobs:

| Name | Role |
| --- | --- |
| `YERK__CONFIG_DIR` | Directory for both default config files |
| `YERK__CONFIG` | Explicit tool `config.toml` path |
| `YERK__CATALOG` | Explicit `catalog.toml` path |
| `YERK__WORKSPACE_STYLE` | Override `[workspace].style` (`workspace-dir` or `project-dir`) |

| View | Where |
| --- | --- |
| Static registry (source tree) | [`.appinfo/meta.toml`](./.appinfo/meta.toml) |
| Full CLI documentation | `yerk help envvars` |
| Live process values | `yerk envvars` |
| Primary vars on `--help` | root and command help (ADR 005) |

Also declared as external (not owned by yerk): `XDG_CONFIG_HOME`, `HOME`,
`PATH`. PRJX names (`PRJX_*` / `PRJX__*`) stay on the PRJX side
(`.prjx-root`, `.config/_project-meta.toml`) and are not listed in `.appinfo`.
Command `--help` shows primary vars only; see ADR 005.

### Catalog fields (MVP)

```toml
tags = ["devel"]             # closed vocabulary; project tags must be members (ADR 010)

[[projects]]
name = "yerk"
domain = "personal"          # namespace + domain-root key when path is relative
remote = "git@github.com:salotz/yerk.git"
path = "devel/yerk"          # project workspace relative to domain (not …/yerk/main)
tags = ["devel"]             # optional; each entry ∈ catalog tags
default_replica = "main"     # optional; else remote HEAD / fallback main
```

## Documentation

Operator and user docs live under [`docs/`](./docs/) as **ad hoc plain Markdown**
organized by [Diátaxis](https://diataxis.fr/)
(tutorials, how-to, reference, explanation).
A real publication system is deferred
([ADR 006](./design/decisions/006-ad-hoc-docs-diataxis.md)).

- Hub: [`docs/README.md`](./docs/README.md)
- Design (goals, domain, ADRs): [`design/`](./design/)
- Contributor tooling: [`contributing/`](./contributing/)

## Layout

- `cmd/yerk` — binary entrypoint
- `internal/cli` — cobra commands
- `internal/config` — XDG tool config + catalog
- `internal/workspace` — domain roots + relative path resolve + style + workspace ensure
- `internal/presence` — missing/present/invalid
- `internal/gitcmd` — git subprocess adapter
- `internal/project` — resolve + status join
- `internal/version` — build identity
- `examples/` — portable config/catalog samples (ADR 007; not host-local state)
- `.appinfo/` — application info + env registry (RFC 030/031)
- `docs/` — operator/user docs (Diátaxis; plain Markdown, ADR 006)
- `design/` — goals, domain language, ADRs
- `contributing/` — bootstrap and day-to-day tooling
- `.agents/` — agent plans and context

## Vocabulary


Vocabulary aligns with
[PRJX](https://github.com/salotz/rfcs/tree/master/rfcs/salotz.028_prjx)
(project, replica, workspace).
