# Plan: initial feature series

## Context

Repository: **yerk** (`salotz.yerk`) — host multi-project manager (catalog,
replicas, status, path resolve).

Design spine (durable): [design/domain-and-near-term.md](../../../../design/domain-and-near-term.md),
[design/goals.md](../../../../design/goals.md), ADRs under
[design/decisions/](../../../../design/decisions/).

This plan coordinates the **next product feature series** after the first
vertical slice already on `main` (config/catalog split, status/path, ensure,
clone, api resources, change-on-default, origin comparison fallback).

Operator work-process (when agent-guidelines is checked out beside this
tree): `agent-guidelines` → `content/personal/work-process.md` (plan
folder layout, `decisions.md` inbox, owner `todo.md`). Plan-local Q&A
lives only in [decisions.md](./decisions.md).

**Rename note:** formerly tracked as ad-hoc `.agents/plans/near-term.md`.
Folder name is now **`initial-feature-series`**.

---

## Goals

1. **Layered placement policy** — per-tree / per-project workspace style and
   related knobs; host init state; warn/error conflict rules; XDG split.
2. **Main replica** vocabulary as hub for worktrees and later host-local materials.
3. **Replica spin-out** — `yerk replica create` with `replica-method`
   (`worktree` | `clone`) for agent session managers.
4. **Yerk identifiers / URIs** — layout-independent ids; `yerk://…` as internal
   currency; short forms expand.
5. **Project/replica info** — forward `get` and reverse path `lookup` for tools.
6. **Config contribution resolve** — which files shape a target’s settings.
7. **Agent context dumps** — tool + directory context for AI agents.
8. Additional **workspace styles** once named; bulk/ensure polish; output formats;
   sync design later. **Parallel probes held.**

Non-goals (this plan): locals staging product, `register` UX, bimtree domain
mapping, daemon/FS watch, go-git migration, shipping `pull`/`push` before sync
ADR, OS `yerk://` handler, parallel probe implementation until reprioritized.

---

## Already shipped (baseline; do not re-litigate)

| Area | Status |
|------|--------|
| P0 domain model / config+catalog | done |
| P1 status/path presence baseline | done |
| P2 git adapter + change probe + origin fallback | done |
| P3 workspace ensure + clone | done |
| P5 `internal/api` + ADR 011 | done (follow-ons open) |
| P6 status UX + change default | done |
| ADR 010 tags; clone `--all`/`--tag` | done |

Deferred by operator: **parallel change probes** (old P7).

---

## Architecture targets (after series)

```text
CLI
  → parse id | path | positional  →  canonical yerk:// URI (P13)
  → load catalog + layered config + host project state (P10)
  → map to internal/api resources (identity + observed fields)
  → layout (effective style) / git adapter (clone, worktree, probe)
  → print (table today; json|yaml with output follow-on)
```

Identity ≠ filesystem path. Config layers ≠ single host `config.toml`.

---

## Phase order

| Phase | Name | Depends |
|-------|------|---------|
| 0 | Lock decisions | — |
| 1 | Identifiers / URI (P13) | Phase 0 subset |
| 2 | Placement policy + host state (P10) | Phase 0 subset; prefers Phase 1 for state keys |
| 3 | Project/replica get + lookup (P14) | Phase 1; Phase 2 for full effective paths |
| 4 | Resolve config stack (P15) | Phase 2 |
| 5 | Replica create + method (P11) | Phase 2; Phase 1 for targets |
| 6 | Agent context dumps (P16) | Phase 3–4 |
| 7 | New workspace styles (P12) | Phase 2 + style Qs locked |
| 8 | Polish: ensure `--tag`, output formats | can interleave |
| 9 | Sync verbs design only (P9) | after core UX |
| — | Parallel probes | **out of execute queue** until operator asks |

Execution: one step at a time on “go” / “execute next step”; no commits unless
asked; answers only in `decisions.md`.

---

## Phase 0 — Decisions

Record and lock answers in [decisions.md](./decisions.md) before large ADRs or
code that freezes precedence.

Open queue summary lives in `decisions.md` (source of truth for status).

Draft ADRs to land under `design/decisions/` when choices stabilize (not plan-local
`Q*` citations in product docs):

- Layered placement + XDG state vs config + dir-local discovery
- Identifier grammar + `yerk://` scheme
- Replica method + create semantics
- (as needed) new workspace styles, get/lookup CLI, context dump stability

**Exit:** critical Qs locked or explicitly deferred with defaults; open queue clean enough to implement Phase 1–2.

---

## Phase 1 — Yerk identifier syntax + URI currency

**Problem.** Commands and tools need stable names independent of workspace
layout. Path `~/tree/personal/devel/wumpus/main` is not the identity.

