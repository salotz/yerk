# Checklist — ci-pipelines

## Phase 0 — Decisions

- [ ] Q1–Q5 locked or deferred with defaults ([decisions.md](./decisions.md))
- [ ] Owner In Progress points here

## Phase 1 — Workflow

- [ ] Workflow file(s) under chosen CI root
- [ ] Jobs: test, vet, build (fmt check if locked)
- [ ] Go version matches `mise.toml` pin
- [ ] Module cache optional but preferred
- [ ] Green on sample PR/push (or dry-run as operator allows)

## Phase 2 — Docs + hygiene

- [ ] `contributing/development.md` points at CI and local `mise run check`
- [ ] No host-private secrets or paths (ADR 007)
- [ ] Optional: status badge only if operator wants it

## Close-out

- [ ] Success criteria in [plan.md](./plan.md) checked
- [ ] Delete this plan folder
- [ ] Drop In Progress line in [../todo.md](../todo.md) (operator confirms)
