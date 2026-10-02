# Plan index — install-distribution

Ephemeral plan: **installable yerk binaries** for operators (mise and similar),
including version identity on shipped builds.

Does **not** cover CI check workflows — that is [../ci-pipelines/](../ci-pipelines/).
Soft preference: CI green before tagging releases, but not a hard merge of plans.

| File | Role |
|------|------|
| [plan.md](./plan.md) | Goals, phases, exit criteria |
| [decisions.md](./decisions.md) | **Only** Q&A inbox (`Q*`) |
| [checklist.md](./checklist.md) | Execution track |
| [../todo.md](../todo.md) | Owner In Progress / Backlog |

## Phase order

| # | Phase | Status |
|---|--------|--------|
| 0 | Lock install/release decisions | pending |
| 1 | Version identity (ldflags / tags) | pending |
| 2 | Release artifacts + channel | pending |
| 3 | Operator install docs + smoke | pending |

## Protocol

- One step at a time on “go” unless operator widens scope.
- No git stage/commit unless asked.
- Never cite `Q*` in product code.
- Close-out: delete this folder; drop from owner `todo.md`.
