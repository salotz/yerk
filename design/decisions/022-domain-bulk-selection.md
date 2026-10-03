# 022. Domain bulk selection (`--domain`)

## Status

Accepted (2026-10-02)

## Context

Catalog identity is required `domain` + `name` (ADR 012). Operators often want
to act on every project in one logical domain (e.g. all `personal/*`) the same
way they already filter by declared tag (ADR 010) or opt into full-catalog
mutate with `--all`.

Until this ADR, bulk domain work meant listing bare ids (`personal/yerk
personal/other…`) or post-filtering `status --output json`. That is awkward for
`materialize` / `workspace ensure` and inconsistent with the tag selector.

Domains are **not** a closed vocabulary like tags: any non-empty string that
appears (or might appear) as `[[projects]].domain` is a valid filter key.
Unknown / empty-match domains are handled with the same read-vs-mutate empty
policy as tags.

## Decision

### Selector

Add `--domain <name>` as a bulk project selector:

| Command | Behavior |
| --- | --- |
| `yerk status --domain <name>` | List projects with `domain == <name>`. Empty match → message, exit 0. Prints `filter.domain=<name>` on human output. |
| `yerk materialize --domain <name>` | Materialize default (or `--replica`) for each match. Empty match → **error**. |
| `yerk workspace ensure --domain <name>` | Ensure workspace dirs for each match. Empty match → **error**. |

`<name>` is trimmed; empty `--domain` is rejected as “domain required” via
`Catalog.SelectByDomain`.

### Mutual exclusion

Mutate bulk selectors remain **one mode**:

- project id args **XOR**
- `--all` **XOR**
- `--tag <name>` **XOR**
- `--domain <name>`

`status` likewise: project args **XOR** `--tag` **XOR** `--domain` (default
still = full catalog). Combining any two is an error.

### Implementation

- `Catalog.FilterDomain` / `Catalog.SelectByDomain` (shared path; domain is not
  validated against a root list).
- `resolveBulkProjectSelection` and materialize selection take `domain` beside
  `tag` / `all`.
- No intersection of tag∩domain in one flag set (use tags on projects, or name
  ids, if a finer set is needed).

### Not in this decision

- Multi-domain flags (`--domain a --domain b`) or domain globs.
- Requiring `[domains]` host roots to exist for a domain filter (placement still
  fails later if a selected project cannot resolve a path).
- Closed domain vocabulary or catalog root `domains = […]`.
- `state update --domain` (still names / `--all` only unless extended later).

## Consequences

- Operator README and command help document `--domain` next to `--tag` / `--all`.
- ADR 010 tag rules unchanged; domain remains orthogonal to tags.
- Domain spine bulk table lists `--domain` as a first-class selector.

## Related

- [010](./010-catalog-tag-vocabulary.md) — tag bulk select (closed vocabulary)
- [012](./012-identifiers-and-yerk-uri.md) — domain as identity segment
- [009](./009-workspace-subcommand-and-ensure-scope.md) — ensure bulk
- [domain-and-near-term.md](../domain-and-near-term.md) — Project selection (bulk)
