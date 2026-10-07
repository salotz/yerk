# Plan index — status-filter-list-color

Ephemeral plan: **status UX** — convenience filters (`--dirty`, `--present`,
…), later `--status=` expressions (same engine as tags), a **cross-project
replica listing**, and **color** on human status tables.

Sparked by operator request (2026-10-07): show only dirty (etc.) rows; list
every replica across projects; color by state.

| File | Role |
|------|------|
| [plan.md](./plan.md) | Goals, phases, exit criteria, design sketch |
| [decisions.md](./decisions.md) | **Only** Q&A inbox (`Q*`) |
| [checklist.md](./checklist.md) | Execution track |
| [../todo.md](../todo.md) | Owner index |

## Phase order

| # | Phase | Status |
|---|--------|--------|
| 0 | Scope + open questions | pending |
| 1 | ADR (filters, replica grain, color) | pending |
| 2 | Convenience filters | pending |
| 3 | Cross-project replica listing | pending |
| 4 | Color on human status | pending |
| 5 | Docs + close | pending |

`--status=` expression language is **deferred** until
[tag-boolean-select](../tag-boolean-select/) has an accepted engine; this plan
only reserves the flag shape and atom vocabulary.

## Protocol

- No git stage/commit unless asked.
- Never cite `Q*` in product code.
- No unimplemented cobra flags: ship behavior with each flag.