### Intent

| Form | Example |
|------|---------|
| Project | `personal/wumpus` |
| Replica | `personal/wumpus/main` |
| Short | `wumpus`, `wumpus/main` if unique |
| Canonical | `yerk://personal/wumpus/main` |

Short forms **expand** to the full URI. Internal APIs and machine output use
URI only. Commands accept identifiers **and** legacy positionals.

```bash
yerk get personal/wumpus
yerk get personal/wumpus/main
yerk get wumpus/main
yerk get yerk://personal/wumpus/main
yerk status personal/wumpus
yerk path yerk://personal/wumpus/main
```

### Work

1. ADR: grammar, scheme, ambiguity, forbidden chars, domain rules.
2. Package parse/format/expand (`internal/id` or similar) + tests.
3. Canonical URI field on `api.*` resources (extend ADR 011).
4. Shared CLI helper: id **or** positional pair → URI → dispatch.
5. Docs: id vs path explanation.

### Exit

Parse/expand covered by tests; at least one command accepts id form; resources
can carry URI.

---

## Phase 2 — Placement policy (layered style + host state)

**Problem.** Host-wide `[workspace].style` is too coarse. Need tree defaults
(e.g. devel vs admin), sticky host binding after init, and clear conflicts.

### Intent

Layers (target): built-in → host `config.toml` → domain → dir-local
`.local/yerk/config.toml` → catalog row → **host project state** → env → CLI.

```toml
# e.g. ~/tree/personal/admin/.local/yerk/config.toml
[workspace]
style = "name-tags"   # locked name; path math in decisions
```

- Ambient drift vs bound state → **warning** (exact winner: Q).
- Explicit CLI contradiction → **error**.
- Init writes binding under **`$XDG_STATE_HOME/yerk`** (not overwriting hand-edited config home).
- **Main replica**: hub checkout (extends `default_replica`); worktrees spin from it.

### XDG guidance (promote to ADR)

| Tree | Role |
|------|------|
| `$XDG_CONFIG_HOME/yerk` | Operator prefs + catalog |
| `$XDG_STATE_HOME/yerk` | Tool-written project bindings / state |
| data / cache | optional later |

Dir-local `.local/yerk/` is tree-scoped (not XDG); discovery rules in ADR.

### Work

1. ADR: layers, conflicts, XDG, dir-local walk, main replica, catalog fields.
2. Merge pipeline → effective placement consumed by layout.
3. State read/write; init path; CLI flags + warn/error.
4. Tests: precedence, conflicts, temp XDG + dir-local trees.
5. Docs + portable examples only (ADR 007).

### Exit

Effective style resolves with tests; state round-trip; unknown future styles
error cleanly until Phase 7.

---

## Phase 3 — Project / replica info (`get` + `lookup`)

**Problem.** Session managers need uniform info (catalog + host paths + ids),
not only the last path segment of cwd.

### Intent

```bash
yerk project get <id-or-name>
yerk replica get <project> <replica>   # or replica id
yerk project lookup <path>
yerk replica lookup <path>
yerk get personal/wumpus[/main]        # universal read
```

`get` = id → resource. `lookup` = path → id + resource.

### Work

1. CLI/ADR: noun commands vs only `get`/`lookup` (see decisions).
2. Payloads: catalog + paths + presence + effective placement + URI.
3. Reverse lookup: exact roots and/or walk-up (Q).
4. Tests; docs for agents/session managers.
5. Prefer early `--output json` if agents need it (Phase 8 can land earlier).

### Exit

Forward and reverse work on temp fixtures; structured output usable by a script.

---

## Phase 4 — Host config layout resolve

**Problem.** Layered config needs observability.

### Intent

```bash
yerk resolve config <project-or-replica-path-or-identifier>
```

Emit contribution stack, e.g. catalog, host config, each dir-local, project
state, env/CLI overlays; order; optional per-key provenance / effective snapshot;
drift warnings.

### Work

1. Same merge pipeline as Phase 2; resolve is the debug surface.
2. Api kinds for contribution / effective config as needed.
3. CLI + tests; docs “why is my style X?”.

### Exit

Known fixture tree prints expected file list and winners.

---

## Phase 5 — `materialize` + replica spin-out

**Problem.** Agents need yerk-mediated replica creation without embedding git
worktree vs clone policy; product CLI must not overload the word `clone`.

### Intent

```bash
yerk materialize <project-id> [replica]   # from remote (ex-yerk clone); bulk --all/--tag
yerk replica create <project-id> <replica-name>
# --method worktree|clone, style/method flags per decisions
```

| Method | MVP behavior |
|--------|----------------|
| `worktree` | `git worktree add` from **main** replica to style path |
| `clone` | `git clone` at style path (method name only; not a top-level command) |

