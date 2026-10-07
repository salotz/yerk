# Plan: tag boolean / expression select (design only)

## Context

Catalog tags are a **closed vocabulary** (ADR 010). Bulk select today is
single-label membership only:

- `yerk status --tag <name>`
- `yerk materialize --tag <name>`
- `yerk workspace ensure --tag <name>`

Shared path: `Catalog.SelectByTag` / `FilterTag`. Domain bulk is separate and
**XOR** with tag / names / `--all` (ADR 022); tag∩domain in one flag set was
explicitly deferred.

Projects may already carry multiple tags (`tags = ["devel", "go"]`), but the
CLI cannot express combinations. Operators want predicates such as:

```text
science AND NOT deprecated
```

Idea under consideration: reuse a general expression language (e.g. Google
**CEL**) instead of rolling ad-hoc boolean logic — weighed against lighter
equivalents and yerk’s small dependency surface (toml + cobra only today).

Product rule: do not ship CLI stubs without behavior. This plan produces an
**ADR** (and pointers). Implementation is a **separate** plan after accept.

Baseline:

- [ADR 010](../../../../design/decisions/010-catalog-tag-vocabulary.md) — closed tag vocabulary; `SelectByTag`
- [ADR 022](../../../../design/decisions/022-domain-bulk-selection.md) — domain bulk; mutual exclusion; no tag∩domain yet
- [design/domain-and-near-term.md](../../../../design/domain-and-near-term.md) — Tag noun; Project selection (bulk)
- `internal/config/catalog.go` — `HasTag`, `FilterTag`, `SelectByTag`

---

## Goals

1. Accepted **ADR** (or explicit reject/defer with rationale) for tag
   combination / expression selection.
2. Clear **engine choice** and **non-goals** (what the language is not).
3. Locked **CLI shape**, validation rules, empty-match policy, bulk XOR rules.
4. Handoff: implement plan folder or explicit backlog defer.

## Non-goals (this plan)

- Implementing any new CLI flags or parsers
- Pulling CEL / expr / other expression deps into the module
- General project predicate language beyond tags (presence, dirty, domain field
  in-expr, placement) — may be named as a **later** horizon only
- Catalog mutation UX / `yerk tags` command (still hand-edit; ADR 010)
- Tag hierarchy, colors, or metadata beyond the name string
- Changing domain closed-vocabulary status

---

## Design sketch (tentative — lock in ADR)

Captured from the 2026-10-03 exploration. Not decisions until ADR + Q-lock.

### Problem shape

Evaluation model is boolean algebra over a **finite closed set** of labels:

```text
context: project.Tags (list of strings ⊆ catalog.Tags)
expr:    membership + AND / OR / NOT + grouping
```

Unknown identifiers should fail like unknown `--tag` (ADR 010), not silently
match nothing.

### Engine options (ranked for yerk-now)

| Rank | Approach | Fit |
|------|----------|-----|
| 1 | **Small boolean DSL** (`science AND NOT deprecated`, optional `&&` / `!`) | Best near-term: tiny, validates against vocabulary, no heavy deps, good errors |
| 2 | **Compositional flags** (`--tag` + `--without-tag`; multi-tag AND/OR) | Zero parser; bridge if most queries are shallow; nesting gets awkward |
| 3 | **K8s-style label selector** (`science,!deprecated`) | Familiar in cluster-land; comma-AND less readable; still a grammar or dep |
| 4 | **expr-lang/expr** | Lighter general Go embed than CEL; still overkill for tags-only |
| 5 | **CEL (`cel-go`)** | Right when predicates grow past tags (typed policy language); heavy for labels-only |
| 6 | **Starlark / Lua / JS** | Reject for this surface |
| — | **jq / JSON pipe escape hatch** | Keep as Unix compose; not first-class bulk mutate |

**Working recommendation:** ship a **minimal boolean tag DSL** (or flags first)
when implementing; **do not** adopt CEL until at least two non-tag predicate
kinds share one language. Revisit CEL/expr under a separate “project predicate
language” ADR if selection outgrows tags.

### CLI shape (candidates — pick in ADR)

- **A.** Overload `--tag 'science AND NOT deprecated'` (simple names remain valid exprs).
- **B.** New flag e.g. `--tags-expr` / `--select` / `--filter`; keep `--tag` single-name forever.
- **C.** Flags only: `--tag science --without-tag deprecated` (± multi `--tag` as AND).

Shell note: prefer word operators `AND` `OR` `NOT` over `&&` `!` for quoting;
document required quotes either way.

### Invariants to preserve

1. Closed vocabulary: every identifier in the expr ∈ `catalog.Tags` (select-time error).
2. Read vs mutate empty-match: status empty-ok; materialize / ensure empty-error (ADR 010/022).
3. Bulk modes remain **one mode** unless ADR deliberately allows intersection
   (today: names XOR `--all` XOR `--tag` XOR `--domain`).
4. Keep `SelectByTag` simple; add `SelectByTagExpr` (or general `Select`) beside it.
5. No unimplemented cobra commands from this design plan.

### Later horizon (out of near-term ADR unless operator expands scope)

If predicates later include domain, presence, change/dirty, host placement, etc.,
then a real expression engine (CEL or expr) may pay off. Design one activation
context once and ADR it separately — do not grow the tag DSL into an accidental
general language.

---

## Phase 0 — Scope

Lock Q1–Q8 in [decisions.md](./decisions.md): need timing, engine, CLI shape,
grammar, XOR/intersection with domain, empty policy, implement timing.

---

## Phase 1 — ADR

Draft ADR under `design/decisions/`:

- Motivation and examples
- Chosen surface (DSL vs flags vs defer engine)
- Grammar or flag matrix
- Validation (closed vocabulary)
- Bulk selector mutual exclusion
- Errors and empty-match
- Explicit non-goals (CEL deferred unless…)
- Related: 010, 022, domain spine bulk table

Operator review → Accepted, revise, or defer-without-ADR.

---

## Phase 2 — Handoff

- Spawn `tag-boolean-select-implement` plan **or** leave backlog defer line.
- Update domain spine “Project selection (bulk)” when ADR accepts.
- Docs stubs only if they do not claim shipped syntax.
- Do not add cobra flags in this plan.

---

## Success criteria

- [ ] ADR accepted, explicitly deferred, or rejected with rationale.
- [ ] Engine/CLI choice recorded (including “flags-only bridge” if chosen).
- [ ] No orphan CLI stubs or new expression deps from this plan.
- [ ] Owner todo reflects implement vs defer.
- [ ] Plan folder removable at close.

## Immediate next step

When In Progress: Phase 0 open-queue pass with operator (especially Q1 engine
and Q2 CLI shape).
