# Domain language and near-term behavior

Status: **draft accepted for near-term** (operator design session 2026-09-25).

Source: idea note `software-project-management-tool`, `design/goals.md`,
PRJX (RFC 28), host domains (RFC 25), and this workshop.

This document freezes vocabulary and the first feature slice. Implementation
plans live under `.agents/plans/`. Longer-horizon features stay in `goals.md`.

---

## Product one-liner

`yerk` is a **host multi-project manager**: it keeps a **catalog** of software
projects, **materializes** their **replicas** into a **workspace** layout,
**resolves** paths by name, and reports **presence** and **change** status.
It speaks PRJX vocabulary; it does not redefine the PRJX spec.

---

## Nouns

| Noun | Meaning |
| --- | --- |
| **Project** | Named unit of software work managed on this host. Catalog identity is a short `name`. Optional `domain` namespaces the project and selects a host domain root when `path` is relative. MVP assumes one primary **remote**; multi-remote is allowed later. |
| **Catalog** | Host-global registry of projects (`catalog.toml`). Source of truth for *what* exists. |
| **Config (tool)** | Host/tool behavior (`config.toml`): workspace **style**, probe defaults, later hooks. Source of truth for *how on this host*. May differ per machine. |
| **Remote** | Clone URI (git URL or path) for the project’s canonical VCS content. |
| **Domain** | Namespace / context label on a project (e.g. `personal`, `examol`). Selects a host domain root when `path` is relative (ADR 008). |
| **Tag** | Declared bulk-select label. Catalog root `tags = […]` is a **closed vocabulary**; each project’s `tags` must be members (ADR 010). Orthogonal to domain. |
| **Project workspace** | On-disk directory that **owns** a project's replicas. Catalog `path` points here (e.g. `…/devel/yerk`). Not a git checkout. |
| **Workspace (policy)** | How replicas are placed relative to each project workspace (`style`). A shared host-wide root for every project is deferred. |
| **Replica** | One concrete on-disk checkout of a project on this host (PRJX), e.g. `…/yerk/main`. |
| **Replica distinguisher** | Token separating replicas of the same project (often default branch short name). |
| **Change status** | Observed git dirtiness / sync flags for a replica. |
| **Presence status** | Whether the expected replica path is missing, present, or invalid on disk. |
| **Local (config)** | Host/replica-only files staged into a checkout (PRJX `.local` / RFC 26). Named for later; out of near-term behavior. |

### Avoid conflating

- Project ≠ directory; a project may have zero or many replicas.
- Replica ≠ branch; branch is git; replica is host placement (often tracking a branch).
- **Project workspace ≠ replica**; catalog `path` is the workspace (`…/yerk`), not `…/yerk/main`.
- Workspace policy ≠ one global root (yet); style applies under each project's path.
- Tag ≠ domain.
- `yerk` ≠ PRJX (tool vs spec).
- **Checkout** is avoided as a product verb; it overloaded too many actions.

---

## Verbs (internal) and CLI (short)

| Intent | Internal verb | Near-term CLI |
| --- | --- | --- |
| Add/update catalog row | register | hand-edit catalog (future `yerk register`) |
| List / show states | status (read model) | `yerk status [project [replica]]` [`--tag`]; change on by default (`--presence-only` opt-out) |
| Path math only | resolve | `yerk path` (alias: `resolve`); bare name → workspace, +replica → checkout |
| Create project workspace dir | materialize workspace | `yerk workspace ensure <proj>…` or `--all` (later also `--tag`) |
| Create replica via git | materialize replica | `yerk clone <proj> [replica]` \| `--all` \| `--tag` |
| Read git state | probe change status | part of status (opt-out flag); not a separate default verb |
| Universal read | get resource | later: `yerk get <kind> …` over API resources |
| Sync | pull / push | later — **design semantics first** (no stub CLI) |
| Apply host-local files | stage locals | later |

Materialize **workspace** (`yerk workspace ensure`) and materialize **replica**
(`yerk clone`) are always separate steps in the model.

### Project selection (bulk)

Commands that act on **many** projects share one selection model (ADR 010):

| Selector | Meaning | Near-term |
| --- | --- | --- |
| (default / all catalog) | Every `[[projects]]` row | `yerk status` with no `--tag` |
| `--tag <name>` | Projects that list declared tag `<name>` | `status --tag`; `clone --tag` |
| project name args | Explicit subset | `workspace ensure <proj>…`, `clone <proj>` |
| `--all` | Explicit full catalog (opt-in bulk mutate) | `workspace ensure --all`, `clone --all` |

Rules:

- `<name>` for `--tag` must appear in catalog root `tags` (unknown → error).
- Declared tag with zero projects → empty match (not an error) for **read** ops
  (`status`); **mutate** ops (`clone`) refuse empty selection.
- For mutate commands that support bulk: project args, `--all`, and `--tag` are
  mutually exclusive (one selector). Names and `--all` also exclusive on
  `workspace ensure` (ADR 009); ensure still lacks `--tag`.
- Implementation path: `Catalog.SelectByTag` + `selectProjects` / clone
  selection helper.

