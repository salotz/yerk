# Reference: Catalog

Status: **stub**

## File

`catalog.toml` under the yerk config directory (or `YERK__CATALOG`).

Missing file ⇒ empty catalog (not an error).

## `[[projects]]` fields (MVP)

| Field | Required | Meaning |
| --- | --- | --- |
| `name` | yes | Catalog key / short identity |
| `remote` | yes | Clone URI (git URL or path) |
| `domain` | no | Namespace label only (not a path map yet) |
| `tags` | no | Free-form labels for bulk select |
| `default_replica` | no | Distinguisher override; else remote HEAD / `main` |
| `path` | no | Fixed absolute replica path override |

## Example shape

```toml
[[projects]]
name = "yerk"
domain = "personal"
remote = "git@github.com:salotz/yerk.git"
tags = ["devel"]
default_replica = "main"
```

## See also

- [How to add a project](../how-to/add-a-project.md)
- [Configuration reference](./configuration.md)
- [ADR 004](../../design/decisions/004-config-and-catalog-split.md)
