# Identifiers vs paths

Status: **stub**

## Why two concepts

A **project** or **replica** has a stable **identity** independent of where it
sits on disk. Workspace **style** can change path math; the id does not.

| Concept | Example | Role |
| --- | --- | --- |
| Bare id | `personal/yerk`, `personal/yerk/main` | Human/CLI shorthand |
| Canonical URI | `yerk://personal/yerk/main` | Machine currency (API, agents) |
| Filesystem path | `…/tree/personal/devel/yerk/main` | Layout result of style + config |

## Forms (ADR 012)

- **Two spellings only:** bare id and `yerk://…`. No `yerk:…` opaque form.
- Catalog row identity: required **`domain`** + **`name`** → bare `domain/name`.
- Unique **short** names (`yerk`, `yerk/main`) expand when unambiguous.
- Ambiguous short names **error** with candidates listed.
- A project id does **not** auto-mean the default replica; replica needs an
  explicit third segment or second CLI argument.

## Commands

`status`, `path`, `get`, `workspace ensure`, and `materialize` accept id forms
(and legacy “name + optional replica” positionals where applicable).

**Forward get** (id → resource):

```sh
yerk get personal/yerk              # project info
yerk get personal/yerk/main         # replica info
yerk project get personal/yerk
yerk replica get personal/yerk main
```

**Reverse lookup** (path → id + resource; walk-up under workspace/replica):

```sh
yerk lookup .
yerk lookup /path/under/replica
yerk project lookup /path/under/workspace
yerk replica lookup /path/under/replica
```

Machine-readable: add `--output json` (single `ProjectInfo` or `ReplicaInfo`
document, `apiVersion: yerk/v1`). See [ADR 015](../../design/decisions/015-get-and-lookup.md).

## See also

- [ADR 012](../../design/decisions/012-identifiers-and-yerk-uri.md)
- [ADR 015](../../design/decisions/015-get-and-lookup.md)
- [API resources (ADR 011)](../../design/decisions/011-api-resources.md)
- [Commands reference](../reference/commands.md)
