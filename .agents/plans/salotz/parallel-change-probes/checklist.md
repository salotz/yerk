# Checklist — parallel-change-probes

## Phase 0 — Decisions

- [ ] Q1–Q5 locked ([decisions.md](./decisions.md))

## Phase 1 — Implementation

- [ ] Parallel probe in status collection path
- [ ] Semaphore / errgroup (or equivalent) with locked limit
- [ ] Stable output ordering

## Phase 2 — Tests

- [ ] Fake Runner records concurrency or at least correctness under parallel call
- [ ] Existing status CLI tests still pass

## Phase 3 — Docs

- [ ] Note in status how-to / explanation if behavior/timing changes
- [ ] Env knob documented if Q5 adds one (ADR 005 / envvars registry)

## Close-out

- [ ] Success criteria checked; delete folder; update todo
