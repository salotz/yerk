# Plan: host-local staging / inject (design only)

## Context

Creating session replicas (worktrees) is cheap; remembering to copy
**host-local** files into each one is not. The repeating incident is DVC
`config.local` (cache dir, link types) present on `main` and missing on every
new worktree. Same class of files: direnv `.envrc.local`, mise local overlays,
editor/host paths, anything that must not be committed.

Product already anticipated this:

- [design/goals.md](../../../../design/goals.md) — “Staging local configuration”
  (workspace `.local` and/or `~/.config/yerk/projects/<project>/locals/…`)
- [design/domain-and-near-term.md](../../../../design/domain-and-near-term.md) —
  noun **Local (config)**; verb **stage locals**; explicitly out of near-term
- [ADR 003](../../../../design/decisions/003-config-xdg-and-env.md) — later
  `~/.config/yerk/projects/<name>/locals/`
- RFC 26 (domain-local `.local` beside a collection) and RFC 28 (PRJX
  `.local` at a project root)

Operator first-pass (this session):

- Library under the yerk config dir: `project-configs/<domain>/<project>/<host>/`
- Formalize **host** as a product noun
- Command `stage` (name fuzzy)
- Formalize project **lifecycle** on a host
- Bundle steps into a declarative `ensure`
- Example flow: workspace dir → stage workspace locals → materialize `main` →
  inject into `main` (then later replicas)

This plan iterates that sketch against shipped ADRs (004, 007, 009, 013, 014,
016), the materialize-DWIM track (declarative replica ensure, not yet locked),
and the git-stage word collision in this operator’s workflow.

Product rule: do not ship CLI stubs without behavior. This plan produces an
**ADR** (and pointers). Implementation is a **separate** plan after accept.

---

## Goals

1. Accepted **ADR** for host-local file apply: source tree, host identity,
   workspace vs replica targets, idempotency, CLI atoms.
2. Written **lifecycle** (ordered steps) in the ADR / domain spine — even if
   the composer command is deferred.
3. Clear **non-goals** (templates, hook framework, RFC 26 collection sibling
   as the only store, sync of locals across hosts).
4. Handoff: implement plan folder or explicit backlog defer.

## Non-goals (this plan)

- Implementing any new cobra commands
- Templating (`{{cache_dir}}`), secret injection, git-crypt
- A general hook system (note integration *points* only)
- Changing `workspace ensure` into a file copier (ADR 009 stays mkdir-only
  unless a Q locks otherwise)
- Reviving top-level `yerk ensure` in the same breath as the atoms (composer
  is a separate Q)
- Pull/push (sync-verbs-design)
- Materialize DWIM (materialize-dwim) — only how locals attach after a replica
  exists
- Committing host-private DVC paths as examples (ADR 007)

---

## Iteration on the first pass

### What to keep

| First pass | Keep |
| --- | --- |
| Host-specialized files (laptop vs HPC) | Yes — host is a real noun |
| Split **workspace** locals vs **replica** inject | Yes — different destinations, different tools |
| Lifecycle as an ordered list of steps | Yes — document even before a composer |
| DVC `config.local` as the motivating file | Yes |
| Missing-on-HPC is valid | Yes — empty apply is success |
| Later hooks | Integration points only |

### What to change

**1. Do not invent a sibling silo `project-configs/`.**

Config dir already has identity-shaped trees in **state**:
`$XDG_STATE_HOME/yerk/projects/<domain>/<project>/`. ADR 003 sketched
`projects/<name>/locals/` (no domain, no host). Align the library with
**domain + name**, under config (operator-authored), not state
(tool-written):

```text
$XDG_CONFIG_HOME/yerk/projects/<domain>/<name>/
```

`project-configs/` duplicates “projects” and drifts from state/ADR 003.

**2. Host is an overlay, not the only directory key.**

The config directory often already lives on one host. A required
`<host>/` segment is noise when files are not shared across machines.
It **is** justified when the same tree is synced (bimker `lib/configs/yerk`)
or copied to HPC.

