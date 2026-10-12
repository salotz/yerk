# Decisions inbox — host-local-staging

Plan-local Q&A only. Durable outcome: ADR under `design/decisions/` (likely
026; 025 reserved if materialize-dwim lands first).
Do **not** cite `Q*` in product code.

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q1 | proposed | Library path under XDG |
| Q2 | proposed | Host identity |
| Q3 | proposed | Host overlay vs required `<host>/` segment |
| Q4 | proposed | CLI names (`locals apply` vs `stage` / `inject`) |
| Q5 | proposed | Copy vs symlink |
| Q6 | proposed | Overwrite / tracked-file policy |
| Q7 | proposed | Auto-apply on materialize / replica create |
| Q8 | proposed | Composer (`project ready` / top-level `ensure`) |
| Q9 | proposed | Workspace library dest vs reserved `.local/yerk/` |
| Q10 | proposed | Missing library / missing host overlay |
| Q11 | proposed | `workspace ensure` auto-stage? |

## Locked

_(none yet)_

---

## Q1 — Library path under XDG

First pass: `$XDG_CONFIG_HOME/yerk/project-configs/<domain>/<project>/<host>/`.

ADR 003 sketch: `projects/<name>/locals/`.
State already uses `projects/<domain>/<project>/`.

**Proposed:** 

```text
$XDG_CONFIG_HOME/yerk/projects/<domain>/<name>/{workspace,replica,hosts/<host>/{workspace,replica}}/
```

No top-level `project-configs/`. No catalog paths. Portable samples later
under `examples/` with a fake host (ADR 007).

### Answer


### Notes

Agent lean: match state identity (`domain` + `name`); config = operator
files, state = tool bindings.

---

## Q2 — Host identity

Need a stable **host name** for overlays (laptop `mantis` vs HPC). Glossary
host name ≠ always the kernel hostname.

**Proposed:** resolve `YERK__HOST` → `config.toml` `[host] name` → short
hostname. Document as a product noun. Register `YERK__HOST` in envvars /
`.appinfo` when implementing.

### Answer


### Notes


---

## Q3 — Host overlay vs required path segment

First pass always had `…/<host>/`.

**Proposed:** common `workspace/` and `replica/` plus optional
`hosts/<host>/` overlay (host wins per relative path). HPC with no overlay
dir is a no-op, not an error.

### Answer


### Notes


---

## Q4 — CLI names

First pass: `stage` (fuzzy). Domain spine internal verb: stage locals.
Goals: apply local configuration. Git “stage” is this operator’s review
vocabulary.

**Proposed:** parent `yerk locals`; verb `apply`.

```text
yerk locals apply <project-id>              # workspace
yerk locals apply <project-id> <replica>    # replica
```

No top-level `stage` / `inject` in v1. Aliases only if you want them.

### Answer


### Notes


---

## Q5 — Copy vs symlink

Replica inject can copy bytes or link to the library / workspace `.local`.
Symlink: one edit, all worktrees; dangling if XDG moves. Copy: simple;
drift if someone edits one replica.

**Proposed:** v1 **copy** only. Symlink / `--link` follow-on.

### Answer


### Notes


---

## Q6 — Overwrite policy

**Proposed:** write if missing; identical bytes = success; different
**untracked** = error unless `--force`; **tracked** = always error; never
`git add`. `--dry-run` lists planned writes.

### Answer


### Notes


---

## Q7 — Auto-apply on replica mutators

The pain is forgetting inject after worktrees.

**Proposed:** after successful `materialize` and `replica create`, apply
replica locals for that distinguisher. `--no-locals` opt-out. Empty library
= skip (success).

### Answer


### Notes


---

## Q8 — Lifecycle composer

First pass: bundle into `ensure`. ADR 009 removed top-level `ensure` and
locked `workspace ensure` = mkdir only. Materialize-DWIM wants `materialize`
as declarative replica ensure.

**Proposed:** document the lifecycle in the ADR **now**. Do **not** ship a
composer in the first implement plan. Candidate later: `yerk project ready`
(not a quiet revival of `yerk ensure` without an ADR that amends 009).

### Answer


### Notes


---

## Q9 — Workspace destination for staged library

First pass: `<project>/.local/config/yerk/…` and `<project>/mise.toml`.
ADR 013 reserves `<dir>/.local/yerk/config.toml` for placement.

**Proposed:** workspace-scoped library files go exactly where `workspace/`
relative paths say (so `mise.toml` at workspace root is allowed).
Recommend operators put yerk-owned extras under `.local/config/yerk/` not
`.local/yerk/`. Apply never writes placement `config.toml`.

### Answer


### Notes


---

## Q10 — Missing overlay

**Proposed:** no `projects/<domain>/<name>/` tree, or no matching
`hosts/<this>/` and no common files → apply succeeds, writes nothing.
Unknown project id is still an error (catalog resolve).

### Answer


### Notes


---

## Q11 — Auto-apply on `workspace ensure`

**Proposed:** **no**. Keep ADR 009 (mkdir only, no git, no extra file
policy). Workspace locals are `yerk locals apply <id>` or a future
composer.

### Answer


### Notes


---

## Constraints from product

- No unimplemented CLI stubs.
- Catalog remains portable (no host file blobs, no `path`).
- Tests: temp dirs only (ADR 007).
- `workspace ensure` scope is ADR 009 unless this ADR explicitly amends it.
- Placement dir-local `.local/yerk/config.toml` (ADR 013) is not the locals
  library.
- Do not cite plan-local Q ids in product code or ADRs as `Q*`.
- Git adapter stays subprocess (ADR 020); locals apply is filesystem, not
  git, except a tracked-file check.

## Changelog

- 2026-10-07 — opened Q1–Q11 with proposed defaults from first-pass iteration.
- 2026-10-07 — Q12 Git LFS skip-smudge (git-config local, host flag).
