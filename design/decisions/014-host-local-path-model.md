# 014. Host-local path model (successor to ADR 008)

## Status

Accepted (2026-10-01); corrected same day:

1. Host paths live in **config**, not catalog (`[[projects]]` rows).
2. **Optional `[domains]` roots** restore the common default join so most
   catalog projects need **no** per-project path row. A full (absolute/`~/`)
   path is required only when the domain root is unused or placement differs
   from the default.

Supersedes **path resolution** in
[ADR 008](./008-domain-roots-and-relative-catalog-paths.md) as the placement
engine, while **reusing** the useful half of ADR 008: host-owned domain roots
as an **optional default base**. Domain remains a **logical namespace** for
ids/URIs (ADR 012). Catalog stays portable (no host `path` field).

## Context

ADR 008 bound each catalog `domain` to a host-absolute root in `config.toml`
`[domains]` and joined **catalog** relative `path` values to that root. That
put **host-local paths into the catalog** (or forced every catalog to be
host-specific).

Operator design then locked:

- Domains are a namespace for ids/URIs (ADR 012).
- Workspace / replica **locations** are **host-local configuration**.
- The **catalog** is the portable project registry (what exists: name, domain,
  remote, tags, optional style/method overrides).
- Most projects on a machine share a per-domain tree (e.g. all `personal`
  workspaces under one root). Requiring a full path row for every catalog
  project is noise.
- Odd placements (home-dot dirs, out-of-tree paths) still need an explicit
  host path.

So the path engine is host-local, but **domain roots are a convenient default**,
not identity-as-filesystem and not a catalog concern.

## Decision

### Split: registry vs host placement

| File | Owns |
| --- | --- |
| `catalog.toml` | Project **identity** and VCS metadata: `name`, `domain`, `remote`, `tags`, `default_replica`, optional `workspace_style` / `replica_method` |
| `config.toml` | Host **placement**: ambient `[workspace].style`, optional **`[domains]`** roots, optional **`[[projects]]`** host rows |

### Domain is a namespace; roots are optional host defaults

1. **Domain** appears in catalog identity, bare ids, and `yerk://` URIs.
2. **`[domains]`** (optional) maps domain name → host-absolute root (`~/…` ok).
3. Placement **never invents** `~/tree/<domain>` from the domain string alone;
   operators set roots explicitly when they want the default join.
4. Domain is **not** required to have a root: projects can use a full host path
   instead.

### Where the project workspace path comes from

Resolve **project workspace** for `domain` + `name`:

| Step | Condition | Result |
| --- | --- | --- |
| 1 | Host `[[projects]]` row has **absolute** or **`~/…`** `path` | That path (after `~` expand) |
| 2 | Host row has **relative** `path` | Require `[domains.<domain>]`; join `root + path` |
| 3 | No host row, or row with **empty** `path` | Require `[domains.<domain>]`; join `root + name` |
| 4 | Else | **Error**: set `[domains.<domain>]` or a full `[[projects]]` path |

```toml
# config.toml — typical host: domain roots only
[workspace]
style = "workspace-dir"

[domains]
personal = "~/tree/personal/devel"
examol   = "~/tree/examol/devel"

# Only exceptions / style overrides need rows:
[[projects]]
name = "bimker"
domain = "personal"
path = "~/.bimker"
# workspace_style = "name-tags"   # optional host-row override (ADR 013)
```

With the roots above, catalog project `personal/yerk` resolves to
`~/tree/personal/devel/yerk` **without** a host project row.

| Host `path` | Domain root | Resolution |
| --- | --- | --- |
| Absolute or `~/…` | optional | Clean absolute path = workspace |
| Relative (e.g. `other/yerk`) | **required** | `<root>/<path>` |
| Empty / missing row | **required** | `<root>/<name>` |
| Empty / missing row | missing | Error |
| Relative | missing | Error |

Match key for host rows is `domain` + `name` (same bare id as the catalog).
Extra host rows without a catalog entry are ignored at catalog resolve;
`config show` still lists them. Host rows may grow fields (style, method, …)
without catalog changes.

**Choose the domain root** so the default `<root>/<name>` matches the usual
tree (often the host’s `devel` directory under a domain, not the domain home
itself). Deeper or sibling layouts use a relative or absolute host `path`.

### Catalog must not carry host workspace paths

- **No `path` field** on catalog `[[projects]]`.
- If a legacy catalog still has `path`, load **errors** with guidance to move
  placement into `config.toml` (`[domains]` and/or `[[projects]]`) and delete
  the catalog field.

### Relation to ADR 008

| ADR 008 | This ADR |
| --- | --- |
| Catalog relative `path` + `[domains]` join | Catalog has **no** path; default join is **`root + name`** |
| Absolute catalog path escape hatch | Absolute/`~/` on host **`[[projects]].path`** |
| Domain root effectively required for relative rows | Domain root **optional**; full host path when unused |
| Domain doubles as fs key for every project | Domain roots are a **host convenience default** |

### Style path math (shipped styles)

Once an absolute **workspace base** exists:

- `workspace-dir` → `<workspace>/<replica>`
- `project-dir` → `<dir(workspace)>/<name>__<replica>`

Effective style comes from the placement merge (ADR 013). Host-row
`workspace_style` is an ambient layer more specific than process env.

### State vs location

Bound **style** (ADR 013) wins over ambient style with a warning. Workspace
**location** is host config (domain default and/or `[[projects]].path`) unless
a later ADR snapshots path into state.

## Consequences

- `internal/config.Config` has `Domains map[string]string` and
  `Projects []HostProject` (`name`, `domain`, optional `path`, …).
- `Config.ProjectWorkspacePath` implements the table above.
- `internal/workspace` resolves workspace via that helper.
- Catalog types reject legacy `path`.
- Examples: portable catalog without paths; sample `config.toml` shows
  `[domains]` plus one override row.
- `yerk catalog show` has no path column; host placement appears via
  `config show` / `status` / `path`.
- Dogfood: set domain roots to the common parent of project workspaces; keep
  `[[projects]]` only for exceptions.

## Related

- [004](./004-config-and-catalog-split.md) — config vs catalog lifetimes
- [008](./008-domain-roots-and-relative-catalog-paths.md) — prior path join
- [012](./012-identifiers-and-yerk-uri.md) — domain in ids
- [013](./013-placement-policy-and-host-state.md) — style layers and state
- [domain-and-near-term.md](../domain-and-near-term.md)
