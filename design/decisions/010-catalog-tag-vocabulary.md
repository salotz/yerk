# 010. Closed catalog tag vocabulary

## Status

Accepted (2026-09-28)

## Context

Catalog tags were free-form strings on each `[[projects]]` row. That allowed
typos (`devel` vs `dev`), one-off labels, and `yerk status --tag` filters that
silently matched nothing. Operators want a single place to define the host’s
label set and to require every project tag to be one of those labels.

Tags remain orthogonal to **domain** (namespace / domain-root key). They are
only for bulk select and grouping.

## Decision

### Top-level declaration

`catalog.toml` declares the vocabulary once at the document root:

```toml
tags = ["devel", "work"]

[[projects]]
name = "yerk"
…
tags = ["devel"]
```

- `tags` at catalog root is the **closed set** of allowed labels.
- Each project’s `tags` array is optional, but every entry **must** be a
  member of the root list (exact string match after rejecting blank/space-padded
  names).
- Duplicate names in the root list or on a single project are errors.
- An empty root list is valid: then no project may list any tags.
- Missing catalog file still yields an empty catalog (not an error). A catalog
  file with projects is validated on load.

### Load-time validation

`LoadCatalog` runs `Catalog.Validate` after TOML parse. Invalid tags fail the
load with a path-qualified error so every command that reads the catalog
surfaces the problem.

### CLI selection

Bulk selection by tag uses the closed vocabulary:

- `yerk status --tag <name>` — requires `<name>` declared. Unknown tags
  error. Declared tag with zero projects is an empty match (not an error).
  When filtering, status prints `filter.tag=<name>`.
- `yerk clone --tag <name>` — same declared-tag rule; empty match is an
  **error** (mutate). Mutually exclusive with project args and `--all`.
- Same `Catalog.SelectByTag` helper is the shared path for bulk ops
  (`workspace ensure --tag` still pending).

`yerk catalog show` prints the declared vocabulary, then the project table
(TAGS column still shows each project’s tags).

### Not in this decision

- No separate `yerk tags` command or mutation UX (hand-edit catalog).
- No implied hierarchy, colors, or tag metadata beyond the name string.
- Domains stay in tool `config.toml` `[domains]`; tags stay in the catalog.

## Consequences

- Existing catalogs that used free-form project tags without a root `tags`
  list must add the declaration (or drop project tags) before commands load.
- Examples and docs show `tags = […]` at the top; sample project tags are
  optional and commented until the operator fills the vocabulary.
- ADR 004 still owns config/catalog file split; this ADR narrows tag semantics
  inside the catalog document.

## Related

- [domain-and-near-term.md](../domain-and-near-term.md) — Tag noun; **Project selection (bulk)** table
- [004](./004-config-and-catalog-split.md) — catalog document role
- [docs/reference/catalog.md](../../docs/reference/catalog.md)
