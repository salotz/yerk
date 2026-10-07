# Decisions inbox — materialize-dwim

Plan-local Q&A only. Durable outcome: ADR 025 under `design/decisions/`
(amends 016). Do **not** cite `Q*` in product code.

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q1 | open | Main missing, named replica requested: error / clone feat / auto-bootstrap main then worktree |
| Q2 | open | Force-clone hurdle: `--method clone`, `--clone` alias, extra `--force`? |
| Q3 | open | Catalog `replica_method` applies to materialize? (`clone` = always independent) |
| Q4 | open | Keep `yerk replica create` (fail-if-exists) or later deprecate |
| Q5 | open | Worktree branch source: local → origin/\<replica\> → new from main HEAD |
| Q6 | open | Bulk `materialize --all/--tag/--domain --replica feat`: same DWIM; auto-bootstrap mains? |
| Q7 | open | Paper: new ADR 025 vs only amend 016 |
| Q8 | open | Growth version: G (repair) vs R (behavior change of named materialize) |

## Locked

| ID | Decision |
|----|----------|
| L1 | **Option 2:** named replica + main **present** → **worktree**, not clone. |
| L2 | `materialize` is the **main / declarative ensure** verb (“do what I mean” given policy + disk). |
| L3 | Independent clone stays possible but is the **exception** (hurdle; Q2). |
| L4 | Destination **already present** → success; do not convert a clone into a worktree. |
| L5 | Materializing **main** (default replica) when missing still **clones** the hub. |
| L6 | Main need not have been seeded by yerk; presence = usable git checkout at the style path. |

### Tentative lean (not locked)

Offered 2026-10-05; operator asked for a plan before confirming the bundle:

- **Q1 C:** auto-bootstrap: clone main, then worktree the named replica.
- **Q2:** `--method clone` (same as create) plus optional short `--clone`. No extra `--force`.
- **Q3:** yes — `replica_method` applies; CLI method wins.
- **Q4:** keep `replica create` this plan (explicit, refuse-if-exists).
- **Q5:** (1) local `refs/heads/<replica>` attach; (2) else `refs/remotes/origin/<replica>` tracking that; (3) else new branch from main HEAD. No fetch in v1.
- **Q6:** same per-project DWIM on bulk; accept auto-bootstrap of mains if Q1 C.
- **Q7:** new **ADR 025**, amend 016 (create stays the strict spin-out).
- **Q8:** **G** (repair / robustness) unless named-materialize-always-clones is treated as a contract → then **R** while B=0.

---

## Constraints from product

- No unimplemented CLI stubs.
- Git via `gitcmd.Runner` (ADR 020); no ad-hoc shell-out in CLI.
- Tests use temp dirs only (ADR 007); do not commit host catalogs/paths.
- Project args still ADR 012 forms (`examol/darpa-nodes`, not `materialize examol md-next-workflow` as domain + replica).
- Do not cite plan-local Q ids in product code or ADRs as `Q*`.
