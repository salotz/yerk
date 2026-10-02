# Plan: install distribution

## Context

**yerk** is a Go standalone binary (ADR 002). Developers already:

```sh
mise install && mise run build
.local/bin/yerk version
```

Operators who do not want a full clone need an **install channel** (mise and
friends) and **versioned artifacts**. CI check workflows are a separate plan
([ci-pipelines](../ci-pipelines/)).

Baseline: [mise.toml](../../../../mise.toml), [internal/version](../../../../internal/version),
[contributing/development.md](../../../../contributing/development.md), ADR 007.

---

## Goals

1. Documented **install path** via mise (or locked alternative).
2. **Version identity** on release binaries (`yerk version` ≈ tag).
3. Stable **artifact** layout for the chosen channel (names, OS/arch, checksums if needed).
4. No host-private paths or tokens in the shared tree (ADR 007).

## Non-goals

- Product feature work (sync, probes, …)
- Replacing local `.local/bin` dev workflow
- OS package managers (deb/rpm/homebrew) unless later pulled in
- go-git / pure-Go packaging stories

---

## Phase 0 — Decisions

Lock Q1–Q5: channel, artifact matrix, version scheme, release automation,
`go install` yes/no.

**Exit:** implement without re-litigating backend choice.

---

## Phase 1 — Version identity

1. Define stamp: `-ldflags "-X …Version=…"` (exact path from code).
2. Wire optional stamp into `mise run build` and/or release script.
3. Document dirty/dev behavior when untagged.

**Exit:** tagged build shows expected version string.

---

## Phase 2 — Artifacts + channel

1. Produce release binaries for locked OS/arch set.
2. Publish where the mise (or other) backend expects them.
3. Register or document the tool definition (ubi URL pattern, aqua registry, etc.).

**Exit:** a clean machine can install yerk without cloning this repo.

---

## Phase 3 — Docs + smoke

1. How-to install/upgrade/verify.
2. Cross-link from README / contributing.
3. Smoke checklist for the operator.

**Exit:** dogfood install from the channel works.

---

## Success criteria

- [ ] Install without full source checkout (locked channel).
- [ ] Release binary version matches tag scheme.
- [ ] Docs sufficient for a new operator.
- [ ] Plan folder removable at close.

---

## Soft dependencies

- **ci-pipelines**: useful before trusting release automation; not required to
  start Phase 0–1 (version stamp can be local).
- GitHub (or other) **releases** if Q1 needs downloadable assets.

## Immediate next step

When promoted to In Progress: Phase 0 lock Q1–Q5 (mise backend + linux-first
defaults are the expected starting point).
