# Plan index — initial-feature-series

Ephemeral multi-session plan for the next product feature series after the
first vertical slice on `main`.

| File | Role |
|------|------|
| [plan.md](./plan.md) | Goals, phases, architecture, exit criteria |
| [decisions.md](./decisions.md) | **Only** Q&A inbox (`Q*`); Phase 0 complete |
| [checklist.md](./checklist.md) | Execution track (boxes); not answers |
| [../todo.md](../todo.md) | Owner In Progress / Backlog index |

Work-process: agent-guidelines personal `work-process.md` (plan folder layout,
no staging unless asked, answers only in `decisions.md`).

## Phase order (execute)

| # | Phase | Status |
|---|--------|--------|
| 0 | Lock decisions | **done** (Q1–Q33 locked) |
| 1 | Identifiers / `yerk://` (P13) | **done** (ADR 012 + `internal/id` + CLI) |
| 2 | Placement policy + host state (P10) | **done** (ADR 013/014; optional `[domains]` default; placement/state + wire) |
| 3 | Project/replica get + lookup (P14) | **done** (ADR 015; get/lookup + ProjectInfo/ReplicaInfo; `--output json`) |
| 4 | `config resolve` stack (P15) | **done** (`yerk config resolve`; placement.Explain + ConfigResolve) |
| 5 | `materialize` + `replica create` (P11) | **next** (partial: materialize rename done; create pending) |
| 6 | Agent context dumps (P16) | pending |
| 7 | New workspace styles (P12) | pending |
| 8 | Polish (ensure `--tag`, output formats) | pending |
| 9 | Sync verbs design only | pending |
| — | Parallel probes | **held** (owner backlog) |

## Protocol

- One step at a time on “go” unless operator widens scope.
- No git stage/commit unless asked.
- Promote durable choices to `design/decisions/` ADRs; never cite `Q*` in product code.
- Close-out: ADRs + docs updated; delete this folder; drop In Progress line in owner `todo.md`.
