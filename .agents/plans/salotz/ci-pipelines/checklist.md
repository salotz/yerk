# Checklist — ci-pipelines

## Phase 0 — Decisions

- [x] Q1–Q5 locked ([decisions.md](./decisions.md))
- [x] Owner In Progress points here

## Phase 1 — Workflow

- [x] Workflow file(s) under chosen CI root (`.github/workflows/ci.yml`)
- [x] Jobs: test, vet, build (no fmt gate)
- [x] Go version matches `mise.toml` pin (`1.27.1`)
- [x] Module cache via `setup-go` cache
- [ ] Green on sample PR/push (needs remote push / Actions enabled)

## Phase 2 — Docs + hygiene

- [x] `contributing/development.md` points at CI and local `mise run check`
- [x] No host-private secrets or paths (ADR 007)
- [x] Optional: status badge only if operator wants it — **skipped**

## Close-out

- [ ] Success criteria in [plan.md](./plan.md) checked after first green remote run
- [ ] Delete this plan folder
- [ ] Drop In Progress line in [../todo.md](../todo.md) (operator confirms)
