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

```sh
yerk lookup .
yerk lookup path/to/file/under/checkout
yerk project lookup /path/under/workspace
yerk replica lookup /path/under/replica
```

Any subdirectory under a known workspace or replica root matches (walk-up).
`replica lookup` errors if the path is only under the project workspace.

## JSON for agents

```sh
yerk get personal/yerk --output json
yerk lookup . --output json
```

Documents are `ProjectInfo` or `ReplicaInfo` with `apiVersion: yerk/v1`
(ADR 015).

## See also

- [Identifiers](../explanation/identifiers.md)
- [Commands](../reference/commands.md)
- [ADR 015](../../design/decisions/015-get-and-lookup.md)
