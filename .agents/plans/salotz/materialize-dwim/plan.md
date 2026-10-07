# Plan: materialize as DWIM replica ensure

## Context

Incident (2026-10-05): project `yerk://examol/darpa-nodes` already lived in a
`workspace-dir` layout (`~/tree/examol/devel/darpa-nodes/main` present as a
usable checkout, **not** seeded by yerk). Operator added the catalog row and
ran a materialize of a session replica (`md-next-workflow`). yerk did
`git clone --branch md-next-workflow` into a sibling directory.

On disk after the fact:

| Path | Kind |
| --- | --- |
| `main/` | Real repo (`.git` directory); hub |
| `md-disulfide-normalization/`, `cd1-2_sequences/` | Worktrees of `main` |
| `md-next-workflow/` | Independent clone (own `.git` dir; not in `git worktree list`) |

This was **specified** behavior, not a path-math bug:

- ADR 016: `materialize` = bootstrap from catalog **remote** (always clone).
- `yerk replica create` = session spin-out; default method **worktree** from
  main. Main presence does not require yerk to have created it.
- `materialize <project> [replica]` still accepts a non-main distinguisher and
  never consults main.

Operator intent: **`materialize` is the main path** — declarative “ensure this
replica exists given project policy and what is already on disk.” A fresh
independent clone should remain possible but require an explicit hurdle.

Chosen direction vs earlier options: **option 2** (auto-worktree when main is
present), not refuse-and-redirect (option 1) and not “named replica forbidden
on materialize” (option 3).

Design spine: [design/domain-and-near-term.md](../../../../design/domain-and-near-term.md).
Current split: [ADR 016](../../../../design/decisions/016-replica-create.md).
CLI: `internal/cli` `materializeOne` → `gitcmd.Clone` only.
Create path: `project.Resolver.CreateReplica`.

---

## Goals

1. `yerk materialize <project> <replica>` **ensures** the replica using
   placement + presence + catalog method policy (DWIM).
2. When main is **present**, a missing named replica is a **worktree** of that
   hub (unless force-clone / catalog `replica_method = clone`).
3. Independent clone remains available behind an explicit hurdle.
4. Durable ADR (025, amends 016) so help/how-tos stop teaching “materialize
   always clones; create is the worktree verb.”
5. Tests: existing main (even if yerk did not create it) → materialize named
   replica must **not** call `Clone`.

## Non-goals

- Converting an already-present independent clone into a worktree (already
  present stays success / no-op).
- Host cleanup of the accidental `md-next-workflow` clone (operator; optional).
- Fetch-then-attach for remote-only branches (v1 uses local refs only, same as
  status; Q5).
- Removing `yerk replica create` in this plan (Q4; keep unless locked otherwise).
- `pull` / `push` (sync-verbs-design).

---

## Phase 0 — Policy

Lock Q1–Q8 in [decisions.md](./decisions.md). Tentative bundle already offered
in session; not operator-locked.

**Exit:** enough to write ADR 025 without re-asking mid-draft.

---

## Phase 1 — ADR

Draft **ADR 025** (materialize DWIM / ensure). Amend ADR 016 (create remains
fail-if-exists + explicit method; materialize is the ensure verb). Update
related links from domain spine verb table, how-tos, command help.

Operator accept → implement. Reject/revise stays in this folder.

---

## Phase 2 — Implement

Likely shape (adjust to locked Qs):

```text
CLI materializeOne
  → project.Resolver.Materialize(ctx, project, replica, opts)
       dest present  → bind + already-present
       dest invalid  → error
       method clone (CLI / catalog) → git clone (today)
       replica is main / default    → git clone (bootstrap hub)
       main present                 → WorktreeAdd (branch rule Q5)
       main missing                 → Q1 (error | clone feat | clone main then worktree)
```

Share git worktree/clone with `CreateReplica`; do not duplicate branch logic
in the CLI package.

Tests (temp dirs only; ADR 007):

- Main present, named replica missing → `WorktreeAdd`, no `Clone`.
- Main missing, named replica → locked Q1 behavior.
- `--method clone` / `--clone` → `Clone` even if main present.
- Catalog `replica_method = clone` → clone unless CLI overrides.
- Dest already present → success, no git.
- `replica create` still refuse-if-present; default worktree unchanged.

---

## Phase 3 — Docs

- `yerk materialize --help` (no “always clone”).
- How-tos: [materialize-a-replica.md](../../../../docs/how-to/materialize-a-replica.md),
  [create-a-replica.md](../../../../docs/how-to/create-a-replica.md).
- Explanation [workspace-and-replicas.md](../../../../docs/explanation/workspace-and-replicas.md).
- README “Use” steps; [commands.md](../../../../docs/reference/commands.md).
- Domain spine verb table.

---

## Success criteria

- [ ] Q1–Q8 locked or explicitly deferred with default.
- [ ] ADR 025 accepted.
- [ ] Named materialize with present main does not clone.
- [ ] Force-clone hurdle works.
- [ ] Tests green (`mise run test` / `check`).
- [ ] Help/how-tos match shipped behavior.
- [ ] Plan folder removable at close.

## Immediate next step

Phase 0: operator lock or flip the tentative bundle in [decisions.md](./decisions.md).
Do not start ADR/code until then.
