# Checklist — initial-feature-series

Track execution here.
Do not put operator answers here (use [decisions.md](./decisions.md)).

## Phase 0 — Decisions

- [x] Q1–Q33 locked in [decisions.md](./decisions.md)
- [x] Open queue clear for implement
- [x] Plan folder + owner `todo.md` In Progress
- [x] Plan tracking files: [README.md](./README.md), this checklist

## Phase 1 — Identifiers / URI

- [x] ADR 012: grammar, `yerk://` scheme, ambiguity, domain required
- [x] Package `internal/id` parse / format / expand + tests
- [x] Catalog: require non-empty `domain`; resolve bare id / short / URI
- [x] `api.*` resources carry canonical URI (extend ADR 011)
- [x] Shared CLI helper: id **or** legacy positionals → project (+ optional replica)
- [x] Wire: `path`, `status`, `materialize`, `workspace ensure`
- [x] Docs: id vs path explanation stub (`docs/explanation/identifiers.md`)

## Phase 2 — Placement policy + host state

- [x] ADR: layers, XDG state vs config, dir-local walk, main replica (ADR 013)
- [x] ADR: path model successor to ADR 008 (host-local; optional `[domains]` default `<root>/<name>`; full path only when needed) (ADR 014)
- [x] Merge pipeline → effective placement (`internal/placement`)
- [x] State read/write under `$XDG_STATE_HOME/yerk/projects/<domain>/<project>/state.json`
- [x] Init binding on ensure / materialize (replica create later)
- [x] Warn ambient drift; error explicit CLI contradiction
- [x] Tests: precedence, conflicts, temp XDG + dir-local
- [x] Docs + portable examples only (ADR 007)
- [x] Env: `YERK__STATE_DIR` in envvars + `.appinfo`
- [x] Follow-up: domain-root default so most projects need no host path row
- [x] Follow-up: host dogfood `[domains]` + exception-only `[[projects]]`
- [x] Close-out tidy: stale ADR/docs refs, `*~` backups removed, repo-map packages

## Phase 3 — get + lookup

- [x] ADR/CLI: `get`, `project|replica get`, `lookup` / noun lookup (ADR 015)
- [x] Payloads: catalog + paths + presence + placement + URI (`ProjectInfo` / `ReplicaInfo`)
- [x] Reverse lookup: walk-up under workspace/replica
- [x] Human default; `--output json` when practical
- [x] Tests + agent-oriented docs

## Phase 4 — `config resolve`

- [x] Command: `yerk config resolve <target>`
- [x] Ordered contribution list + effective placement keys
- [x] Tests on fixture trees; docs “why is my style X?”
- [x] Follow-up: omit missing `state.json` from files list
- [x] Follow-up: `yerk state update` explicit rebind (not “lock”)

## Phase 5 — materialize + replica create

- [x] Rename `yerk clone` → `yerk materialize` (no alias)
- [x] Docs/help/tests/examples updated for materialize
- [x] ADR 016: `replica create` method, prerequisites, idempotency
- [x] `gitcmd` worktree add
- [x] `yerk replica create` worktree + clone-method paths
- [x] Hard error if main missing for worktree method
- [x] Tests with fake/real git as appropriate

## Phase 6 — Agent context dumps

- [x] `yerk context` / `yerk context dir [path]`
- [x] Compose get/lookup + resolve + static help
- [x] JSON stability note (`apiVersion: yerk/v1`) / ADR 017
- [x] Golden tests; short agents how-to

## Phase 7 — Additional workspace styles  **← next**

- [ ] `name-tags` path math (container workspace; bare main; `__` siblings)
- [ ] Parameterized inline `workspace_style` table (`style`, `main_dir`, `replica_dir`)
- [ ] Table-driven layout; tests; no host-private paths in-repo

## Phase 8 — Polish (interleave OK)

- [ ] `workspace ensure --tag` (XOR with names / `--all`)
- [ ] `--output json|yaml|table` shared flag surface
- [ ] Optional go-git ADR note (docs only)

## Phase 9 — Sync verbs (design only)

- [ ] Design note / ADR for `pull` / `push` (no CLI stubs)
- [ ] Implement only in a later plan after ADR accept

## Held — Parallel change probes

- [ ] Out of this plan; see owner [todo.md](../todo.md) Backlog

## Close-out

- [ ] Success criteria in [plan.md](./plan.md) checked
- [ ] Durable ADRs under `design/decisions/`; no `Q*` in product tree
- [ ] Design/docs updated (domain spine, commands, how-tos)
- [ ] Delete this plan folder
- [ ] Drop In Progress line in [../todo.md](../todo.md) (operator confirms)
