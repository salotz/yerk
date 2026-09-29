# How to check presence and change status

Status: **stub**

## Goal

List cataloged projects with **presence** on disk and, when useful, git
**change** flags via `yerk status` and `yerk status --git`.

## Prerequisites

- Non-empty catalog (or acceptance of empty output)
- For `--git`: present replicas and `git` available

## Steps (to write)

1. `yerk status` — name, domain, replica, presence, path
2. Select by tag: `yerk status --tag <name>`
   - `<name>` must be in the catalog root `tags` list (unknown tag → error)
   - Output includes `filter.tag=<name>` when filtering
   - Declared tag with no projects → “No projects matched tag …”
3. `yerk status --git` — add change probes for present replicas
4. Read presence vs change columns without conflating them

## See also

- [Status model (explanation)](../explanation/status-model.md)
- [Commands reference](../reference/commands.md)
