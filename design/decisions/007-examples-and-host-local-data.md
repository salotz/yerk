# 007. Examples tree and no host-local data in-repo

## Status

Accepted (2026-09-28)

## Context

Starter config documents lived at the repo root as `*.example.toml`, which
does not scale and collides with “what is product source vs sample input.”

Separately, a `fixtures/operator-host/` tree held **this developer’s real host
paths, remotes, and catalog**. That is unacceptable for a shared project:

- It is not portable across machines or contributors.
- It invites leaking private layout and remote choices into git.
- CI and agents cannot rely on it; “works on my fixture” is not a test.
- It confuses **examples** (portable samples) with **operator state** (XDG /
  local-only).

Unit tests already use temp dirs; that is the right pattern for automation.

## Decision

1. **Portable samples live under `examples/`.**
   - Preferred names: `examples/config.toml`, `examples/catalog.toml`
     (not root-level `*.example.toml`).
   - Content must be **generic**: placeholder paths, example remotes, no
     machine-specific home directories, usernames, or private host layout.
   - Do **not** expose `config example` / `catalog example` (or similar)
     as operational CLI subcommands. Samples stay in `examples/` only;
     a future special subcommand tree could dump them if needed.

2. **Do not commit host-local or machine-specific fixtures.**
   - No `fixtures/` (or similar) that encode one operator’s real catalog,
     workspace roots, or remotes.
   - Operator state belongs under `$XDG_CONFIG_HOME/yerk` (or a path the
     operator sets via `YERK__CONFIG_DIR` / file overrides)—outside this
     repository, or in a **private** local path that is gitignored if ever
     kept next to a checkout.
   - Automated tests use **temp directories** and synthetic catalogs only.

3. **Mise / docs must not depend on host-private trees.**
   - No `mise` tasks that assume `fixtures/operator-host` or similar.
   - How-tos show copying `examples/*.toml` into a temp or XDG dir, or
     pointing `YERK__CONFIG_DIR` at a throwaway directory the reader creates.

4. **Future sample material** (multi-file demos, golden outputs) also goes
   under `examples/` (or `examples/<name>/`) and stays portable. Prefer small,
   documented samples over large snapshots of a real machine.

## Consequences

- Root stays free of sample TOML noise; layout is obvious to newcomers.
- Contributors cannot “check in my host” by accident as the default pattern.
- Agents and plans must not recreate machine-specific fixture trees or tasks.
- Slightly more setup for a first manual try (copy examples → dir), which is
  the correct cost versus shipping private host state.

## Related

- [ADR 004](./004-config-and-catalog-split.md) — config vs catalog files
- [ADR 006](./006-ad-hoc-docs-diataxis.md) — docs location/format
- [AGENTS.md](../../AGENTS.md) — agent constraints (examples + no host fixtures)
