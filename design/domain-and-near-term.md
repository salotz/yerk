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
| **Project** | Named unit of software work managed on this host. Catalog identity is required `domain` + `name` (bare id `domain/name`, URI `yerk://…`; ADR 012). MVP assumes one primary **remote**; multi-remote is allowed later. |
| **Catalog** | Host-global registry of projects (`catalog.toml`). Source of truth for *what* exists. |
| **Config (tool)** | Host/tool behavior (`config.toml`): workspace **style**, optional **`[domains]`** roots, optional host **`[[projects]]`** rows (path overrides, later knobs). Source of truth for *how on this host*. May differ per machine. |
| **Remote** | Clone URI (git URL or path) for the project’s canonical VCS content. |
| **Domain** | Logical namespace on a project (e.g. `personal`). Required for ids/URIs (ADR 012). Optional host `[domains]` root is a convenience default, not identity (ADR 014). |
| **Tag** | Declared bulk-select label. Catalog root `tags = […]` is a **closed vocabulary**; each project’s `tags` must be members (ADR 010). Orthogonal to domain. |
| **Project workspace** | On-disk directory that **owns** a project's replicas. Default `<domains[domain]>/<name>` or host `[[projects]]` path (ADR 014). Not a git checkout. |
| **Workspace (policy)** | How replicas are placed relative to each project workspace (`style`). Effective style merges host / dir-local / catalog / **host state** / env / CLI (ADR 013). |
| **Replica** | One concrete on-disk checkout of a project on this host (PRJX), e.g. `…/yerk/main`. |
| **Replica distinguisher** | Token separating replicas of the same project (often default branch short name). |
| **Change status** | Observed git dirtiness / sync flags for a replica. |
| **Presence status** | Whether the expected replica path is missing, present, or invalid on disk. |
| **Local (config)** | Host/replica-only files staged into a checkout (PRJX `.local` / RFC 26). Named for later; out of near-term behavior. |

### Avoid conflating

- Project ≠ directory; a project may have zero or many replicas.
- Replica ≠ branch; branch is git; replica is host placement (often tracking a branch).
- **Project workspace ≠ replica**; workspace is the parent layout base (`…/yerk`), not `…/yerk/main`.
- Workspace policy ≠ one global root (yet); style applies under each project's path.
- Tag ≠ domain.
- `yerk` ≠ PRJX (tool vs spec).
- **Checkout** is avoided as a product verb; it overloaded too many actions.

---

## Verbs (internal) and CLI (short)

| Intent | Internal verb | Near-term CLI |
| --- | --- | --- |
| Add/update catalog row | register | hand-edit catalog (future `yerk register`) |
| List / show states | status (read model) | `yerk status` multi-project overall; `status <project>` all live replicas; `status <project> <replica>` one row; `--tag`; change on by default (`--presence-only` opt-out) |
| Path math only | resolve | `yerk path` (alias: `resolve`); bare name → workspace, +replica → checkout |
| Create project workspace dir | materialize workspace | `yerk workspace ensure <proj>…` or `--all` (later also `--tag`) |
| Create replica via git | materialize replica | `yerk materialize <id> [replica]` \| `--all` \| `--tag` |
| Session replica spin-out | create replica | `yerk replica create <id> <replica>` (`--method worktree\|clone`; ADR 016) |
| Read git state | probe change status | part of status (opt-out flag); not a separate default verb |
| Universal read | get / lookup | `yerk get <id>`; `yerk lookup <path>`; `project|replica get|lookup`; `--output json` (ADR 015) |
| Explain placement | resolve config | `yerk config resolve <project-id>` (contribution stack; ADR 013) |
| Sync | pull / push | later — **design semantics first** (no stub CLI) |
| Apply host-local files | stage locals | later |

Materialize **workspace** (`yerk workspace ensure`) and materialize **replica**
(`yerk materialize`) are always separate steps in the model. Session spin-out
(`yerk replica create`) is a third path: worktree from main or clone-method
from remote (not bulk materialize). Product CLI does
not use a top-level `clone` command (git still runs `git clone` under the hood).

### Project selection (bulk)

Commands that act on **many** projects share one selection model (ADR 010):

| Selector | Meaning | Near-term |
| --- | --- | --- |
| (default / all catalog) | Every `[[projects]]` row | `yerk status` with no `--tag` |
| `--tag <name>` | Projects that list declared tag `<name>` | `status --tag`; `materialize --tag` |
| project id args | Explicit subset (ADR 012 forms) | `workspace ensure <id>…`, `materialize <id>` |
| `--all` | Explicit full catalog (opt-in bulk mutate) | `workspace ensure --all`, `materialize --all` |

