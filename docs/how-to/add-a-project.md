# How to add a project to the catalog

Status: **stub**

## Goal

Add or update a `[[projects]]` entry so `yerk` knows a project’s name, remote,
and optional labels.

## Prerequisites

- Familiarity with the config directory layout (see tutorial or reference)
- Ability to edit TOML under the yerk config dir (or `YERK__CATALOG`)

## Steps (to write)

1. Open `catalog.toml` (or the path from `YERK__CATALOG`)
2. Add a `[[projects]]` table (`name`, `remote`, optional `domain`, `tags`, …)
3. Confirm with `yerk status` / `yerk path <name>`

## See also

- [Catalog fields (reference)](../reference/catalog.md)
- [Tutorial: first catalog and status](../tutorials/first-catalog-and-status.md)
