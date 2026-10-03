# Checklist — install-distribution

## Phase 0 — Decisions

- [ ] Q1–Q5 locked or deferred with defaults ([decisions.md](./decisions.md))
- [ ] Short ADR or design note if channel choice is durable

## Phase 1 — Version identity

- [x] Tag / version scheme documented (ADR 023; `mise run version-show` / `version-bump`)
- [ ] `mise run build` (and release build) can stamp `internal/version` from tag
- [ ] `yerk version` reflects stamp on tagged builds; dev builds still sensible

## Phase 2 — Artifacts + channel

- [ ] Release asset naming stable enough for mise/ubi/aqua (per Q1)
- [ ] Checksums if locked
- [ ] Channel config or docs for discovering assets (no secrets in-tree)

## Phase 3 — Docs + smoke

- [ ] How-to: install / upgrade / verify (`docs/how-to/` or contributing)
- [ ] Smoke: install path → `yerk version` + `--help`
- [ ] Point dogfood at XDG config (ADR 007); examples only in-repo

## Close-out

- [ ] Success criteria in [plan.md](./plan.md) checked
- [ ] Durable ADRs under `design/decisions/` if any
- [ ] Delete this plan folder
- [ ] Update [../todo.md](../todo.md) (operator confirms)
