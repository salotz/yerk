# How to get project or replica info

Status: **draft**

## Forward (by identifier)

```sh
yerk get personal/yerk
yerk get personal/yerk/main
yerk get yerk://personal/yerk/main
yerk project get yerk          # short unique name
yerk replica get yerk main
```

A **project** id never means the default replica. Add the replica segment (or
second argument) when you need replica info.

## Reverse (by path)

Default `lookup` prints only the canonical URI (path → id). Expand with `get`
when you need the full resource:

```sh
yerk lookup .
# → yerk://personal/yerk/main

yerk lookup path/to/file/under/checkout
yerk project lookup /path/under/workspace   # → yerk://personal/yerk
yerk replica lookup /path/under/replica     # → yerk://personal/yerk/main

yerk get "$(yerk lookup .)"                 # human key/value for that URI
```

Any subdirectory under a known workspace or replica root matches (walk-up).
`replica lookup` errors if the path is only under the project workspace.

## Structured output for agents

```sh
yerk get personal/yerk --output json
yerk get personal/yerk --output yaml
yerk lookup . --output json                 # full ReplicaInfo / ProjectInfo
yerk get "$(yerk lookup .)" --output json   # same via id
```

Documents are `ProjectInfo` or `ReplicaInfo` with `apiVersion: yerk/v1`
(ADR 015). On `get`, `--output table` keeps the human key/value layout; on
`lookup`, default/`table` stay URI-only. Shared flag surface: ADR 019.

## See also

- [Identifiers](../explanation/identifiers.md)
- [Commands](../reference/commands.md)
- [ADR 015](../../design/decisions/015-get-and-lookup.md)
- [ADR 019](../../design/decisions/019-output-formats.md)
