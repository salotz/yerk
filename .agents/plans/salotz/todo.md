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

### Domain model glossary

Add a durable **glossary** of product nouns (project, catalog, domain, workspace,
replica, presence/change, identifier/`yerk://`, materialize, …) for operators and
agents. Prefer a single canonical page (e.g. under `docs/explanation/` or
`design/architecture/`) and thin cross-links from concepts/identifiers stubs.
Suggested folder once spawned: `domain-model-glossary`. Soft dep: current
vocabulary is still concentrated in `design/domain-and-near-term.md` (pre-rename).

### Domain spine → architecture layout

Rename/move the design spine out of the catch-all
`design/domain-and-near-term.md` name into a clearer **architecture** home
(e.g. `design/architecture/domain-model.md` or similar; drop “near-term” from
the title once the first slice is historical). Update AGENTS.md, ADR Related
links, docs hubs, and plan pointers that cite the old path. Soft dep: may
bundle with glossary so one pass rewires links. Suggested folder:
`domain-spine-architecture-move`.

### Dead documentation cleanup (+ agent/role)

Inventory and remove or rewrite **stale/dead docs**: outdated command names
(`clone` leftovers), superseded path/identity prose, empty Diátaxis stubs that
lie, and links into deleted plan paths. Prefer a small reusable **agent role /
recipe** (checklist: link crawl, command-help vs docs drift, ADR status vs
text) so future cleanups are not one-off chat. Safety: preview-only deletes
until operator accepts; no host-private paths. Suggested folder:
`docs-deadwood-cleanup-role`.

