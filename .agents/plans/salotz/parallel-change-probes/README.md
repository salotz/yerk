# Plan index — parallel-change-probes

Ephemeral plan: **parallelize** git change probes in `yerk status` (and any
shared probe path), with deterministic tests via the git adapter seam.

| File | Role |
|------|------|
| [plan.md](./plan.md) | Goals, phases, exit criteria |
| [decisions.md](./decisions.md) | Q&A inbox |
| [checklist.md](./checklist.md) | Execution track |
| [../todo.md](../todo.md) | Owner index |

## Phase order

| # | Phase | Status |
|---|--------|--------|
| 0 | Lock concurrency / ordering decisions | pending |
| 1 | Parallel collector + limits | pending |
| 2 | Fake/deterministic adapter tests | pending |
| 3 | Docs + dogfood | pending |
