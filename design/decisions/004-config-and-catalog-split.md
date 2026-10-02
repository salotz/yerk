# 004. Split tool config and project catalog

## Status

Accepted (2026-09-25)

## Context

ADR 003 placed the host catalog and workspace knobs in a single
`$XDG_CONFIG_HOME/yerk/config.toml`. Operator design clarified two different
lifetimes and edit patterns:

- **Tool/host behavior** (workspace style; later probe/output defaults)
  may change per machine and is about *how* `yerk` runs on this host.
- **Project registry** (name, remote, domain, tags, project-workspace path) is
  about *what* projects exist and is often edited by agents or shared mental
  model as a pure list.

A single file conflates those concerns and makes “generate/update the listing”
riskier (accidental edits to host paths).

## Decision

Two documents under the yerk config directory:

| File | Role |
| --- | --- |
| `config.toml` | Tool and host behavior only (`[workspace]`, future tool sections). |
| `catalog.toml` | Project catalog only (`[[projects]]` entries). |

Environment:

- `YERK__CONFIG` — explicit **tool** config file path (unchanged meaning).
- `YERK__CONFIG_DIR` — config directory containing both files by default.
- `YERK__CATALOG` — explicit **catalog** file path.

Missing `config.toml` ⇒ defaults. Missing `catalog.toml` ⇒ empty catalog
(not an error).

MVP catalog fields per project: `name`, `remote`, required `domain` (ADR 012),
`tags`, `default_replica`, optional style/method overrides. **No host workspace
`path`** in the catalog ([ADR 014](./014-host-local-path-model.md)).

Catalog root also declares a closed **tag vocabulary** (`tags = […]`); project
`tags` must be members of that list ([ADR 010](./010-catalog-tag-vocabulary.md)).

`domain` is an id namespace label. Optional host `[domains]` roots and
`[[projects]]` path rows live in tool config (ADR 014).

## Consequences

- Loaders become `LoadConfig` + `LoadCatalog` (names flexible in code).
- `yerk config show` should summarize tool config; catalog listing stays on
  `yerk status` / future `yerk catalog`.
- Example files and docs must show both documents.
- ADR 003 remains valid for XDG + `YERK__` prefix; this ADR narrows file roles.
- Existing single-file experiments should migrate `[[projects]]` out of
  `config.toml` into `catalog.toml`.
