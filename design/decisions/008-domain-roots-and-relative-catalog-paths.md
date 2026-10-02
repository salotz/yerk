# 008. Domain roots and relative catalog paths

## Status

**Superseded for path resolution** by [ADR 014](./014-host-local-path-model.md)
(2026-10-01). Domain remains an id namespace (ADR 012). ADR 014 keeps optional
`[domains]` as a **host default** (`<root>/<name>`), but catalog no longer
carries `path`; overrides live on host `[[projects]]`.

Accepted (2026-09-28) for the original domain-root model.

## Context

Catalog `path` was an absolute **project workspace** directory so style math
could place replicas (`…/yerk` → `…/yerk/main`). Absolute paths make the
catalog host-specific: the same project list cannot move between machines
without rewriting every row.

Host layout already groups work under **domains** (RFC 25 / PRJX vocabulary):
e.g. `~/tree/personal`, with projects under segments such as `devel/yerk`.
Catalog `domain` was only a namespace **label**; it did not participate in
path resolution. Operators asked to reintroduce relative paths by binding
each domain name to a host-absolute root in tool config.

## Decision

### Tool config (`config.toml`) owns host domain roots

```toml
[workspace]
style = "workspace-dir"

[domains]
personal = "~/tree/personal"
# examol = "/other/host/path/for/examol"
```

- Keys are domain names (same strings as catalog `domain`).
- Values are **host-absolute** roots (leading `~/` expanded to the process
  user home).
- Missing or empty domain for a project that uses a relative path is an error
  at resolve time.

### Catalog (`catalog.toml`) prefers portable relative paths

```toml
[[projects]]
name = "yerk"
domain = "personal"
remote = "git@github.com:salotz/yerk.git"
path = "devel/yerk"   # → <domains.personal>/devel/yerk
```

**Resolve project workspace:**

| Catalog `path` | Resolution |
| --- | --- |
| Absolute (or `~/…` expanded absolute) | Use as workspace (host escape hatch). |
| Relative | Require non-empty `domain`; join `domains[domain]` + path. |

**Then** apply `[workspace].style` under that workspace (unchanged):

- `workspace-dir` → `<workspace>/<replica>`
- `project-dir` → `<dir(workspace)>/<name>__<replica>`

`domain` is therefore both a **namespace label** and the key into
`[domains]` when `path` is relative.

### What this is not

- Not a single shared `[workspace].root` for every project on the host.
- Not automatic inventing of domain roots from `~/tree/<domain>`; operators
  configure roots explicitly per machine.
- Catalog remains hand-edited; no requirement that catalog be shared across
  hosts, only that relative form *can* be shared when domains align.

## Consequences

- `config.toml` is the host-specific half (domain roots + style); `catalog.toml`
  can stay portable when paths are relative.
- ADR 003 / 004 still hold: XDG split, `YERK__` overlays for style/files.
  No new env var for domain roots in this ADR (edit config; revisit if needed).
- Workspace package takes full `config.Config` (style + domains), not only
  `[workspace]`.
- Docs, examples, `yerk config show`, and host dogfood config must show
  `[domains]` and relative `path`.
- Absolute catalog paths remain valid for one-off or odd placements.
- Tests should prefer relative path + domain root fixtures.

## Related

- [domain-and-near-term.md](../domain-and-near-term.md) — nouns / placement
- [003](./003-config-xdg-and-env.md), [004](./004-config-and-catalog-split.md)
- Host domains: RFC 25; PRJX path vocabulary: RFC 28
