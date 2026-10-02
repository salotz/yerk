# Reference: Catalog

Status: **stub**

## Layout

| Item | Notes |
| --- | --- |
| File | `$XDG_CONFIG_HOME/yerk/catalog.toml` (or `YERK__CATALOG`) |
| Role | Portable **registry** of projects on this host’s mental model |
| Not owned here | Host workspace paths (see [configuration](./configuration.md), ADR 014) |

## Root fields

| Field | Required | Meaning |
| --- | --- | --- |
| `tags` | no | Closed vocabulary; project `tags` must be members (ADR 010) |

## `[[projects]]` fields

| Field | Required | Meaning |
| --- | --- | --- |
| `name` | yes | Short name; bare id `domain/name` |
| `domain` | yes | Logical namespace (ids/URIs; ADR 012). Optional host `[domains]` root uses the same string |
| `remote` | yes* | Clone URI (*required for materialize) |
| `tags` | no | Subset of root `tags` |
| `default_replica` | no | Default distinguisher when ops omit one |
| `workspace_style` | no | Ambient per-project style override |
| `replica_method` | no | Later: replica create method |

**No `path` field.** Host workspace paths come from `config.toml`:

- optional `[domains.<domain>]` → default `<root>/<name>`
- optional `[[projects]]` host row for overrides

A legacy catalog `path` key is a **load error**.

```toml
tags = ["devel"]

[[projects]]
name = "example"
domain = "personal"
remote = "git@github.com:example/example.git"
tags = ["devel"]
# default_replica = "main"
# workspace_style = "workspace-dir"
# workspace_style = "name-tags"
# workspace_style = { style = "name-tags", main_dir = "~/.app", replica_dir = "~/tree/…/app" }
# replica_method = "worktree"
```

Host counterpart (paths stay out of the catalog):

```toml
# config.toml
[domains]
personal = "~/tree/personal/devel"
# personal/example → ~/tree/personal/devel/example
```

## See also

- [Configuration reference](./configuration.md)
- [ADR 010](../../design/decisions/010-catalog-tag-vocabulary.md),
  [ADR 012](../../design/decisions/012-identifiers-and-yerk-uri.md),
  [ADR 014](../../design/decisions/014-host-local-path-model.md),
  [ADR 016](../../design/decisions/016-replica-create.md),
  [ADR 018](../../design/decisions/018-name-tags-workspace-style.md)
