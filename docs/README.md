# Documentation

User- and operator-facing documentation for `yerk`.

## Status

**Ad hoc plain Markdown** in-tree.
There is no doc site or publication pipeline yet; a better system for actual
publication is deferred (see [ADR 006](../design/decisions/006-ad-hoc-docs-diataxis.md)).

Content follows [Diátaxis](https://diataxis.fr/): four documentation types for
four reader needs. Prefer the matching type over one mixed page.

## Diátaxis map

| Type | Need | Directory |
| --- | --- | --- |
| [Tutorials](./tutorials/) | Learning — guided lessons | `docs/tutorials/` |
| [How-to guides](./how-to/) | Task — accomplish a concrete job | `docs/how-to/` |
| [Reference](./reference/) | Information — accurate lookup | `docs/reference/` |
| [Explanation](./explanation/) | Understanding — why and context | `docs/explanation/` |

Pages under each type are mostly **stubs** until content is written.

## Scope vs other trees

| Tree | Audience | Contents |
| --- | --- | --- |
| `docs/` | Operators and users of the CLI | Diátaxis user docs |
| [`design/`](../design/) | Maintainers | Goals, domain language, ADRs |
| [`contributing/`](../contributing/) | Contributors | Bootstrap, build, test, tooling |

## Flags and environment variables

| Layer | Role |
| --- | --- |
| [`.appinfo/meta.toml`](../.appinfo/meta.toml) | Static product + env registry (RFC 030/031) |
| `internal/envvars` + CLI help | Runtime help and live dumps (ADR 005) |

- `yerk --help` / `yerk <command> --help` — command surface (primary env subset)
- `yerk help envvars` — full environment variable reference
- `yerk envvars` — live values in this process

Repo reference pages should link these rather than fork a third full dump.

## Start here

Until tutorials and how-tos are filled in, use the root
[README](../README.md) quick start and configuration sections, then return
here as stubs gain content.
