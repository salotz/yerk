# Decisions inbox — parallel-change-probes

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q1 | open | Concurrency limit (fixed N vs GOMAXPROCS-derived) |
| Q2 | open | Result ordering (catalog order vs completion order + stable sort) |
| Q3 | open | Error policy (fail one row vs fail command) |
| Q4 | open | Whether presence-only paths stay serial (cheap) |
| Q5 | open | Config/env knob for limit (`YERK__…`) vs constant first |

## Locked

_(none yet)_

## Constraints

- Keep `gitcmd.Runner` seam (ADR 020); do not invent ad-hoc shell-outs.
- Human status table order should remain predictable for operators.
