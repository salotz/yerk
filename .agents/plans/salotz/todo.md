# salotz plans

Owner index for `.agents/plans/salotz/`.
Work-process: agent-guidelines personal `work-process.md` (In Progress / Backlog only).

## In Progress

### initial-feature-series

Next product feature series after the first vertical slice: placement policy,
identifiers/URIs, get/lookup, resolve config, replica create, agent context,
styles. Plan: [./initial-feature-series/](./initial-feature-series/).

## Backlog

### Parallel change probes

Serial status probes are enough for now. Later: errgroup/semaphore parallel
probes + deterministic fake git adapter tests. Came out of initial-feature-series
(held by operator). Soft dep: stable serial path and api status resources (already
on main).
