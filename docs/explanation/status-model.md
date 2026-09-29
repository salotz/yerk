# Explanation: Status model

Status: **stub**

## Intent

Explain why **presence** (disk vs expected path) and **change** (git probe)
are separate, and how to read `yerk status` without mixing those jobs.

## Presence (to expand)

| Value | Meaning |
| --- | --- |
| `missing` | Expected path does not exist |
| `present` | Usable git checkout at the path |
| `invalid` | Path exists but is not a usable checkout |

## Change (to expand)

- Only meaningful when presence is `present`
- Orthogonal flags from git probes (on by default; `--presence-only` skips)
- Serial probes first; parallelization is a later hardening detail

## Scopes

| Invocation | View |
| --- | --- |
| `yerk status` / `yerk status <project>` | **ProjectStatus** — workspace path + default-replica summary |
| `yerk status <project> <replica>` | **ReplicaStatus** — checkout path, presence, change, branch |

## Implementation note

Observed status is modeled as `internal/api` resources (`ReplicaStatus`,
`ProjectStatus`) — not CLI-private row structs ([ADR 011](../../design/decisions/011-api-resources.md)).
Human tables are one printer over those types.

## See also

- [How to check status](../how-to/check-status.md)
- [Commands reference](../reference/commands.md)
- [design/domain-and-near-term.md](../../design/domain-and-near-term.md) (status section)
- [ADR 011](../../design/decisions/011-api-resources.md)
