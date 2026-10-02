# 003. XDG config and `YERK__` env

## Status

Accepted (2026-09-25)

## Context

The tool needs a host-global catalog (many projects) separate from any one
repo checkout. PRJX already defines project-local and domain-local config;
`yerk` must not conflate those with its own operator catalog.

## Decision

- Primary config file: `$XDG_CONFIG_HOME/yerk/config.toml`
- Format: TOML
- Overrides:
  - `YERK__CONFIG` — explicit file path
  - `YERK__CONFIG_DIR` — alternate config directory
  - `YERK__WORKSPACE_STYLE` — optional style overlay (`workspace-dir` |
    `project-dir`); applied under each resolved project workspace path
  - Optional domain roots live in `config.toml` `[domains]` (host-absolute; see
    [ADR 014](./014-host-local-path-model.md)); not a single
    shared `[workspace].root`
- Missing config file is valid (empty catalog + defaults)
- PRJX discovery stays on PRJX terms (`.prjx-root`, `PRJX__…`)

## Consequences

- Project-local staging paths under `~/.config/yerk/projects/<name>/locals/`
  can be added later without changing the root file convention
- Tests can point `YERK__CONFIG` at temp files
- Environment variable **names and metadata** live in `internal/envvars` as
  the single source of truth. **Which** help surface shows a full list vs a
  primary summary is decided in [ADR 005](./005-cli-help-and-envvars.md)
  (cobra-aware: root and command `--help` are selective; full reference is
  `yerk help envvars` / `yerk envvars`). Early drafts of this ADR required a
  comprehensive dump on every `--help`; that is superseded by ADR 005.
- Do not register `YERK__*` under PRJX `[project.env-vars]` (wrong prefix).
  That section declares **project-local** `PRJX__*` leaves only (RFC 28).
  Document tool vars in the envvars package + CLI help; describe the split in
  `.config/_project-meta.toml` comments.
