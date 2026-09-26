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
  - `YERK__WORKSPACE_ROOT` / `YERK__WORKSPACE_STYLE` — common workspace knobs
- Missing config file is valid (empty catalog + defaults)
- PRJX discovery stays on PRJX terms (`.prjx-root`, `PRJX__…`)

## Consequences

- Project-local staging paths under `~/.config/yerk/projects/<name>/locals/`
  can be added later without changing the root file convention
- Tests can point `YERK__CONFIG` at temp files