**Q12:** rename top-level `clone` → **`materialize`** immediately, no alias.

### Work

1. Rename CLI `clone` → `materialize` (help, docs, tests, examples).
2. ADR: method, prerequisites, flags, idempotency for `replica create`.
3. `gitcmd` worktree add; wire create through resolve + api.
4. Tests with fake git; docs for session spin-out vs materialize-from-remote.

### Exit

`materialize` dogfoods; create worktree and clone-method paths work in tests.

---

## Phase 6 — Context dumps for AI agents

### Intent

```bash
yerk context              # tool: vocabulary, commands, XDG, how-to
yerk context dir [path]   # directory: ids, paths, short config/status
```

Compose Phase 3–4 + static help; structured mode for agents.

### Work

1. CLI naming per decisions; stability note for JSON.
2. Implement without duplicating business logic.
3. Golden tests; short agents how-to.

### Exit

`context` and `context dir` useful on dogfood host without scraping markdown.

---

## Phase 7 — Additional workspace styles

Depends on Phase 2 + locked style names/math.

- **name-extensions** (working title) for trees like `…/admin`.
- **singleton / inplace** (or similar) for special homes (`~/.bimker`).
- Table-driven layout; tests; no host-private paths in-repo.

---

## Phase 8 — Polish (can interleave)

- [ ] `workspace ensure --tag` (same XOR as clone).
- [ ] `--output json|yaml|table`; schemas as needed.
- [ ] Optional go-git ADR note (docs only).

---

## Phase 9 — Sync verbs (design only)

- [ ] Design note / ADR for `pull` / `push` (no CLI stubs until accept).
- [ ] Implement only in a later plan after ADR.

---

## Held — Parallel change probes

Serial probes remain. Parallelize (errgroup/semaphore + fake adapter tests)
only when operator reprioritizes.

---

## Risks and mitigations

| Risk | Mitigation |
|------|------------|
| Implementing layout before URI/state keys | Phase 1 before or lockstep with Phase 2 |
| Silent relocation of replicas | Bound state + warn/error; no ambient restyle |
| Confusing clone vs replica create | Document option A; ADR |
| Agent JSON churn | Freeze keys in ADR when shipping context/get |
| Plan Q ids leaking into code | Promote to ADRs; never cite `Q*` in product tree |
| Host paths in examples | ADR 007; temp dirs only |

---

## Success criteria

- [ ] Canonical `yerk://domain/project[/replica]`; short ids expand; ambiguity errors.
- [ ] Layered style/method with dir-local + XDG state; conflicts warn/error as locked.
- [ ] `project`/`replica` get + path lookup (and/or universal `get`) for tools.
- [ ] `resolve config` shows contribution stack for a target.
- [ ] `replica create` via worktree or clone using effective policy.
- [ ] `context` / `context dir` dumps for agents.
- [ ] New styles implemented once named; ensure `--tag` when scheduled.
- [ ] Durable choices promoted to `design/decisions/` ADRs; plan folder removable at close.

---

## Understanding check (feature requests → phases)

| Request | Phase |
|---------|--------|
| Layered / per-repo workspace style | 2 |
| Dir-local `.local/yerk/config.toml` | 2 |
| Init host-local binding; XDG config vs state | 2 |
| Ambient warn / CLI error | 2 |
| Main replica hub | 2, 5 |
| `replica create` + `replica-method` | 5 |
| Hold parallelization | Held |
| Id `domain/project[/replica]` + `yerk://` | 1 |
| Commands take ids | 1 |
| project/replica get + lookup | 3 |
| Config files contribution dump | 4 |
| Agent context dumps | 6 |

---

## Execution protocol

- Answers **only** in [decisions.md](./decisions.md) (`Q*`).
- Chat: short open-queue ping; not the answer archive.
- One execute step per “go” unless operator widens scope.
- No git stage/commit unless asked.
- After a step: draft commit message, test commands, truncated agent test output.
- Close-out: ADRs + design/docs updated; delete this plan folder; drop In Progress
  line in [../todo.md](../todo.md).

## Immediate next step

**Done this chunk:** Phase 6 — ADR 017; `yerk context` + `context dir`;
ToolContext / DirContext API; compose lookup + placement + short status;
tests + agents how-to. (Also earlier: multi-replica status overall.)

**Next (default on go):** **Phase 7** — additional workspace styles
(`name-tags` path math + parameterized table), **or** Phase 8 polish
(`workspace ensure --tag`, shared `--output`).

Alternates if operator widens scope:

1. Phase 8 slice — `workspace ensure --tag` / broader `--output` formats  
2. Phase 9 — sync verbs design only  
3. Held — parallel change probes (only if reprioritized)
