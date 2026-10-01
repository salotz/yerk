# Explanation

**Understanding-oriented** (Diátaxis).

Background, vocabulary, and design rationale so readers can see **why**
`yerk` behaves as it does. Not step-by-step procedures and not a flag catalog.

Publication remains ad hoc plain Markdown
([ADR 006](../../design/decisions/006-ad-hoc-docs-diataxis.md)).

Deeper maintainer design lives under [`design/`](../../design/); these pages
should stay operator-facing and link out for ADR-level detail.

## Pages

| Page | Status | Intent |
| --- | --- | --- |
| [Concepts](./concepts.md) | Stub | Project, catalog, workspace, replica, domain, tag |
| [Identifiers](./identifiers.md) | Stub | Bare id / `yerk://` vs filesystem path |
| [Workspace and replicas](./workspace-and-replicas.md) | Stub | Layout policy vs on-disk checkouts |
| [Status model](./status-model.md) | Stub | Presence vs change; why they are separate |
| [yerk and PRJX](./yerk-and-prjx.md) | Stub | Tool vs spec; shared vocabulary |
