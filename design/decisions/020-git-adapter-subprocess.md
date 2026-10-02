# 020. Git adapter: subprocess default (not go-git)

## Status

Accepted (2026-10-02) — docs-only lock of the existing design note

## Context

yerk talks to git through `internal/gitcmd` (`Runner`: clone, probe, worktree
add, branch ensure, …). The product already assumes a developer host with git
on `PATH`. Occasional discussion revisits **go-git** (or hybrid) for pure-Go
hosts or in-process control.

The tradeoff was already stated in
[domain-and-near-term.md](../domain-and-near-term.md); this ADR records it as a
durable decision so implementers do not re-litigate mid-feature.

## Decision

- **Default implementation = `git` CLI subprocess** behind the `Runner`
  interface.
- **Do not migrate to go-git** in the current feature series.
- Revisit only if a concrete need appears: hosts without git, process scaling
  after parallel probes, or impedance that subprocess cannot cover cleanly.

### Why subprocess (summary)

| | Subprocess `git` | go-git |
| --- | --- | --- |
| Behavior | Matches operator git (SSH agent, hooks, worktrees, credentials) | Subset / impedance; edge cases often shell out anyway |
| Debug | `GIT_TRACE`, familiar UX | Library-specific |
| Deps | Requires `git` on `PATH` | Pure Go; larger surface |
| Ops | Process overhead; parse porcelain carefully | In-process control |

yerk is an operator tool on a PRJX-style host where git is already required for
day-to-day work. Identical behavior beats eliminating the binary dependency.

### Adapter boundary

Keep **`internal/gitcmd.Runner`** as the seam. Call sites must not shell out
ad hoc. A future go-git or hybrid backend can implement the same interface
without rewriting CLI/project logic.

## Consequences

- CI and dogfood hosts need `git` on `PATH` (already true for materialize /
  replica create / status change probes).
- Parallel probe work (held backlog) stays on the subprocess adapter first.
- Docs and onboarding should say “requires git”, not “pure Go VCS”.

## Related

- [domain-and-near-term.md](../domain-and-near-term.md) — “Why git subprocess”
- [016](./016-replica-create.md) — worktree / clone via gitcmd
- [002](./002-go-standalone-binary.md) — Go binary product shape
