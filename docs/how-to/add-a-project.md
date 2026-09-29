# How to add a project to the catalog

Status: **stub**

## Goal

Add or update a `[[projects]]` entry so `yerk` knows a project’s name, remote,
domain, and workspace path.

## Prerequisites

- Familiarity with the config directory layout (see tutorial or reference)
- Domain root already set in `config.toml` `[domains]` for the project’s domain
- Ability to edit TOML under the yerk config dir (or `YERK__CATALOG`)

## Steps (to write)

1. Ensure `config.toml` has a `[domains]` entry for the project’s domain
   (host-absolute root, e.g. `personal = "~/tree/personal"`).
2. Open `catalog.toml` (or the path from `YERK__CATALOG`).
3. Ensure any labels you need appear in the top-level `tags = […]` list
   (closed vocabulary; ADR 010).
4. Add a `[[projects]]` table: `name`, `remote`, `domain`, relative `path`
   (project workspace under the domain root, e.g. `devel/yerk`), optional
   `tags` (must be members of the catalog `tags` list), `default_replica`.
5. Confirm with `yerk status` / `yerk path <name>` (workspace) /
   `yerk path <name> <replica>`.
6. Optional: `yerk workspace ensure <name>` to create the project workspace
   directory (not a replica checkout).

## See also

- [Catalog fields (reference)](../reference/catalog.md)
- [Configuration reference](../reference/configuration.md)
- [ADR 008](../../design/decisions/008-domain-roots-and-relative-catalog-paths.md),
  [ADR 010](../../design/decisions/010-catalog-tag-vocabulary.md)
- [Tutorial: first catalog and status](../tutorials/first-catalog-and-status.md)
