# Decisions inbox — tag-boolean-select

Plan-local Q&A only. Durable outcome: one ADR under `design/decisions/` (or
explicit defer note in domain spine / owner todo).
Do **not** cite `Q*` in product code.

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q1 | open | Engine: small boolean DSL vs compositional flags vs label-selector vs expr vs CEL vs defer |
| Q2 | open | CLI shape: overload `--tag`, new flag (`--tags-expr` / `--select`), or flags-only |
| Q3 | open | Grammar: `AND`/`OR`/`NOT` words only vs also `&&`/`\|\|`/`!`; required parens? |
| Q4 | open | Multi-tag without full expr: repeated `--tag` = AND? OR? disallowed? |
| Q5 | open | Intersection with `--domain` / names (reopen ADR 022 XOR or keep one-mode) |
| Q6 | open | Unknown identifier policy (always error — confirm) and empty-match read vs mutate |
| Q7 | open | Horizon: tags-only forever in this ADR vs named escape to future predicate language |
| Q8 | open | Implement plan timing (after ADR vs long backlog) |

## Locked

_(none yet)_

### Tentative lean (not locked)

From 2026-10-03 exploration — operator has **not** confirmed:

- Prefer **minimal boolean DSL** (or flags bridge) over CEL for tags-only.
- Defer CEL until ≥2 non-tag predicate kinds share one language.
- Preserve ADR 010 closed-vocabulary errors and ADR 022 one-mode bulk unless
  explicitly reopened.

---

## Constraints from product

- No unimplemented CLI stubs (domain spine / ADR practice).
- Tag vocabulary stays closed at catalog root (ADR 010).
- Domain remains orthogonal; bulk selectors mutually exclusive unless a new ADR
  changes that (ADR 022).
- Keep module deps minimal unless the ADR justifies an embeddable engine.
- Do not invent tag hierarchy/metadata in this track.
