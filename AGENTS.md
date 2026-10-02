# Agent guidelines

This project follows the guidelines at https://github.com/salotz/agent-guidelines
(local checkout often `~/tree/personal/devel/agent-guidelines`).

- Read `content/shared/generic-agent-guidelines.md` (or equivalent) for agent-assisted work.
- Prefer task-specific tools over shell.
- Follow salotz RFC 22 (project layout), RFC 23/24 (host context), RFC 28 (PRJX).
- Cache remote resources locally when possible.
- Use compacted inlining for referenced standards.
- Check `content/shared/summaries/` before fetching full external standards.

See agent-guidelines `content/shared/glossary.md` for term definitions.

## This repository

FQ name `salotz.yerk`.
Go CLI shipped as a standalone binary named `yerk`.

Host multi-project manager: register many software projects on one machine,
check out replicas into workspace layouts, status across repos, and (planned)
stage host-local configuration. Speaks PRJX vocabulary.

### Pointers

- [README.md](./README.md): human hub
- [docs/](./docs/): operator/user docs (Diátaxis stubs; ad hoc Markdown, ADR 006)
- [`.appinfo/meta.toml`](./.appinfo/meta.toml): application info + env registry (RFC 030/031)
- [design/goals.md](./design/goals.md): product goals from the idea note
- [design/decisions/](./design/decisions/): ADRs
- [contributing/development.md](./contributing/development.md): bootstrap → build → test
- [`examples/`](./examples/): portable config/catalog samples (ADR 007)
- [`.agents/`](./.agents/): plans and context
- `.prjx-root`, [`.config/`](./.config/): PRJX root + portable metadata
- `mise.toml`, `.editorconfig`, `.bootstrap/`: tooling

### Tooling entrypoints

```sh
sh .bootstrap/host-tool-check
mise install && mise run preload
mise run build
mise run test
mise run check
```

Prefer `mise run` / `mise exec` so pins and `.local` Go caches apply without
shell activation.

### Implementation notes for agents

- Language: Go. Module path: `github.com/salotz/yerk`.
- Binary entrypoint: `cmd/yerk`.
- Keep packages small under `internal/…`.
- Config is TOML under `~/.config/yerk` with `YERK__…` overrides:
  `config.toml` (tool/host) + `catalog.toml` (projects). See ADR 004.
- Env var registry: `internal/envvars` drives CLI help/live dumps (ADR 005).
  Static declarations for discovery: `.appinfo/meta.toml` (RFC 030/031).
  Keep both aligned. Root and command `--help` show **primary** vars;
  docs are `yerk help envvars`; live values are `yerk envvars`. No `--help-all`.
- Do not invent a second product root; code lives at repo root Go layout,
  design/docs beside it.
- Near-term CLI surface (implemented only; no stub commands): `status`,
  `path`/`resolve`, `get`, `lookup`, `project get|lookup`,
  `replica get|lookup|create`, `workspace ensure`, `materialize`, `config`,
  `context` / `context dir`, `state update`, `catalog`, `envvars`, `version`.
  Project args accept bare
  id / short unique name / `yerk://…` (ADR 012; package `internal/id`).
  `domain` is required on every catalog project. `get`/`lookup` return
  project or replica info (placement + paths + presence); `--output json`
  (ADR 015). `config resolve <project-id>` dumps ordered placement
  contributions + effective style (ADR 013); missing state.json is not
  listed under files. `state update` rebinds host project state from
  ambient (or `--workspace-style`); creates or overwrites bindings (unlike
  one-shot bind on ensure/materialize/create).
  `workspace ensure` creates the project workspace directory only (not a
  replica leaf); requires project ids or `--all` (ADR 009). `materialize`
  (renamed from `clone`; no alias) supports single project or bulk
  `--all` / `--tag` (XOR with names). `replica create` spins a session
  replica via `worktree` (default; main must be present) or `clone`
  method (`--method` / catalog `replica_method`; ADR 016); refuse if dest
  present. Catalog root `tags = […]` is a closed vocabulary; project
  `tags` must be members (ADR 010). Catalog edits are hand-edit for now.
  API resources (ADR 011, `internal/api`) carry canonical `uri`. Status
  UX: change on by default (`--presence-only`), origin comparison
  fallback. Further series lives under ephemeral `.agents/plans/<owner>/`
  (do not cite plan-local Q ids in product code). `pull`/`push` only after
  sync semantics ADR.
- **Examples vs host state (ADR 007):** portable samples live under
  `examples/` (e.g. `examples/config.toml`, `examples/catalog.toml`). Do
  **not** put `*.example.toml` at the repo root. Do **not** commit
  machine-specific fixtures (real home paths, remotes, host catalogs) or
  mise tasks that depend on them. Operator config belongs in XDG /
  `YERK__CONFIG_DIR` outside the shared tree (or gitignored local-only).
  Tests use temp dirs only.
- **Host-local agent context (RFC 23 / 26):** do not put operator dogfood
  notes in this remote `.agents/`. On this machine the yerk **project
  workspace** is the parent of this replica (`…/yerk/`, checkout `main/`);
  host placement is optional `[domains]` (default `<root>/<name>`) plus
  optional `[[projects]]` path overrides (ADR 014); catalog has no path.
  Workspace-local agent context lives in `…/yerk/.agents/` (closer than
  in-replica remote context). Application dogfood config remains
  `~/.config/yerk/` and is updated from that workspace-local note when
  schema/env defaults change—never committed here (ADR 007). Project
  workspace is the parent of the replica leaf (e.g. `…/yerk` vs `…/yerk/main`).
- Design spine: `design/domain-and-near-term.md`.
- Design source note (operator org):
  `notes/todo/ideas/20260925T115055--software-project-management-tool__dev_software_todo.org`
