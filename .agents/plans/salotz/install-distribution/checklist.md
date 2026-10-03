# Checklist — install-distribution

## Phase 0 — Decisions

- [ ] Q1–Q5 locked or deferred with defaults ([decisions.md](./decisions.md))
- [ ] Short ADR or design note if channel choice is durable

## Phase 1 — Version identity

- [x] Tag / version scheme documented (ADR 023; `mise run version-show` / `version-bump`)
- [x] `mise run build` (and release build) stamps `internal/version` via `.tasks/build-yerk`
- [x] Dev builds default `0.0.0-dev`; tagged/YERK_VERSION builds stamp B.R.G

## Phase 1b — Release CI (partial channel)

- [x] `.github/workflows/release.yml` on `vB.R.G` → GH Release linux/amd64 + checksums
- [ ] First remote tag run green (operator)
- [ ] mise/ubi (or other) install channel still open (Q1)

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
