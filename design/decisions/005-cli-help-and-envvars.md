# 005. CLI help layout and environment variable docs (cobra)

## Status

Accepted (2026-09-26); amended same day for live `yerk envvars` values.

## Context

Environment variable documentation started as a **comprehensive dump** on
`yerk --help` and every `yerk <cmd> --help`, generated from
`internal/envvars` (see ADR 003 early consequence).

That is accurate but hard to scan. Operators want:

1. A **selected summary** of the most relevant tool vars on main help
2. A **full reference** somewhere discoverable, split into **global** vs
   **command-scoped**, with command-scoped entries listing which commands
   apply
3. Each subcommand `--help` showing the **most important** vars for that
   command, not the entire product surface
4. A **runtime view** of what the process actually sees for those knobs
   (application-scoped `env`), separate from documentation

The CLI is built on [spf13/cobra](https://github.com/spf13/cobra). Help UX
must stay idiomatic for that library so we do not re-implement help.

Related product direction (operator): declare `YERK__*` public env vars in a
registry; help text and live dumps are *views* of that registry, not second
hand-written sources.

## Decision

### Single registry, multiple views

- Keep **`internal/envvars` as the single source of truth** for CLI help and
  live dumps: names, summaries, **type / policy / default / enum values**
  (RFC 031/032), scope, global vs command attachment, and which entries are
  “primary” for short help.
- Static discovery uses **`.appinfo/meta.toml`** (RFC 030/031). Tool-owned
  and platform rows there **must match** the registry’s type, policy,
  default story, and enum values. Help formatters relay those fields so
  operators see the same facts as the static file.
- Do **not** scatter env prose into each command’s `Long` by hand.
- Command packages only choose *which formatter* to attach (root summary vs
  command primary vs full reference vs live values).
- **PRJX** names are not product env: full help uses a short pointer to
  RFC 28 / project metadata, not a dump of every `PRJX_*` key. They stay
  out of `.appinfo` env tables.

### Surfaces

| Surface | Content |
| --- | --- |
| `yerk --help` / `yerk help` | Product blurb + **primary** tool env summary + pointers |
| `yerk help envvars` | **Full documentation** reference: global, then command-scoped (with command lists), then platform / PRJX groups |
| `yerk envvars` | **Live values**: table of registry names and current process values (app-scoped `env`) |
| `yerk <cmd> --help` | Command prose + **primary** env vars for that command + pointers |

Docs and live values are deliberately different commands so operators are not
surprised by a wall of prose when they want `NAME` / value, and not surprised
by missing semantics when they want the reference.

### Live values behavior (`yerk envvars`)

- Rows come from the registry’s **concrete** names (skip pattern rows like
  `PRJX__*`).
- Value is `os.Getenv(name)` for this process; empty string displays as
  `(unset)` (distinct from a set-but-empty value only if we later need that;
  today empty getenv is shown as unset).
- Also list any process keys matching `YERK__` or `PRJX__` prefixes that are
  **not** already in the registry (forward-compatible / stray knobs).
- Do not resolve effective config paths here (that is `yerk config path` /
  `catalog path`); this command is the raw environment view.
- Optional later: `--unset-only`, machine-readable formats, filter by scope.

### Explicit non-goals (cobra-aware)

- **No `--help-all` flag.** Cobra owns `-h` / `--help` via `HelpFunc` and
  exits before `Run`. Prefer `yerk help envvars` for depth.
- **Do not** merge `YERK__*` into pflag’s flag help block.
- Root help is **not** required to list every PRJX or platform name.

### Cobra mechanics

- Register an **`envvars` command** with:
  - `Long` = full documentation reference (so `yerk help envvars` and
    `yerk envvars --help` show docs)
  - `Run` = print live values table
- Keep using cobra’s normal `--help` / `help` for other commands; only change
  generated `Long` content from the registry.

### Primary vs full

- **Primary** (`HelpPrimary`): root summary and matching command `--help`.
- **Full** (docs): complete product documentation list on `help envvars`.
- **Live**: concrete registry names + extra matching process keys.

Today all tool-owned `YERK__*` names are primary. Non-primary examples:
`HOME`, most `PRJX_*` names until wired. `PATH` may be primary on commands
that invoke git.

## Consequences

- ADR 003’s earlier “comprehensive on every `--help`” consequence is
  **superseded**; ADR 003 still owns XDG paths and `YERK__` naming.
- Operators learn semantics via `yerk help envvars`; inspect the running
  environment via `yerk envvars`.
- Adding a new env var: registry entry (type/policy/default) + primary bit +
  command list + matching `.appinfo/meta.toml` row; help and live dump update
  together.
- Tests assert summary vs full docs vs live dump separately, including
  type/policy/default lines on help surfaces.
- Future public `YERK__` declaration standards plug into the same registry.

## Alternatives considered

### Same text for `yerk envvars` and `yerk help envvars`

Rejected after operator feedback: run path should be live values like shell
`env`, not the documentation page.

### `--help-all` on every command

Rejected for MVP: fights cobra’s help path.

### Comprehensive root help only

Rejected: poor scanability.

### Hand-written Long strings per command

Rejected: drifts; contradicts single-registry approach.
