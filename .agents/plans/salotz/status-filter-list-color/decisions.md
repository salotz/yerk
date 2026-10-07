# Decisions inbox — status-filter-list-color

Plan-local Q&A only. Durable outcome: ADR under `design/decisions/` (next
free id; not 025).
Do **not** cite `Q*` in product code.

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q1 | proposed | v1 convenience flag set |
| Q2 | proposed | Combine semantics (AND vs OR vs XOR) |
| Q3 | proposed | Filter grain: project overall vs replica rows |
| Q4 | proposed | `--status=` timing / share tag DSL |
| Q5 | proposed | Cross-project replica listing CLI shape |
| Q6 | proposed | Color: auto TTY, palette, deps, `--color` |
| Q7 | proposed | `--presence-only` vs change filters |
| Q8 | proposed | Empty match, preamble, structured output |
| Q9 | proposed | `--present` / `--missing` on **project** overall (all-present vs any) |
| Q10 | proposed | `--ahead` / `--behind` in v1? |

## Locked

_(none yet)_

---

## Q1 — v1 convenience flag set

Status: proposed

### Prompt
Which boolean flags ship with the first implementation?

### Context
Operator asked for `--dirty`, `--clean`, `--missing`, `--present`, “etc.”
Change also has `untracked`, `ahead:N`, `behind:N`, `no-upstream`, `error`.
Overall presence also has `partial`, `invalid`, `none`.

### Answer
<!-- operator writes here -->
Proposed: `--present --missing --invalid --partial --dirty --clean --untracked`.
Hold `--ahead --behind --error --none` for `--status=` or a follow-on.

### Notes
Keep the set small and name-aligned with displayed tokens.

---

## Q2 — Combine semantics

Status: proposed

### Prompt
If several convenience flags are passed, is that AND, OR, or forbidden?

### Context
Later `--status='dirty AND present'` implies AND is the expression default.
`--dirty --clean` is a contradiction at replica grain.

### Answer
<!-- operator writes here -->
Proposed: **AND** across flags. Contradiction → empty list, not an error.
OR only via later `--status=`. Convenience flags XOR with `--status=` when
that flag exists. Catalog `--tag` / `--domain` remain orthogonal (apply
first, then status filter).

### Notes
—

---

## Q3 — Filter grain

Status: proposed

### Prompt
On each invocation, what is filtered?

### Context
Project table uses **overall** rollup (union of problem change tokens;
presence `all-present|partial|missing|…`). Replica table uses per-checkout
presence + change bag.

### Answer
<!-- operator writes here -->
Proposed:

- `yerk status` → filter **project** rows by overall atoms.
- `yerk status --replicas` → filter **replica** rows.
- `yerk status <project>` → keep project summary; filter the replica table.
- `yerk status <project> <replica>` → emit row if match, else empty-ok.

### Notes
—

---

## Q4 — `--status=` expression

Status: proposed

### Prompt
When does `--status=` ship, and which engine?

### Context
[tag-boolean-select](../tag-boolean-select/) is design-only; engine unset
(lean: small boolean DSL, CEL deferred). Operator wants the **same** logic
system as tags. Example: `--status='dirty AND present'` (typo `diry` in the
request).

### Answer
<!-- operator writes here -->
Proposed: **defer implementation**. ADR here names the flag and atoms, forbids
a stub, and requires sharing the tag-select engine once that ADR accepts.
This plan ships convenience flags only.

### Notes
Do not grow an ad-hoc status parser that later conflicts with the tag DSL.

---

## Q5 — Replica listing CLI

Status: proposed

### Prompt
How do operators list all replicas across all projects?

### Context
`yerk replica` has get / lookup / create only. Live replicas are discovered
on disk, not cataloged as first-class rows (except default included even if
missing).

### Answer
<!-- operator writes here -->
Proposed: **`yerk status --replicas`** flattens selected projects into the
existing replica table. Same `--tag` / `--domain` / names selection as
status. Error if combined with a replica argument. Optional `yerk replica
list` alias: **no** unless operator wants the noun.

### Notes
Preserves “project table is the default view” (domain spine).

---

## Q6 — Color

Status: proposed

### Prompt
When to color, which tokens, which dependency, which flag?

### Context
Module deps today: cobra + toml only. `text/tabwriter` breaks if ANSI is
counted as width. json/yaml must stay plain (ADR 019). gita colors remote
*situation*; yerk should color **words**.

### Answer
<!-- operator writes here -->
Proposed: no new deps; auto color on TTY unless `NO_COLOR`; `--color=auto|always|never`;
color presence/change **tokens** (palette in plan.md); never color structured
output; pad then wrap so columns stay aligned.

### Notes
Palette can be tweaked in ADR without a new Q if operator prefers different
hues.

---

## Q7 — `--presence-only` vs change filters

Status: proposed

### Prompt
What if `--presence-only --dirty`?

### Context
Change column is `-` when probes are skipped. Matching `--dirty` would
silently match nothing.

### Answer
<!-- operator writes here -->
Proposed: **error** if any change atom flag is set together with
`--presence-only`. Presence flags (`--present` / `--missing` / …) remain
valid.

### Notes
—

---

## Q8 — Empty match, preamble, structured output

Status: proposed

### Prompt
Empty filter: message vs silent empty table? Does json include dropped rows?
Preamble line?

### Context
`status --tag` with zero projects prints `No projects matched tag …` and
exits 0. Status is a read op.

### Answer
<!-- operator writes here -->
Proposed: empty-ok, short message (`No projects matched status filter.` /
`No replicas matched status filter.`), exit 0. json/yaml: empty array (or
omit), **not** unfiltered data. Human preamble: `filter.status=dirty,present`
(canonical atom order).

### Notes
—

---

## Q9 — `--present` / `--missing` on project overall

Status: proposed

### Prompt
Does `--present` on the project table mean `all-present`, or “any replica
present” (includes `partial`)?

### Context
Operator example is “show only the dirty projects/replicas.” For presence,
`partial` is a third state. Replica grain is unambiguous (`present` vs
`missing`).

### Answer
<!-- operator writes here -->
Proposed: project grain maps `--present` → overall `all-present`;
`--missing` → overall `missing`; `--partial` → `partial`; `--invalid` →
`invalid`. “Any replica present” is later `--status=` (e.g. `present OR
partial`) or `--replicas --present`.

### Notes
—

---

## Q10 — ahead / behind in v1

Status: proposed

### Prompt
Ship `--ahead` / `--behind` convenience flags now?

### Context
Tokens are `ahead:N` / `behind:N`. A flag would mean “ahead count > 0”
(any N). Numeric thresholds (`--ahead=2`) are extra surface.

### Answer
<!-- operator writes here -->
Proposed: **not in v1**. Use later `--status='ahead'` (prefix match) or a
follow-on flag. Avoid value-taking flags in the first convenience set.

### Notes
—

---

## Constraints from product

- No unimplemented CLI stubs.
- Presence ≠ change; do not treat missing as clean.
- Catalog bulk XOR (names / `--tag` / `--domain`) unchanged (ADR 022).
- Examples vs host state (ADR 007): tests use temp dirs only.
- Keep module deps minimal unless Q6 adds a color library.
- Never cite plan `Q*` in product code or durable docs.

## Changelog

- 2026-10-07: inbox created; Q1–Q10 proposed from operator request.
