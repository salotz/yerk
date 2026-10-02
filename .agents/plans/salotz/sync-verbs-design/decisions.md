# Decisions inbox — sync-verbs-design

Plan-local Q&A only. Durable outcome: one or more ADRs under `design/decisions/`.
Do **not** cite `Q*` in product code.

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q1 | open | Verb set: `pull`/`push` only vs also `fetch`, sync-status helpers |
| Q2 | open | Scope unit: replica vs project (all live replicas) vs tag bulk |
| Q3 | open | Safety: refuse dirty? ff-only? explicit force flags? |
| Q4 | open | Relation to `materialize` / `replica create` / `status` change flags |
| Q5 | open | Network defaults and offline behavior |
| Q6 | open | Implement plan timing (immediate after ADR vs backlog) |

## Locked

_(none yet)_

---

## Constraints from product

- No unimplemented CLI stubs (domain spine / ADR practice).
- Git stays subprocess adapter (ADR 020) unless a new ADR says otherwise.
- Status already surfaces ahead/behind-style change bags; sync should not fork a
  second mental model without reason.
