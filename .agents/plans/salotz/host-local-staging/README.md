# Plan index — host-local-staging

Ephemeral plan: **design only** for applying host-local (untracked) configuration
into project workspaces and replicas. **No CLI stubs** until an ADR is accepted
and a later implement plan starts.

Sparked by operator idea (2026-10-07): worktrees keep shipping without DVC
`config.local` (and similar) because yerk does not yet stage/inject locals.
Domain spine already names the internal verb (“stage locals”) as out of
near-term; this plan promotes it.

Same session follow-on (2026-10-07): **Git LFS skip-smudge** is host-local
configuration too, but it is **git config** (`git lfs install --local
--skip-smudge` in `.git/config`), not a working-tree file. Shared across
worktrees. See plan.md “Git LFS skip-smudge” and Q12.

| File | Role |
|------|------|
| [plan.md](./plan.md) | Goals, iterated design, phases, exit criteria |
| [decisions.md](./decisions.md) | **Only** Q&A inbox (`Q*`) |
| [checklist.md](./checklist.md) | Execution track |
| [../todo.md](../todo.md) | Owner index |

## Phase order

| # | Phase | Status |
|---|--------|--------|
| 0 | Scope + open questions | in progress |
| 1 | ADR draft + review | pending |
| 2 | Docs pointers; implement plan stub or explicit defer | pending |

## Protocol

- Design/docs only unless operator opens an implement plan.
- No git stage/commit unless asked.
- Never cite `Q*` in product code.
- Q&A only in [decisions.md](./decisions.md) (work-process).
