# Checklist — materialize-dwim

## Phase 0 — Policy

- [ ] Q1–Q8 discussed / locked enough to draft ([decisions.md](./decisions.md))
- [ ] Tentative bundle accepted, flipped, or items deferred with defaults
- [ ] Non-goals still true (no clone→worktree conversion; no fetch in v1)

## Phase 1 — ADR

- [ ] ADR 025 drafted under `design/decisions/`
- [ ] ADR 016 consequences / related links updated (create vs materialize)
- [ ] Domain spine verb table updated
- [ ] Operator accept, revise, or reject

## Phase 2 — Implement

- [ ] `Resolver.Materialize` (or equivalent) owns DWIM; CLI stays thin
- [ ] Main present + named replica missing → worktree, not clone
- [ ] Force-clone hurdle (locked Q2)
- [ ] Catalog `replica_method` if Q3 yes
- [ ] Missing-main path matches Q1
- [ ] Branch rule matches Q5 (gitcmd if new origin-ref attach is needed)
- [ ] Tests as in [plan.md](./plan.md) Phase 2
- [ ] `mise run test` / `check` green

## Phase 3 — Docs

- [ ] `materialize` command Long/Short
- [ ] How-tos: materialize, replica create
- [ ] Explanation workspace-and-replicas; README Use; commands reference
- [ ] No leftover “materialize always clones” as the whole story

## Close-out

- [ ] Success criteria checked
- [ ] Delete this plan folder
- [ ] Update [../todo.md](../todo.md)
