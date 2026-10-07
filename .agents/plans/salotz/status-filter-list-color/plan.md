# Plan: status filters, replica listing, and color

## Context

`yerk status` today is **project-first** and unfiltered after probes:

| Invocation | View |
|------------|------|
| `yerk status` | Multi-project table: **overall** presence/change + replica count + workspace path |
| `yerk status <project>` | Project summary + **every live replica** row |
| `yerk status <project> <replica>` | One replica row |
| `yerk status --tag` / `--domain` | Project list, catalog selection only |

There is **no** post-probe filter (`--dirty`, `--missing`, …), **no** way to
list replicas across the catalog (only inside one project), and **no** color.
Human printers use `text/tabwriter` with plain tokens.

Change is a **bag of flags** (not one enum); presence is a separate axis.
Project overall presence is `all-present | partial | missing | invalid | none`;
replica presence is `present | missing | invalid`. See
[status-model.md](../../../../docs/explanation/status-model.md).

Related (do not duplicate):

- [tag-boolean-select](../tag-boolean-select/) — design-only for tag
  expressions. `--status=` must **reuse that engine**, not invent a second
  DSL. Until that ADR accepts, ship **convenience flags only**.
- [parallel-change-probes](../parallel-change-probes/) — speed; independent.
  Filters still run **after** collection.

Baseline:

- [design/domain-and-near-term.md](../../../../design/domain-and-near-term.md) — status scopes; multi-replica list is not the default view
- [ADR 010](../../../../design/decisions/010-catalog-tag-vocabulary.md) / [ADR 022](../../../../design/decisions/022-domain-bulk-selection.md) — catalog selectors (`--tag` / `--domain`) stay XOR with names
- [ADR 011](../../../../design/decisions/011-api-resources.md) / [ADR 019](../../../../design/decisions/019-output-formats.md) — `ProjectStatus` / `ReplicaStatus`; json/yaml uncolored
- `internal/cli` `newStatusCmd` / printers; `internal/project.ProjectStatus`; `gitcmd.ProbeResult.Flags`

---

## Goals

1. Filter status output by **observed state**: convenience flags such as
   `--dirty`, `--clean`, `--missing`, `--present` (and a small closed set of
   siblings).
2. Reserve a generic **`--status=`** expression for later, using the **same**
   boolean system as catalog tags (e.g. `--status='dirty AND present'`).
   Not implemented in this plan.
3. Add a **cross-project replica listing** (every live replica, not one
   project at a time).
4. Color human status **presence** and **change** cells so states separate at
   a glance (TTY auto; json/yaml untouched).
5. Accepted **ADR** covering filter semantics, replica grain, and color
   policy (next free id; do **not** claim 025 — materialize-dwim already
   sketches that number).

## Non-goals

- Implementing `--status=` / pulling CEL / sharing a parser before
  tag-boolean-select locks an engine
- Changing catalog selectors (`--tag` / `--domain` XOR rules)
- Fetch-then-probe, stash, staged-vs-unstaged split
- Color on non-status commands
- Parallel probes (other plan)
- New product root or extra git adapter

---

## Design sketch (tentative — lock in ADR + Qs)

### 1. Status atoms

Closed vocabulary of **atoms** used by convenience flags and, later,
`--status=`. Case-insensitive identifiers.

| Axis | Replica grain | Project overall grain |
|------|---------------|------------------------|
| Presence | `present`, `missing`, `invalid` | `all-present`, `partial`, `missing`, `invalid`, `none` |
| Change | `clean`, `dirty`, `untracked`, `no-upstream`, `sync-unknown`, `error`; `ahead` / `behind` match `ahead:N` / `behind:N` | same tokens on the **union** bag (overall drops lone `clean` when mixed) |

`--present` on a **project** row means overall presence is `all-present`
(not “at least one replica present”). `--missing` on a project row means
overall `missing` (every collected replica missing). Partial / mixed needs
`--partial` or later `--status=`. Confirm in Qs.

Convenience flags proposed for v1 (boolean, no values):

```text
--present --missing --invalid --partial
--dirty --clean --untracked
```

Hold `--ahead` / `--behind` / `--error` for `--status=` or a follow-on unless
Q1 expands v1. `--presence-only` already exists (skip probes); it is **not**
a row filter.

### 2. Convenience flag algebra

Working recommendation (Q2):

- Multiple flags are **AND** (must satisfy every flag).
- `--dirty --present` → dirty **and** present.
- `--dirty --clean` → empty match (contradiction), not an error.
- **OR** waits on `--status='dirty OR untracked'`.
- Catalog selectors stay a **separate** axis: `--tag devel --dirty` is valid
  (select projects by tag, then filter observed state). Not a bulk-XOR
  violation: XOR is among *which projects to consider*, not post-probe filter.

`--status=` (later) is mutually exclusive with convenience flags (one
predicate). Unknown atoms error (like unknown `--tag`).

### 3. Where the filter applies (grain)

| Invocation | What is filtered |
|------------|------------------|
| `yerk status` (project table) | **Project** rows, matching **overall** atoms |
| `yerk status --replicas` (proposed) | **Replica** rows across selected projects |
| `yerk status <project>` | Replica table inside the detail view (summary still printed? Q3) |
| `yerk status <project> <replica>` | Print the row if it matches; empty-ok if not (read op) |

