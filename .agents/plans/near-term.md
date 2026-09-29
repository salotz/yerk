# Near-term implementation plan

Derived from [design/domain-and-near-term.md](../../design/domain-and-near-term.md)
and ADR 004. Prefer this over the older stub list in `todo.md` when they
conflict.

## Goal

Vertical slice for the operator’s near-term features:

1. Catalog listing with **presence** (+ **change**) status  
2. Workspace **ensure** (`yerk workspace ensure`; project workspace dir only)  
3. **Clone** default-branch replica (materialize replica)  
4. **Path resolve**  
5. **Explicit API-style resources** (model-driven pilot) under a clear package  
6. Status UX: project vs replica scope, args, default change probe  

## Work packages

### P0 — Domain model in code (no network) ✅

- [x] Split `internal/config`: tool `Config` vs `Catalog` / `Project` load from
      `config.toml` + `catalog.toml` (`YERK__CATALOG`).
- [x] Portable samples under `examples/` (ADR 007; no CLI `example` subcommands).
- [x] Keep `internal/workspace` pure path math; tests for both styles + override.
- [x] Presence classifier: `missing` | `present` | `invalid`.

### P1 — Read model: `yerk status` / `yerk path` ✅ (baseline)

- [x] `yerk path <project> [replica]`
- [x] `yerk status` — catalog listing (presence; optional `--git`)
- [x] Default replica: catalog override or fallback `main` (ls-remote on clone)
- [x] Closed tag vocabulary + load validation (ADR 010)
- [x] `yerk status --tag` via `Catalog.SelectByTag`

### P2 — Git adapter (subprocess) ✅

- [x] `internal/gitcmd` interface + CLI implementation
- [x] Default branch via `ls-remote --symref` (clone path)
- [x] Change probe (today behind `status --git`)
- [x] Serial probes first

### P3 — Materialize ✅

- [x] `yerk workspace ensure <proj>…` | `--all` (workspace dir only; ADR 009)
- [x] `yerk clone <project> [replica]`
- [x] Examples + temp/XDG only (ADR 007)

### P5 — Explicit resources (model-driven pilot) ⬅️ foundation before more surface

Pilot a general “API resources” pattern even if not highest short-term ROI.
Goal: one typed model layer that CLI, serializers, schemas, and future
backends all share.

- [ ] Introduce `internal/api` (or `internal/resource`) with **stable resource
      types** (kubectl-style objects), distinct from TOML load structs and CLI
      row DTOs. Candidates:
      - `Catalog` / `Project` (registry)
      - `ProjectStatus` (project-scoped view: workspace path, presence of
        workspace, summary of replicas / default replica)
      - `ReplicaStatus` (replica-scoped: path, presence, change, branch, …)
      - later: `Domain`, `Config` snapshot, etc.
- [ ] Clear ownership: load/adapters **map into** resources; commands **print**
      resources; no business logic only in cobra `RunE`.
- [ ] Document type fields in Go (`//` + optional struct tags for json/yaml).
- [ ] ADR: model-driven resources + package layout (accept before large status
      rewrite locks ad-hoc `Row` shapes).
- [ ] Follow-ons (can land incrementally after types exist):
  - [ ] `--output json|yaml|table` (table remains default for humans)
  - [ ] JSON Schema (and/or OpenAPI-ish) generated or hand-maintained beside types
  - [ ] `yerk get <resource> [name…]` universal read (status may become a
        specialized view or alias over get)
  - [ ] Storage/backends later (TOML files today → optional DB/etc. without
        rewriting command logic)

### P6 — Status UX + change-on-disk (serial first) ✅

Align status with **project** vs **replica** and make change status first-class.

**Bugs / UX debt in current default list:**

- [x] Default `yerk status` path column emphasizes **project workspace**
      (`ProjectStatus.WorkspacePath`).
- [x] Drop **REPLICA** column from the default **project** list view (replica
      identity on replica-scoped status only).

**Invocation model:**

| Invocation | Scope |
| --- | --- |
| `yerk status` | All catalog projects (or `--tag`) — **project** status rows |
| `yerk status <project>` | Single **project** status |
| `yerk status <project> <replica>` | Single **replica** status |
| `yerk status --tag <name>` | Project list filtered by declared tag |

**Change status:**

- [x] **On by default** for status (disk/git change probe when presence allows).
- [x] **`--presence-only`** disables change probing (fast presence-only scans).
- [x] Serial change collection (`ProjectStatuses` / `ReplicaStatus`).
- [ ] Then **P7 parallelize** probes.

**Semantics (implemented):**

- Project rows: workspace path + default-replica presence/change summary
  (not a fake replica row; no REPLICA column).
- Replica rows: full path/presence/change/branch for that distinguisher.
- Multi-replica discovery under a workspace remains out of default project view.

### P7 — Parallel change probes ⬅️ next

- [ ] Parallelize change probes (errgroup / semaphore) **after** P6 serial path
      and resource types are stable.
- [ ] Unit test with fake git adapter (deterministic concurrency).

### P8 — Selection + bulk ops polish

- [x] `yerk clone --all` / `yerk clone --tag` via `SelectByTag` + XOR with
      names / `--all` (empty mutate selection errors).
- [ ] Extend `--tag` to `workspace ensure` (same XOR model).
- [ ] Optional: ADR note on git-subprocess vs go-git.

### P9 — Sync verbs (design before CLI)

Push/pull need non-confusing semantics (project vs replica, default replica,
tag bulk, dirty trees, upstream missing). **No CLI stubs** until decided.

- [ ] Design note / ADR: `pull` / `push` scope, safety, and selection.
- [ ] Implement only after design accept (single + tag; wire through resources).

## Non-goals this plan

- Local config staging  
- `register` write path (unless needed for tests)  
- Domain → bimtree path mapping  
- Daemon / FS watch  
- go-git migration  
- Shipping `pull`/`push` before sync semantics ADR  

## Suggested order (current)

1. ~~**P5** resource types + package + ADR~~ ✅
2. ~~**P6** status UX~~ ✅
3. **P7** parallel probes.
4. **P8** finish `--tag` on `workspace ensure` (clone bulk done).
5. Output formats / schema / `yerk get` as P5 follow-ons.
6. **P9** push/pull design → implement.

## Done when (rolling)

Operator can:

```text
yerk status                         # projects; change on unless opted out
yerk status yerk                    # one project
yerk status yerk main               # one replica
yerk path yerk
yerk workspace ensure yerk
yerk clone yerk
```

…and code has stable resource types suitable for table/json and later `get`.
