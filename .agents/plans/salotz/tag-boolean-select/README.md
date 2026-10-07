# Plan index — tag-boolean-select

Ephemeral plan: **design only** for boolean / expression selection over catalog
tags (e.g. `science AND NOT deprecated`). **No CLI stubs** until an ADR is
accepted and a later implement plan starts.

Sparked by operator idea note (2026-10-03): avoid ad-hoc tag logic; consider
CEL vs lighter alternatives.

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
| 1 | ADR draft + review | pending |
| 2 | Docs pointers; implement plan stub or explicit defer | pending |

## Protocol

- Design/docs only unless operator opens an implement plan.
- No git stage/commit unless asked.
- Never cite `Q*` in product code.
