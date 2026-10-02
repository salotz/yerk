# Plan: CI pipelines

## Context

Repository: **yerk** (`salotz.yerk`). Core product usable after
initial-feature-series. Local quality bar already exists:

```sh
mise run check   # go test ./... && go vet ./...
mise run build   # → .local/bin/yerk
```

This plan adds **remote CI** that mirrors that bar. Install/release channels are
out of scope ([install-distribution](../install-distribution/)).

Design spine: [design/domain-and-near-term.md](../../../../design/domain-and-near-term.md).
Local tasks: [mise.toml](../../../../mise.toml), [contributing/development.md](../../../../contributing/development.md).

---

## Goals

1. Automated **test + vet + build** on the locked trigger policy.
2. **Same Go pin** as local mise (no silent drift).
3. Fast, readable failures; no host-private config in workflows (ADR 007).

## Non-goals

- Publishing release assets / mise install backends
- Multi-OS release matrices (unless Q5 expands CI only)
- Coverage gates, security scanners, dependency bots (later plans if wanted)
- Changing product CLI behavior

---

## Phase 0 — Decisions

Lock Q1–Q5 in [decisions.md](./decisions.md) (host, triggers, jobs, Go pin, OS).

**Exit:** enough to write a workflow without re-asking.

---

## Phase 1 — Workflow

### Intent

Every locked event runs roughly:

```text
go test ./...
go vet ./...
go build -o /tmp/yerk ./cmd/yerk
```

### Work

1. Add workflow under the chosen CI root (e.g. `.github/workflows/ci.yml`).
2. Install/use Go from the pin; optional module cache.
3. Fail the job on any non-zero step.

### Exit

CI runs green on a known-good tree; breaks when tests fail.

---

## Phase 2 — Docs + hygiene

1. Link CI from `contributing/development.md` (local check ≡ CI intent).
2. Confirm no secrets or machine paths in workflow YAML.
3. Optional badge in README only if operator asks.

### Exit

A new contributor can see how CI relates to `mise run check` / `build`.

---

## Success criteria

- [ ] CI runs test + vet + build with pinned Go.
- [ ] Triggers match locked policy.
- [ ] Docs mention CI ↔ local tasks.
- [ ] Plan folder removable at close.

---

## Execution protocol

- Answers only in [decisions.md](./decisions.md).
- One execute step per “go” unless operator widens scope.
- No git stage/commit unless asked.

## Immediate next step

**On go:** Phase 0 — lock Q1–Q5 (or accept defaults in decisions.md), then
implement Phase 1 workflow.
