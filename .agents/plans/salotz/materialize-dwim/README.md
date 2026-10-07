# Plan index — materialize-dwim

Ephemeral plan: make **`yerk materialize`** the declarative / DWIM path for
putting a replica on disk (worktree when main is present; clone only as the
exception). Sparked by operator incident: `yerk materialize examol/darpa-nodes
md-next-workflow` independently cloned an LFS remote instead of worktreeing
from an existing (not yerk-seeded) `main`.

| File | Role |
|------|------|
| [plan.md](./plan.md) | Goals, phases, exit criteria, incident recap |
| [decisions.md](./decisions.md) | **Only** Q&A inbox (`Q*`) |
| [checklist.md](./checklist.md) | Execution track |
| [../todo.md](../todo.md) | Owner index |

## Phase order

| # | Phase | Status |
|---|--------|--------|
| 0 | Lock remaining policy (Q1–Q8) | pending |
| 1 | ADR 025 (amends 016) + operator accept | pending |
| 2 | Implement DWIM materialize + tests | pending |
| 3 | Help / how-tos / domain spine | pending |

## Protocol

- Do not implement until Q1–Q8 are locked enough to draft ADR 025.
- No git stage/commit unless asked.
- Never cite `Q*` in product code.
- Close-out: delete this folder; update owner `todo.md`.
