# Plan: sync verbs (design only)

## Context

yerk can **materialize** replicas and report **change** (including ahead/behind
style flags). It does not yet define **pull** / **push** (or broader sync).

Product rule: do not ship CLI stubs without behavior. This plan produces an
**ADR** (and pointers). Implementation is a **separate** plan after accept.

Baseline: status model docs, ADR 016 (replica create), ADR 020 (git subprocess),
[design/domain-and-near-term.md](../../../../design/domain-and-near-term.md).

---

## Goals

1. Accepted **ADR** for sync verb semantics and safety.
2. Clear **non-goals** (e.g. multi-remote orchestration, mutagen, …).
3. Handoff: implement plan folder or explicit backlog defer.

## Non-goals (this plan)

- Implementing `pull`/`push` CLI
- Changing status column layout except doc cross-links
- Parallel probes (separate plan)

---

## Phase 0 — Scope

Lock Q1–Q6: verb set, selection scope, dirty/force policy, relation to existing
verbs, network defaults, when to implement.

---

## Phase 1 — ADR

Draft ADR: commands, flags sketch, idempotency, errors, bulk `--tag`/`--all`
policy, interaction with bound placement (sync does not restyle paths).

Operator review → Accepted or revise.

---

## Phase 2 — Handoff

- Spawn `sync-verbs-implement` plan **or** leave a backlog line.
- Update domain spine “push/pull design” bullet to point at the ADR.
- Do not add cobra commands in this plan.

---

## Success criteria

- [ ] ADR accepted (or explicitly rejected with rationale).
- [ ] No orphan CLI stubs.
- [ ] Owner todo reflects implement vs defer.
- [ ] Plan folder removable at close.

## Immediate next step

When In Progress: Phase 0 open-queue pass with operator.
