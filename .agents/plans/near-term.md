# Near-term implementation plan

Derived from [design/domain-and-near-term.md](../../design/domain-and-near-term.md)
and ADR 004. Prefer this over the older stub list in `todo.md` when they
conflict.

## Goal

Vertical slice for the operator’s near-term features:

1. Catalog listing with **presence** status  
2. Optional **change** status (`--git`)  
3. Workspace **ensure** (`yerk workspace ensure`; project workspace dir only)  
4. **Clone** default-branch replica (materialize replica)  
5. **Path resolve**

## Work packages

### P0 — Domain model in code (no network)

- [x] Split `internal/config`: tool `Config` vs `Catalog` / `Project` load from
      `config.toml` + `catalog.toml` (`YERK__CATALOG`).
- [x] Portable samples under `examples/` + CLI example subcommands (ADR 007).
- [x] Keep `internal/workspace` pure path math; tests for both styles + override.
- [x] Presence classifier: `missing` | `present` | `invalid`.

### P1 — Read model: `yerk status` / `yerk path`

- [x] `yerk path <project> [replica]`
- [x] `yerk status` — name, domain, replica, presence, path
- [x] Default replica: catalog override or fallback `main` (ls-remote in P2/clone)
- [x] Closed tag vocabulary + load validation (ADR 010)
- [x] `yerk status --tag` via `Catalog.SelectByTag` (`filter.tag=` line; unknown tag errors)

### P2 — Git adapter (subprocess)

- [x] `internal/gitcmd` interface + CLI implementation
- [x] Default branch via `ls-remote --symref` (clone path)
- [x] `yerk status --git` change probe
- [x] Serial probes first

### P3 — Materialize

- [x] `yerk workspace ensure [project…]` (project workspace dir only; ADR 009)
- [x] `yerk clone <project> [replica]`
- [x] Dropped host-local `fixtures/`; use `examples/` + temp/XDG only (ADR 007)

### P4 — Hardening

- [ ] Parallel change probes (errgroup / semaphore); unit test with fake adapter.
- [ ] `yerk pull` / `yerk push` only when implemented (no CLI stubs).
- [ ] Extend `--tag` selection to bulk mutate ops (`workspace ensure`, `clone`) using `SelectByTag`; define XOR with names/`--all`.
- [x] Refresh README + examples
- [ ] Optional: ADR note on git-subprocess vs go-git (pointer from design note).

## Non-goals this plan

- Local config staging  
- `register` write path (unless needed for tests)  
- Domain → bimtree path mapping  
- Daemon / FS watch  
- go-git migration  

## Suggested order of mergeable chunks

1. Catalog/config split + examples + tests  
2. `path` + presence-only `status`  
3. Git adapter + `--git` + default branch  
4. `workspace ensure` + `clone`  
5. Parallel probes

## Done when

Operator can hand-write `catalog.toml`, set `[domains]` roots, and run:

```text
yerk status
yerk path yerk
yerk workspace ensure yerk
yerk clone yerk
yerk status --git
```

…against at least one real remote on the host.
