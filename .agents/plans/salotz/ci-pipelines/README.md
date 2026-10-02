# Plan index — ci-pipelines

Ephemeral plan: **continuous integration** for yerk (test / vet / build).

Does **not** cover binary install channels or mise plugin packaging — that is
[../install-distribution/](../install-distribution/).

| File | Role |
|------|------|
| [plan.md](./plan.md) | Goals, phases, exit criteria |
| [decisions.md](./decisions.md) | **Only** Q&A inbox (`Q*`) |
| [checklist.md](./checklist.md) | Execution track |
| [../todo.md](../todo.md) | Owner In Progress / Backlog |

## Phase order

| # | Phase | Status |
|---|--------|--------|
| 0 | Lock CI decisions | **next** |
| 1 | Workflow: check + build | pending |
| 2 | Docs + hygiene | pending |

## Protocol

- One step at a time on “go” unless operator widens scope.
- No git stage/commit unless asked.
- Promote durable choices to ADRs only if needed; never cite `Q*` in product code.
- Close-out: delete this folder; drop In Progress line in owner `todo.md`.