```text
projects/<domain>/<name>/
  workspace/                 # all hosts
  replica/
  hosts/<host>/
    workspace/               # overlay, host wins per relative path
    replica/
```

**3. Do not name the product verb `stage`.**

Internal docs can keep “stage locals.” CLI `stage` collides with git
index staging (this operator’s review UI). Prefer parent **`yerk locals`**
and verb **`apply`**.

**4. Do not overload `workspace ensure` or rush a top-level `ensure`.**

ADR 009 removed top-level `ensure` and locked `workspace ensure` = mkdir
project workspace only. Materialize-DWIM is separately claiming
“declarative replica ensure.” A third `ensure` without atoms would mush
those boundaries.

Atoms first; optional composer (`yerk project ready` or a carefully
scoped `yerk ensure`) after. **Painkiller:** auto-apply replica locals
after successful `materialize` / `replica create`.

**5. Reserve `.local/yerk/`.**

ADR 013 walks `<dir>/.local/yerk/config.toml` for **placement style**.
Staged DVC files must not live in that directory. Operator path
`<workspace>/.local/config/yerk/` is the better workspace library.

**6. XDG library is the SSOT; workspace `.local` is a materialization;
replica files are what tools read.**

RFC 26 prefers sharing one file over copying into every worktree. DVC
does not read a pointer at the workspace — it wants
`<replica>/.dvc/config.local`. So replica inject is mandatory for that
class. Workspace staging remains useful as the RFC 26/PRJX-facing copy
and as the symlink target if we choose links later.

---

## Proposed model (defaults for Qs)

### Nouns

| Noun | Meaning |
| --- | --- |
| **Host name** | Operator identity of this host system (glossary). Not required to equal the kernel hostname (HPC login node, container). |
| **Locals library** | Operator-authored trees under `$XDG_CONFIG_HOME/yerk/projects/<domain>/<name>/`. |
| **Workspace locals** | Files whose destination is the **project workspace** (not a git checkout). |
| **Replica locals** | Files whose destination is a **replica root** (untracked, inside the checkout). |
| **Apply** | Copy (v1) library files to a destination tree. Idempotent. |

### Host identity

Resolve **this host** as:

1. `YERK__HOST` if set
2. Else `config.toml` `[host] name = "mantis"`
3. Else short hostname (`os.Hostname()`, strip domain)

Unknown / unset host still allows **common** `workspace/` and `replica/`
trees. Overlay `hosts/<name>/` is used only when the resolved name matches
a directory.

### Library layout

```text
$XDG_CONFIG_HOME/yerk/
  config.toml
  catalog.toml
  projects/
    personal/
      wepy/
        replica/
          .dvc/config.local          # optional common
        hosts/
          mantis/
            replica/
              .dvc/config.local      # laptop cache path
            workspace/
              .local/config/yerk/    # optional library copy at workspace
          # no hosts/hpc-x/ → HPC apply is a no-op
```

Merge: **common then host overlay**, file-level. Host file replaces the
same relative path. No deep TOML merge in v1 (these are opaque blobs).

Relative paths under `replica/` are relative to the replica root.
Relative paths under `workspace/` are relative to the project workspace.

No manifest file in v1. Convention is the mapping. Optional
`locals.toml` later if ignore-lists / chmod / link-vs-copy per file
are needed.

Catalog does **not** store these files (portable registry). Config dir
does (host-local; ADR 007 — examples use fake host names only).

### Destinations (operator example)

| Step | Dest |
| --- | --- |
| Apply workspace | `<ws>/.local/config/yerk/.dvc/config.local`, `<ws>/mise.toml` if present in library `workspace/` |
| Apply replica `main` | `<ws>/main/.dvc/config.local` (style-aware replica path) |

### Apply semantics (v1)

