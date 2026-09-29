# Explanation: Concepts

Status: **stub**

## Intent

Introduce the nouns operators meet in help text and status output, without
turning into a procedure or a full design doc.

## Terms (to expand)

- **Project** — named unit in the catalog
- **Catalog** — host registry of projects (`catalog.toml`)
- **Config (tool)** — how `yerk` runs on this host (`config.toml`)
- **Remote** — canonical VCS URI for a project
- **Domain** — namespace label (not yet a filesystem map)
- **Tag** — declared bulk-select label (catalog root vocabulary; ADR 010)
- **Workspace** — root + layout **policy** for placing replicas
- **Replica** — one concrete on-disk checkout on this host
- **Presence / change** — see [status model](./status-model.md)

## Avoid conflating

- Project ≠ single directory
- Replica ≠ git branch
- Workspace ≠ one checkout path
- Tag ≠ domain
- `yerk` ≠ PRJX

## See also

- [Workspace and replicas](./workspace-and-replicas.md)
- [yerk and PRJX](./yerk-and-prjx.md)
- Maintainer spine: [design/domain-and-near-term.md](../../design/domain-and-near-term.md)
