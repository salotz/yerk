# salotz plans

Owner index for `.agents/plans/salotz/`.
Work-process: agent-guidelines personal work-process (In Progress / Backlog only).

One logical feature (or tight feature set) per plan folder. Do not merge
unrelated tracks into a single “mega plan.”

## In Progress

### ci-pipelines

Check workflow shipped (`.github/workflows/ci.yml`). **Next:** confirm remote
green, then close-out.
Plan: [./ci-pipelines/](./ci-pipelines/).

## Backlog

### install-distribution

Version stamp + GH Release + **packslip** on `v*` tags (ADR 024). `v0.0.1` live.
**Still open:** multi-OS (Q2); `go install` (Q5); install how-to polish.
**Todo:** when a **released** packslip CLI ships `packslip pin` / `verify --pin`
(not in 1.4.0), run pin on a trusted yerk bundle and publish `ps1_…` in README
(+ ADR 024 / development.md). Do not hand-derive pins for product docs.
Plan: [./install-distribution/](./install-distribution/).

### sync-verbs-design

Design-only ADR for `pull` / `push` (no CLI stubs). Carried from
initial-feature-series Phase 9.
Plan: [./sync-verbs-design/](./sync-verbs-design/).

### tag-boolean-select

Design-only ADR for boolean / expression bulk select over catalog tags
(e.g. `science AND NOT deprecated`). Lean: small DSL or flags first; CEL
deferred unless predicates grow past tags. No CLI stubs.
Plan: [./tag-boolean-select/](./tag-boolean-select/).

### parallel-change-probes

Parallelize status git change probes (limit, stable order, fake adapter tests).
Plan: [./parallel-change-probes/](./parallel-change-probes/).

### domain-model-glossary

Canonical glossary of product nouns; cross-links from docs hubs.
Plan: [./domain-model-glossary/](./domain-model-glossary/).

### domain-spine-architecture-move

Rename/move `design/domain-and-near-term.md` → architecture home; rewire links.
May bundle link pass with glossary.
Plan: [./domain-spine-architecture-move/](./domain-spine-architecture-move/).

### docs-deadwood-cleanup

Stale docs inventory/fix + reusable agent cleanup checklist/role.
Plan: [./docs-deadwood-cleanup/](./docs-deadwood-cleanup/).

## Done (recent)

### initial-feature-series

Closed 2026-10-02. Phases 0–8 shipped. Plan folder removed. Sync design and
distribution tracks split into the backlog plans above (not one combined plan).

### distribution-and-followons

Never executed as a mega-plan; **split** 2026-10-02 into ci-pipelines (In
Progress), install-distribution, sync-verbs-design, and the former backlog
items as first-class plan folders.