Rules:

- `<name>` for `--tag` must appear in catalog root `tags` (unknown → error).
- Declared tag with zero projects → empty match (not an error) for **read** ops
  (`status`); **mutate** ops (`materialize`) refuse empty selection.
- For mutate commands that support bulk: project args, `--all`, and `--tag` are
  mutually exclusive (one selector). Names and `--all` also exclusive on
  `workspace ensure` (ADR 009); ensure still lacks `--tag`.
- Implementation path: `Catalog.SelectByTag` + `selectProjects` / materialize
  selection helper.

---

## Identity

Canonical forms (ADR 012):

```text
project:  <domain>/<name>             # bare id; catalog fields domain + name
          yerk://<domain>/<name>      # canonical URI
replica:  <domain>/<name>/<replica>
          yerk://<domain>/<name>/<replica>
short:    <name> or <name>/<replica>  # unique short name expands; else error
```

Default replica distinguisher when omitted (operation policy, not id expand):

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

**Sync (local remote-tracking refs; no fetch unless a future `--fetch`):**

| Flag | Meaning |
| --- | --- |
| `ahead:N` | Local has N commits not in the comparison ref. |
| `behind:N` | Comparison ref has N commits not in local. |
| `no-upstream` | No comparison ref (`@{upstream}` and `origin/<branch>` both missing). |
| `sync-unknown` | Compare ref chosen but counts failed or skipped. |

**Comparison ref (MVP):** `@{upstream}` if set, else `origin/<current-branch>`
when that remote-tracking ref exists. Remote name hard-coded `origin` for now.

**Probe-level:** omit change flags when presence ≠ `present`; `error` if git probe fails unexpectedly.

Display: space-separated flags; show `clean` only when no other local flags apply.
In-sync (ahead=0, behind=0) adds no sync token.

**Follow-on:** `@{push}`, non-`origin` remotes, catalog remote name, triangular
workflows — see [status-model.md](../docs/explanation/status-model.md).

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

Push/pull stay blocked on a dedicated semantics design so project vs replica
sync is not confusing. Active implementation coordination lives under
ephemeral `.agents/plans/<owner>/` (not durable product docs).

---

## Workspace placement (near-term)

1. **Resolve project workspace** from host config only ([ADR 014](./decisions/014-host-local-path-model.md)):

   | Host placement | Workspace |
   | --- | --- |
   | `[[projects]]` absolute / `~/` path | that path |
   | `[[projects]]` relative path | `<domains[domain]>/<path>` |
   | No host path (typical) | `<domains[domain]>/<name>` |

   Catalog has **no** `path`. Domain roots are **optional**; without a root,
   supply a full host path. Choose the root so `<root>/<name>` matches the
   usual tree (often a host `devel` directory).

2. Tool config **`[workspace].style`** (plus placement layers, ADR 013) maps
   replica distinguisher `R` under that workspace:

| Style | Replica path | Example |
| --- | --- | --- |
| `workspace-dir` | `<workspace>/<R>` | ws=`…/yerk`, R=`main` → `…/yerk/main` |
| `project-dir` | `<dir(workspace)>/<name>__<R>` | ws=`…/yerk`, R=`main` → `…/yerk__main` |

Domain roots are **per domain**, not one shared root for every project name.


## Config vs catalog files

Under `$XDG_CONFIG_HOME/yerk` (see ADR 004):

| File | Owns |
| --- | --- |
| `config.toml` | Tool/host behavior: `[workspace].style`, optional `[domains]` roots, optional host `[[projects]]` overrides. |
| `catalog.toml` | Root `tags = […]` vocabulary + portable `[[projects]]` registry (no host paths). |

Env:

- `YERK__CONFIG` / `YERK__CONFIG_DIR` — tool config (existing direction).
- `YERK__CATALOG` — explicit catalog file path.
- `YERK__STATE_DIR` — project bindings (ADR 013).
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
# no path — host config places the workspace
```

### Tool config (MVP sketch)

```toml
[workspace]
style = "workspace-dir"   # → <workspace>/<replica>

[domains]
personal = "~/tree/personal/devel"   # personal/yerk → …/devel/yerk

# [[projects]] only for exceptions:
# name = "bimker"
# domain = "personal"
# path = "~/.bimker"
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
- [../.agents/plans/](../.agents/plans/) — ephemeral implementation plans (owner folders)
