# How to add a project to the catalog

Status: **stub**

## Goal

Add or update a catalog `[[projects]]` entry so `yerk` knows a project’s name,
remote, and domain. Host placement comes from `config.toml` (ADR 014).

## Prerequisites

- Familiarity with the config directory layout (see tutorial or reference)
- Ability to edit TOML under the yerk config dir (or `YERK__CATALOG` /
  `YERK__CONFIG`)
- For the common case: a `[domains]` root already set for the project’s domain

## Steps (to write)

1. Ensure `config.toml` has a `[domains]` entry for the project’s domain when
   you want the default workspace `<root>/<name>` (host-absolute root, e.g.
   `personal = "~/tree/personal/devel"`). Skip this only if you will set a full
   host `[[projects]]` path instead.
2. Open `catalog.toml` (or the path from `YERK__CATALOG`).
3. Ensure any labels you need appear in the top-level `tags = […]` list
   (closed vocabulary; ADR 010).
4. Add a catalog `[[projects]]` table: `name`, `remote`, `domain`, optional
   `tags` (must be members of the catalog `tags` list), `default_replica`.
   **Do not** put a host workspace `path` in the catalog.
5. Only if placement is not `<domain-root>/<name>`, add a host row in
   `config.toml`:

   ```toml
   [[projects]]
   name = "odd"
   domain = "personal"
   path = "~/somewhere/else/odd"   # or relative under the domain root
   ```

6. Confirm with `yerk status` / `yerk path <name>` (workspace) /
   `yerk path <name> <replica>`.
7. Optional: `yerk workspace ensure <name>` to create the project workspace
   directory (not a replica checkout).

## See also

- [Catalog fields (reference)](../reference/catalog.md)
- [Configuration reference](../reference/configuration.md)
- [ADR 014](../../design/decisions/014-host-local-path-model.md),
  [ADR 010](../../design/decisions/010-catalog-tag-vocabulary.md)
- [Tutorial: first catalog and status](../tutorials/first-catalog-and-status.md)