Collect first, then filter (need probes for change atoms). Empty filter
result is **ok** for status (same as empty `--tag` match). Structured
`--output json|yaml` is the filtered resource list, not the unfiltered
universe.

`--presence-only` + a **change** flag (`--dirty`, `--clean`, `--untracked`)
→ **error** (cannot evaluate). Presence flags still work.

Human preamble should show the filter, e.g. `filter.status=dirty,present`,
alongside `filter.tag=` / `filter.domain=`.

### 4. Cross-project replica listing

Gap: live replicas are only listed under `yerk status <project>`.
`yerk replica` has get / lookup / create — no list.

Working recommendation (Q5):

```text
yerk status --replicas
yerk status --replicas --tag devel --dirty
```

Same project selection as `yerk status` (`--tag` / `--domain` / all / one
project), then flatten `ProjectStatus.Replicas` into the existing replica
table (`PROJECT REPLICA DOMAIN PRESENCE CHANGE BRANCH PATH`).

`--replicas` with `status <project> <replica>` is an error (already one
replica). `--replicas` with `status <project>` is redundant with the detail
replica table; either ignore or error — pick in Q5.

Optional noun alias `yerk replica list` → same printer. Default **no**:
observed listing is `status`; keep `replica` for get/lookup/create until a
pure inventory (no probes) is needed.

Domain spine today: “multi-replica lists are not the default view.” This
flag is the **opt-in** default-view exception.

### 5. Color

Human table / detail only. Never json/yaml.

Working recommendation (Q6):

- **No new module dependency** (still cobra + toml). Tiny `internal/cli`
  (or `internal/termcolor`) ANSI helper.
- Default **auto**: color when stdout is a TTY and `NO_COLOR` is unset.
- `--color=auto|always|never` (and honor `NO_COLOR`; `FORCE_COLOR` optional).
- Color **tokens** in PRESENCE and CHANGE cells (and overall key/value
  lines), not entire rows.

Sketch palette (gita-adjacent, still words not symbols):

| Token | Color |
|-------|--------|
| `present`, `all-present`, `clean` | green |
| `partial`, `missing`, `behind:N`, `no-upstream` | yellow |
| `dirty`, `untracked`, `invalid`, `error` | red |
| `ahead:N` | magenta |
| `sync-unknown`, `none`, `-` | dim / default |

**Tabwriter trap:** ANSI sequences inflate width and break columns. Compute
plain-string padding first, wrap color after, or bypass tabwriter for those
cells. Tests compare uncolored output (`--color=never` or non-TTY).

### 6. Code layout

Keep packages small:

- Matcher: `internal/statusfilter` (or under `internal/project`) — pure
  functions on `api.ProjectStatus` / `api.ReplicaStatus`; unit-test without
  cobra.
- Printers stay in `internal/cli`; color only there.
- Do not grow `internal/api` kinds for filters.
- Env: only if `--color` needs a `YERK__` override; otherwise `NO_COLOR` is
  enough (ADR 005 if a yerk-specific knob is added).

### 7. Docs

When behavior ships (not before):

- Domain spine status scopes table
- [status-model.md](../../../../docs/explanation/status-model.md)
- [check-status.md](../../../../docs/how-to/check-status.md)
- [commands.md](../../../../docs/reference/commands.md)
- `yerk status --help`

Do not document `--status=` syntax until it exists.

---

## Phases

### Phase 0 — Scope

Lock Q1–Q10 in [decisions.md](./decisions.md). Especially: flag set, AND
semantics, project vs replica grain, `--replicas` shape, color policy,
`--status=` defer.

**Exit:** enough to write the ADR without re-asking mid-draft.

### Phase 1 — ADR

Draft next ADR under `design/decisions/`:

- Status atoms and convenience flags
- AND combine; catalog selectors remain orthogonal
- Filter grain per invocation
- `--replicas` opt-in flatten
- Color auto / `NO_COLOR` / no json color; tabwriter rule
- Explicit defer of `--status=` to tag-boolean-select engine
- Related: 010, 011, 019, 022, domain spine

Operator review → accepted or revise. No cobra flags in this phase.

### Phase 2 — Convenience filters

Implement matcher + flags on `yerk status`. Tests: temp dirs only (ADR 007).
Empty match ok. `--presence-only` vs change flags errors. Preamble
`filter.status=…`. json/yaml filtered lists.

### Phase 3 — Replica listing

`--replicas` flatten (or locked alternative). Reuse replica table printer.
Filters apply at replica grain. Tests for multi-project live replicas +
tag ∩ dirty.

### Phase 4 — Color

TTY auto + `--color`. Palette on presence/change tokens. Alignment tests
with `--color=always` vs never. No deps unless Q6 overturns.

### Phase 5 — Docs + close

Update spine / status-model / how-to / commands / help. Owner todo: leave
`--status=` as backlog line pointing at tag-boolean-select (or a future
`status-expr-implement` plan). Remove this folder when shipped.

---

## Success criteria

- [ ] ADR accepted (or explicit reject/defer with rationale).
- [ ] `yerk status --dirty` (and locked siblings) shows only matching rows.
- [ ] Cross-project replica listing exists and is documented.
- [ ] Human status uses color on TTY; `--output json` stays plain.
- [ ] `--status=` not stubbed; defer recorded.
- [ ] Tests green (`mise run test` / `check`); plan removable at close.

## Immediate next step

When In Progress: Phase 0 — operator answers (or locks proposed answers) in
[decisions.md](./decisions.md).