---

## Identity

Near-term:

```text
project:  <name>                      # catalog key
ref:      <name>[/<distinguisher>]    # default distinguisher if omitted
fq hint:  <domain>.<name>             # display / future PRJX align; not required in paths yet
```

Default replica distinguisher when omitted:

1. Optional per-project override in catalog (`default_replica`), else
2. Remote default branch short name (`git ls-remote --symref <uri> HEAD`), else
3. Fallback `main`.

---

## Status model

### Presence status (disk ↔ expected replica path)

| Value | Meaning |
| --- | --- |
| `missing` | Catalog expects a path; it does not exist. |
| `present` | Path exists and is a usable git checkout (`.git` file or directory / worktree). |
| `invalid` | Path exists but is not a usable git checkout. |

Column / docs label: **presence** (not “placement”).

### Change status (git probe; only when presence=`present`)

Orthogonal **flags** (combinatorial), not a single exclusive enum.

**Local (MVP; on by default unless `--presence-only`):**

| Flag | Meaning |
| --- | --- |
| `clean` | No staged, unstaged, or untracked changes. |
| `dirty` | Staged and/or unstaged modifications to tracked files. |
| `untracked` | Untracked files present (separate from `dirty`). |

**Sync (include in probe when cheap enough; may be `sync-unknown` without network):**

| Flag | Meaning |
| --- | --- |
| `ahead` | Local has commits not in upstream. |
| `behind` | Upstream has commits not in local. |
| `no-upstream` | No tracking branch. |
| `sync-unknown` | Upstream query failed or skipped. |

**Probe-level:** omit change flags when presence ≠ `present`; `error` if git probe fails unexpectedly.

Display: space-separated flags; show `clean` only when no other local flags apply.

### Status scopes (CLI)

| Invocation | Resource view | Notes |
| --- | --- | --- |
| `yerk status` | **Project** list (catalog / `--tag`) | Default human table. |
| `yerk status <project>` | One **project** | |
| `yerk status <project> <replica>` | One **replica** | Full presence + change for that checkout. |
| `yerk status --tag <name>` | **Project** list filtered | Declared tag only (ADR 010). |

**Project** status (default list / single project):

- Emphasize **project workspace** path (not “only the default replica path”).
- Do **not** show a **REPLICA** column on the default project view (replica
  identity belongs on replica-scoped status). Multi-replica lists are not the
  default view.
- May summarize default-replica presence/change without pretending the row *is*
  a replica.

**Replica** status (`status <project> <replica>`):

- Owns path, presence, change flags, branch, etc. for that distinguisher.

**Change status:**

- **On by default** when probing is applicable (present checkout).
- Opt-out: **`--presence-only`** skips git change probes (fast presence scan).
- Historical `--git` as opt-in is inverted: change is default; `--git` is a
  hidden no-op for old scripts.
- `status --fetch` (fetch then probe) remains later, not required now.

Near-term probes target the **default replica** when summarizing a project,
not every child directory under the workspace. Multi-replica discovery can
follow once resources/`get` are in place.

### Explicit resources (model-driven)

Pilot kubectl-style **API resources** as the shared model (even if ROI is
partly ecosystem-pattern learning):

- Package: `internal/api` (name TBD in ADR) — stable types, not TOML-only structs
  and not CLI tabwriter rows.
- Kinds include at least project registry objects, **ProjectStatus**,
  **ReplicaStatus**; map loaders/adapters → resources → printers.
- Enables later: `--output json|yaml|table`, JSON Schema / Go type docs,
  universal `yerk get`, alternate backends (files today → DB later) without
  rewriting command logic.

See plan P5–P6 in `.agents/plans/near-term.md`. Push/pull stay blocked on a
dedicated semantics design (P9) so project vs replica sync is not confusing.

---

## Workspace placement (near-term)

1. **Resolve project workspace** from catalog + host domains ([ADR 008](./decisions/008-domain-roots-and-relative-catalog-paths.md)):

   | Catalog `path` | Workspace |
   | --- | --- |
   | Relative | `<domains[domain]>/<path>` (domain required) |
   | Absolute (`~/` ok) | that path (host escape hatch) |

2. Tool config **`[workspace].style`** maps replica distinguisher `R` under
   that workspace:

| Style | Replica path | Example |
| --- | --- | --- |
| `workspace-dir` | `<workspace>/<R>` | ws=`…/yerk`, R=`main` → `…/yerk/main` |
| `project-dir` | `<dir(workspace)>/<name>__<R>` | ws=`…/yerk`, R=`main` → `…/yerk__main` |

Domain roots are **per domain**, not one shared root for every project name.


## Config vs catalog files

Under `$XDG_CONFIG_HOME/yerk` (see ADR 004):

| File | Owns |
| --- | --- |
| `config.toml` | Tool/host behavior: `[workspace].style`, `[domains]` roots, future defaults. |
| `catalog.toml` | Root `tags = […]` vocabulary + `[[projects]]` registry (prefer relative project-workspace `path`). |

Env:

