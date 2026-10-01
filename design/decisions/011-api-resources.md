# 011. Model-driven API resources

## Status

Accepted (2026-09-28)

## Context

Near-term commands load TOML into `internal/config` structs, join layout /
presence / git in `internal/project`, and print ad-hoc table rows from the
CLI (`text/tabwriter` over `project.Row`). That works for the first slice but
locks shapes in the wrong layer:

- TOML load structs are storage/schema for files, not a product API.
- CLI row DTOs are presentation; they should not own identity or probe fields.
- Planned follow-ons need one shared model: `--output json|yaml|table`,
  JSON Schema / type docs, universal `yerk get <kind>`, and later backends
  without rewriting cobra `RunE` logic.

Domain language already distinguishes **project** vs **replica** status
scopes ([domain-and-near-term.md](../domain-and-near-term.md)). This ADR
was the model pilot before status UX rewrite landed.

## Decision

### Package

Introduce **`internal/api`**: stable, versioned **resource types** shared by
collectors, printers, and future serializers.

- Not TOML file structs (`internal/config` stays the load/validate layer).
- Not CLI formatting (cobra and tabwriter stay in `internal/cli`).
- Name `api` (not `resource`) to match “API resources” / kubectl-style
  mental model without implying HTTP yet.

### Object shape

Resources are plain Go structs with:

- `APIVersion` — currently `yerk/v1`
- `Kind` — type name string (`Project`, `Catalog`, `ProjectStatus`,
  `ReplicaStatus`, …)
- Documented exported fields and `json` tags (yaml can reuse json names
  later)

No full Kubernetes runtime.Object / scheme machinery. Keep the package
small; add list wrappers only when a printer or `get` needs them.

### Initial kinds

| Kind | Role |
| --- | --- |
| `Project` | Registry object: catalog identity and desired placement fields |
| `Catalog` | Tag vocabulary + list of `Project` |
| `ReplicaStatus` | Observed state for one replica (path, presence, change, branch) |
| `ProjectStatus` | Project-scoped view (workspace path + optional default-replica summary) |

Load adapters **map into** these types (`config.Project` → `api.Project`,
status collection → `ReplicaStatus` / later `ProjectStatus`). Commands
**print** resources (or thin views derived only for column layout).

### Ownership boundaries

```text
config.Load*     → config structs (TOML)
       ↓ map
api.* resources  ← project.Resolver / collectors fill observed fields
       ↓ print
cli (table today; json|yaml later)
```

Business rules (default replica, path math, presence, git probe) stay out of
cobra `RunE` beyond flag parse, selection, and print.

### Out of scope for this ADR (follow-ons)

- `--output json|yaml|table` flag surface
- Generated or hand JSON Schema / OpenAPI
- `yerk get <resource> [name…]`
- Non-file backends
- Changing default `yerk status` columns (that is P6; types here must not
  block project- vs replica-scoped views)
- Identifier / `yerk://` grammar — see [012](./012-identifiers-and-yerk-uri.md)
  (`uri` field on resources)

## Consequences

- Status collection returns `[]api.ReplicaStatus` (today’s default-replica
  probe) instead of a CLI-private `Row` type.
- `ProjectStatus` exists for P6 project list UX; default human table may
  still print replica-oriented columns until P6 lands.
- New observed fields prefer landing on api types first, then printers.
- ADR 004 / 010 remain authority for catalog file shape; api `Project` is
  the in-process product view of a registry row, not a second file schema.

## Related

- [domain-and-near-term.md](../domain-and-near-term.md) — Explicit resources; status scopes
- [004](./004-config-and-catalog-split.md) — config vs catalog files
- [010](./010-catalog-tag-vocabulary.md) — closed tags on the catalog document
