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

- [ ] ADR: layers, XDG state vs config, dir-local walk, main replica
- [ ] ADR: path model successor to ADR 008 (host-local; no domain-root join)
- [ ] Merge pipeline → effective placement
- [ ] State read/write under `$XDG_STATE_HOME/yerk/projects/<domain>/<project>/state.json`
- [ ] Init binding on ensure / materialize / replica create
- [ ] Warn ambient drift; error explicit CLI contradiction
- [ ] Tests: precedence, conflicts, temp XDG + dir-local
- [ ] Docs + portable examples only (ADR 007)
- [ ] Env: `YERK__STATE_DIR` in envvars + `.appinfo`

## Phase 3 — get + lookup

- [ ] ADR/CLI: `get`, `project|replica get`, `lookup` / noun lookup
- [ ] Payloads: catalog + paths + presence + placement + URI
- [ ] Reverse lookup: walk-up under workspace/replica
- [ ] Human default; `--output json` when practical
- [ ] Tests + agent-oriented docs

## Phase 4 — `config resolve`

- [ ] Command: `yerk config resolve <target>`
- [ ] Ordered contribution list + effective placement keys
- [ ] Tests on fixture trees; docs “why is my style X?”

## Phase 5 — materialize + replica create

- [x] Rename `yerk clone` → `yerk materialize` (no alias)
- [x] Docs/help/tests/examples updated for materialize
- [ ] ADR: `replica create` method, prerequisites, idempotency
- [ ] `gitcmd` worktree add
- [ ] `yerk replica create` worktree + clone-method paths
- [ ] Hard error if main missing for worktree method
- [ ] Tests with fake/real git as appropriate

## Phase 6 — Agent context dumps

- [ ] `yerk context` / `yerk context dir [path]`
- [ ] Compose get/lookup + resolve + static help
- [ ] JSON stability note (`apiVersion: yerk/v1`)
- [ ] Golden tests; short agents how-to

## Phase 7 — Additional workspace styles

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