| Dest state | Behavior |
| --- | --- |
| Missing | Create parents; copy file |
| Exists, same bytes | Success (no-op) |
| Exists, different, **untracked** | **Error** unless `--force` |
| Exists, **git-tracked** | **Error** (never clobber tracked files) |
| Dest is a directory when source is a file | Error |

- Never `git add`.
- Missing library for that project/host → **success, nothing to do**
  (HPC with no overlay).
- Bulk selectors: same XOR as other mutators (ids / `--all` / `--tag` /
  `--domain`).
- Replica apply requires presence = usable checkout; else error.
- v1 **copy** (predictable). Symlink (`--link`, or replica → workspace
  library) is a follow-on.

### Lifecycle (document in ADR / spine)

Ordered steps for one project on one host:

| # | Step | CLI atom today / proposed | Notes |
| --- | --- | --- | --- |
| 1 | Register | hand-edit `catalog.toml` | Portable *what* |
| 2 | Place | hand-edit `config.toml` | Host *how* (ADR 014) |
| 3 | Ensure workspace | `yerk workspace ensure` | mkdir only (ADR 009) |
| 4 | Apply workspace locals | `yerk locals apply <id>` | new |
| 5 | Materialize replica | `yerk materialize` | git checkout (DWIM track separate) |
| 6 | Apply replica locals | `yerk locals apply <id> <replica>` | new |
| 7 | Bind style | side effect of mutators | ADR 013 |
| — | Sync | later | sync-verbs-design |
| — | Hooks | later | after 3, after 5/6 |

Composer (not v1 of this feature unless Q8 locks it):

```text
yerk project ready <id>          # 3–6 for default replica
yerk project ready <id> <rep>    # 3–4 if needed, ensure replica, 6
```

Name TBD (`ready` vs restoring `ensure`). Do not fold this into
`workspace ensure`.

### CLI (proposed)

```text
yerk locals apply <project-id>                 # workspace dest
yerk locals apply <project-id> <replica>       # replica dest
yerk locals apply yerk://domain/name/replica
yerk locals apply --all                        # workspace for each project
yerk locals apply --all --replicas             # every present replica (later if heavy)
```

Flags: `--force`, `--dry-run`. Opt-out on git mutators: `--no-locals`.

**Auto-apply (painkiller):** on successful `materialize` and
`replica create`, run replica apply for that distinguisher. Skip if
library empty. `workspace ensure` does **not** auto-apply (keeps mkdir
pure).

### Status / get

v1: no new status column. Failures are command errors. Follow-on: replica
status flag `locals-missing` / `locals-drift` once apply exists.

### Packages (implement plan, not this one)

```text
internal/locals   # library resolve, merge overlay, copy apply
CLI               # yerk locals apply
project.Resolver  # call apply after materialize / create
```

Tests: temp `YERK__CONFIG_DIR` only (ADR 007).

---

## Phases

### Phase 0 — Policy

Lock Q1–Q11 in [decisions.md](./decisions.md).

**Exit:** enough to write the ADR without re-asking mid-draft.

### Phase 1 — ADR

Draft **ADR 026** (number at write time; next after 025 if materialize-dwim
takes 025). Update domain spine: noun Local, verb apply locals, lifecycle
table. Amend ADR 003 path sketch. Cross-link ADR 007, 009, 013.

Operator accept → implement plan. Reject/revise stays in this folder.

### Phase 2 — Handoff

- Spawn `host-local-staging-implement` **or** backlog defer in owner todo.
- Do not add cobra commands in this plan.
- How-to stub only after the command ships (or a stub that does not claim
  a shipped verb).

---

## Success criteria

- [ ] Qs locked or explicitly deferred with default
- [ ] ADR accepted (or rejected with rationale)
- [ ] Domain spine lifecycle + locals noun/verb updated **when ADR accepted**
- [ ] No CLI stubs from this plan
- [ ] Owner todo reflects implement vs defer
- [ ] Plan folder removable at close

## Immediate next step

Phase 0: operator lock or flip the proposed bundle in
[decisions.md](./decisions.md) (include Q12 LFS skip-smudge).
