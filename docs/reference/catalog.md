# Reference: Catalog

Status: **stub**

## File

`catalog.toml` under the yerk config directory (or `YERK__CATALOG`).

Missing file ⇒ empty catalog (not an error).

## Top-level fields

| Field | Required | Meaning |
| --- | --- | --- |
| `tags` | no (empty ok) | Closed vocabulary of labels. Every project tag must be a member ([ADR 010](../../design/decisions/010-catalog-tag-vocabulary.md)). |
| `[[projects]]` | — | Project rows (see below). |

Duplicate or blank names in `tags` are errors. Load validates the whole file.

## `[[projects]]` fields (MVP)

| Field | Required | Meaning |
| --- | --- | --- |
| `name` | yes | Catalog key / short identity |
| `remote` | yes | Clone URI (git URL or path) |
| `path` | yes (near-term) | **Project workspace** (owns replicas). Prefer **relative** to `[domains.<domain>]`. Absolute allowed. Not a checkout. |
| `domain` | yes if path relative | Namespace + key into config `[domains]` |
| `tags` | no | Subset of the catalog root `tags` list (bulk select) |
| `default_replica` | no | Distinguisher override; else remote HEAD / `main` |

Resolve workspace, then style:

- relative `path` → `<domains[domain]>/<path>`
- `workspace-dir` → default replica at `<workspace>/<default_replica>`

Example: domain root `~/tree/personal`, path `devel/yerk`, replica `main` →
`~/tree/personal/devel/yerk/main`.

## Example shape

```toml
tags = ["devel", "work"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@github.com:salotz/yerk.git"
path = "devel/yerk"
tags = ["devel"]
default_replica = "main"
```

## See also

- [How to add a project](../how-to/add-a-project.md)
- [Configuration reference](./configuration.md)
- [ADR 004](../../design/decisions/004-config-and-catalog-split.md), [ADR 008](../../design/decisions/008-domain-roots-and-relative-catalog-paths.md), [ADR 010](../../design/decisions/010-catalog-tag-vocabulary.md)
