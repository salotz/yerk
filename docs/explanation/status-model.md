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
- Orthogonal flags from git probes (`--git`), not a single exclusive enum
- Serial probes first; parallelization is a later hardening detail

## See also

- [How to check status](../how-to/check-status.md)
- [Commands reference](../reference/commands.md)
- [design/domain-and-near-term.md](../../design/domain-and-near-term.md) (status section)