- `YERK__CONFIG` / `YERK__CONFIG_DIR` — tool config (existing direction).
- `YERK__CATALOG` — explicit catalog file path.
- `YERK__WORKSPACE_STYLE` — optional style overlay.

Missing catalog ⇒ empty registry. Missing config ⇒ defaults.

### Catalog row (MVP sketch)

```toml
tags = ["devel"]          # closed vocabulary (ADR 010); project tags ⊆ this list

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@github.com:salotz/yerk.git"
tags = ["devel"]
# default_replica = ""    # empty → remote HEAD branch name
path = "devel/yerk"       # → <domains.personal>/devel/yerk (not …/yerk/main)
```

### Tool config (MVP sketch)

```toml
[workspace]
style = "workspace-dir"   # → <workspace>/<replica>

[domains]
personal = "~/tree/personal"
```

---

## Near-term feature slice

In scope:

1. **Status** — project vs replica scopes; workspace-oriented project view;
   **change on by default** (opt-out); serial then parallel probes.
2. **Tag vocabulary + bulk select** — closed catalog `tags`; `status --tag`
   (ADR 010); extend selector to ensure/clone.
3. **Ensure** project workspace directories (ADR 009).
4. **Clone** default-branch replica.
5. **Resolve** paths (`yerk path`).
6. **Explicit API resources** pilot — typed model package, then json/yaml,
   schemas, `yerk get` as follow-ons.
7. **Push/pull design** before any sync CLI.

Explicitly out of near-term (or blocked):

- Local config staging, FS watch daemon, port/resource tracking, mutagen, bulk
  domain-tree mapping, monorepo presets, `register` mutation UX.
- **Shipping `pull`/`push`** before sync semantics are designed.
- Do not ship unimplemented CLI stubs; add commands when behavior exists.

---

## Architecture (near-term)

```text
CLI (cobra)
  → load config.toml + catalog.toml
  → map / collect into internal/api resources
  → layout (pure path math; ensure dirs)
  → git adapter (default branch, clone, presence, change probe)
  → print resources (table today; json|yaml later)
```

Principles:

1. Catalog = *what*; config = *how on this host*.
2. Layout is pure (no git, no network).
3. Only the git adapter talks to git.
4. Materialize workspace ≠ materialize replica ≠ stage locals (later).
5. Status is a read model: join(catalog, resolve, presence[, change]) → api
   status resources.
6. No daemon in near-term.
7. Parallel git probes are **planned early** in the adapter API (e.g. worker
   pool) but can ship serial first; the domain model must not assume serial-only
   semantics.
8. Product types live in `internal/api`; TOML load structs stay in
   `internal/config`; CLI only selects and prints (ADR 011).

---

## Technical defaults (near-term)

| Topic | Choice | Rationale |
| --- | --- | --- |
| Git integration | Shell out to `git` on `PATH` | Host-consistent; full porcelain/plumbing without re-implementing; easy `ls-remote`, `status --porcelain=v2`, worktrees. Swap-able behind an interface. See note below. |
| Default branch | `git ls-remote --symref <remote> HEAD` | Works before clone. |
| Catalog edits | Hand-edit (agents OK); `register` later | Matches operator workflow. |
| Ensure | `mkdir -p` project workspace only; never delete; no replica leaf | Safe materialize workspace ([ADR 009](./decisions/009-workspace-subcommand-and-ensure-scope.md)). |
| Clone into non-empty path | Refuse | Predictable. |
| Parallel status | Design for N-way probe; implement after vertical slice | Operator priority: sooner rather than buried; after domain model works. |

### Why `git` subprocess instead of go-git?

Not because subprocess is “simpler” in the abstract — it is a **tradeoff**:

- **go-git** pros: pure Go, no `git` binary dependency, in-process control.
- **go-git** cons: subset/impedance mismatch with real git (SSH agents, hooks,
  worktrees, sparse, credential helpers, auth oddities); status/sync edge cases
  often end up shelling out anyway; larger dependency surface for an operator
  tool that already assumes a dev host with git.

- **`git` CLI** pros: identical behavior to the operator’s git; one mental
  model; trivial to debug (`GIT_TRACE`); covers default-branch discovery,
  porcelain status, worktrees, fetch later.
- **`git` CLI** cons: requires `git` on `PATH`; process overhead; parsing
  porcelain carefully; harder in minimal containers without git.

**Decision:** adapter interface in Go, **default implementation = `git`
subprocess**. Revisit go-git (or hybrid) if we target hosts without git or hit
process scaling limits after parallelization.

---

## Related docs

- [goals.md](./goals.md) — product north star and post-MVP directions
- [decisions/](./decisions/) — ADRs (esp. 003 XDG, 004 config vs catalog, 005 CLI help / envvars)
- [../docs/](../docs/) — operator/user docs (Diátaxis; ad hoc Markdown, [ADR 006](./decisions/006-ad-hoc-docs-diataxis.md))
- [../.appinfo/meta.toml](../.appinfo/meta.toml) — application info + env registry (RFC 030/031)
- [../.agents/plans/near-term.md](../.agents/plans/near-term.md) — implementation plan
