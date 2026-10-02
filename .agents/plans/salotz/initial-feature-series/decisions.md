# Decisions

Single inbox for operator↔agent prompts on this plan
(**initial-feature-series**).

Plan-local ids only (`Q1`, `Q2`, …) — do **not** cite these in durable repo
docs, code, or ADRs as the sole record. Promote lasting choices to
`design/decisions/` ADRs at lock/close-out.

Process: agent-guidelines personal **work-process** (Q-record shape).
Statuses: `open` | `proposed` | `locked`.

---

## Open queue

| ID | Status | Need |
|----|--------|------|
| Q1 | locked | Bound state vs ambient config when they disagree |
| Q2 | locked | What counts as project “initialize” (writes host state) |
| Q3 | locked | Dir-local `.local/yerk/` discovery anchor and stop rules |
| Q4 | locked | Domain-level config shape (moot; domains ≠ path roots for style) |
| Q5 | locked | `YERK__WORKSPACE_STYLE` ambient vs explicit |
| Q6 | locked | Catalog field names (style / method / main replica) |
| Q7 | locked | Host project state path layout under XDG state |
| Q8 | locked | `name-tags` style path math (suffix config later) |
| Q9 | locked | Special homes via style params / derivations (syntax later) |
| Q10 | locked | Main missing + `worktree` create behavior |
| Q11 | locked | Replica distinguisher vs git branch/ref |
| Q12 | locked | Rename today’s `yerk clone` → `yerk materialize` (no alias) |
| Q13 | locked | Idempotent `replica create` if path exists |
| Q14 | locked | Which commands take style/method CLI flags |
| Q15 | locked | Identifier forms: bare path + one canonical `yerk://` URI |
| Q16 | locked | Domain required on every project / canonical URI |
| Q17 | locked | Ambiguous short project name |
| Q18 | locked | Project URI vs default-replica when getting by id |
| Q19 | locked | State directory key encoding (with Q7 → `state.json`) |
| Q20 | locked | Get/lookup command tree shape |
| Q21 | locked | Lookup: any subdirectory under workspace/replica |
| Q22 | locked | Default output for get/lookup (human vs JSON) |
| Q23 | locked | `resolve config` dump depth |
| Q24 | locked | Command name: `yerk config resolve` (not `resolve config`) |
| Q25 | locked | Context dump names (revisit later OK) |
| Q26 | locked | JSON key stability promise for agents (v0) |
| Q27 | locked | XDG split for config vs state |
| Q28 | locked | Parallel probes remain deferred |
| Q29 | locked | Path is host-local; optional `[domains]` default `<root>/<name>`; full `[[projects]]` path when needed (ADR 014) |
| Q30 | locked | Catalog identity: `name` + required `domain` → id `domain/name` |
| Q31 | locked | Merge into Q12: command name `yerk materialize` |
| Q32 | locked | Inline table only for parameterized `workspace_style` (no named derivations) |
| Q33 | locked | `name-tags` workspace = parent container of bare main + siblings |

---

## Q1 — Bound state vs ambient winner

Status: locked

### Prompt

When host project state says style `workspace-dir` but dir-local or catalog now
says another style, what should yerk do on materialize/resolve?

CLI flags that contradict state remain **error** (already agreed in intent).

### Answer

Keep **bound state** as effective style for path math; emit a **warning**
naming the conflicting ambient sources; require an explicit rebind command (or
documented state edit + verify) to change binding. Do not hard-stop all
commands on ambient drift.

### Notes

Operator accepted 2026-09-30.

---

## Q2 — What counts as “initialize”

Status: locked

### Prompt

When is host project state first written?

### Answer

Write/update binding on first successful **`workspace ensure`** for that
project **or** on first successful materialize under the project (`clone` /
renamed bootstrap, or `replica create`). Ensure alone may bind style before any
git.

- Later optional `yerk project init`: **fail-if-exists** (explicit).
- **`workspace ensure`** on an already-initialized project: **no-op** (success).

### Notes

Operator refined proposal 2026-09-30 (`or` not only ensure; init fail-if-exists;
ensure no-op if initialized).

---

## Q3 — Dir-local discovery

Status: locked

### Prompt

From which path do we walk for `.local/yerk/config.toml`, and where do we stop?

### Answer

1. Anchor = resolved **project workspace** path when the target is a known
   project; else absolute path argument; else **cwd**.
2. Walk **up** toward **user home**, collecting each
   `<dir>/.local/yerk/config.toml`.
3. Merge: **near wins** for the same key (closer to anchor overrides farther).
4. **Stop at user home** (do not walk above `$HOME`).

### Notes

Operator 2026-09-30: stop at home (not domain root). Near-wins locked for ADR.

---

## Q4 — Domain config shape

Status: locked

### Prompt

How do domain-level defaults appear?

### Answer

**Moot for placement.** Domains are a **logical namespace** for projects, not a
required mapping onto a single filesystem root for workspace style.

Domain-level configuration (if any later) must **not** drive host-local
placement knobs (style, method, etc.). Those stay host/dir-local/state/CLI.

### Notes

Operator 2026-09-30. Domain is namespace only for placement/style. Filesystem
path rules locked in **Q29** (host-local / style dirs; ADR 008 join superseded).

---

## Q5 — `YERK__WORKSPACE_STYLE` strength

Status: locked

### Prompt

Is the existing env overlay ambient (warn on state conflict) or explicit
(error like CLI)?

### Answer

**Ambient** (same class as files): warn on conflict with bound state; useful for
CI without implying intent as strong as `--workspace-style`.

### Notes

Operator accepted 2026-09-30.

---

## Q6 — Catalog field names

Status: locked

### Prompt

Names for optional per-project overrides and main replica field.

### Answer

- `workspace_style` (optional on project row)
- `replica_method` (optional)
- keep `default_replica`; prose synonym **main replica** until a rename is worth it

### Notes

Operator accepted 2026-09-30. Catalog **row identity** shape may still change
(`id = "personal/wumpus"` mock in Q9) — see **Q30**.

---

## Q7 — Host project state layout

Status: locked

### Prompt

File layout under `$XDG_STATE_HOME/yerk` for per-project bindings.

### Answer

Per-project directory; **JSON** state (stdlib-friendly, machine-oriented):

```text
$XDG_STATE_HOME/yerk/projects/<domain>/<project>/state.json
```

(`<domain>/<project>` nesting from Q19.) Not a monolithic `projects.toml`.

Override dir: `YERK__STATE_DIR` (register in envvars / `.appinfo`).

### Notes

Operator 2026-09-30: **`state.json`** not TOML. Q19 path nesting applies; older
Q19 sample said `state.toml` — superseded by this lock.

---

## Q8 — `name-tags` style (was “name-extensions”)

Status: locked

### Prompt

Exact **name** and **path math** for admin-like trees.

### Answer

- Style name: **`name-tags`** (defer shipping if needed; name is locked).
- **Main replica** path = project bare directory name under the style’s base,
  e.g. main for `wumpus` → `…/projects/wumpus` (directory name = project name).
- **Other replicas** sit **beside** main with a distinguisher suffix, default
  separator `__`: e.g. `…/projects/wumpus__feat1` for replica `feat1`.
- Making the separator configurable is an **even later** feature.

### Notes

Operator 2026-09-30. **Q33**: what `workspace ensure` / project-workspace path
mean under this style (main dir vs parent `projects/`). **Q9** / **Q32**:
parameterized bases (`main_dir`, `replica_dir`).

---

## Q9 — Special home repos (e.g. `~/.bimker`)

Status: locked

### Prompt

(Reworded after operator feedback.) How do projects whose **main** checkout
lives in a special directory (e.g. `~/.bimker`) while **other replicas** live
elsewhere (e.g. under a devel tree) fit the style system?

### Answer

Treat as **`name-tags` (or similar) with parameters**, not a wholly separate
one-off style:

- Optional knobs such as `main_dir` vs `replica_dir` (mock only):

  ```toml
  # mock — final syntax in Q32
  workspace_style = { name = "name-tags", main_dir = "~/.bimker", replica_dir = "~/tree/personal/devel/bimker" }
  ```

- Each style should support **named derivations** and/or **inline** parameter
  blocks for specific repos.

### Notes

Operator 2026-09-30. Syntax not locked — **Q32**. Agent prompt quality note
acknowledged.

---

## Q10 — Main missing + worktree create

Status: locked

### Prompt

`yerk replica create --method worktree` when main replica is missing/absent?

### Answer

**Hard error** with actionable message (“bootstrap/ensure main first”). No
auto-clone in v1.

### Notes

Operator accepted 2026-09-30.

---

## Q11 — Replica name vs branch

Status: locked

### Prompt

Is replica distinguisher always the git branch name, or can `--branch`/`--ref`
differ?

### Answer

Default **replica name = branch name**. Optional `--branch` later if
distinguisher must differ — not MVP.

### Notes

Operator accepted 2026-09-30.

---

## Q12 — Today’s `yerk clone` naming

Status: locked

### Prompt

Keep `clone` as bootstrap-from-remote, alias it, or **rename** so it does not
collide mentally with `replica create`?

### Answer

Rename immediately to **`yerk materialize`**.

- No deprecation period, no hidden `clone` alias.
- Bulk flags (`--all` / `--tag`) move with it.
- `replica create` remains session spin-out (worktree/clone **method**).
- Internal/docs verb “materialize replica (from remote)” matches.

### Notes

Operator 2026-09-30 (via Q31). Git subprocess still runs `git clone`; product
CLI must not say `yerk clone`.

---

## Q13 — Idempotent create

Status: locked

### Prompt

If destination path already exists / replica present?

### Answer

**Refuse** non-empty / already-present path (match current clone safety).
Optional later `--if-absent`.

### Notes

Operator accepted 2026-09-30.

---

## Q14 — Flags on which commands

Status: locked

### Prompt

Which commands accept `--workspace-style` / `--replica-method`?

### Answer

- **Mutate placement:** renamed bootstrap-from-remote, `replica create`,
  `workspace ensure` (style only where relevant).
- **Read-only:** `status`, `path`, `get`, `lookup`, `config resolve`,
  `context` — **report** effective policy only; do not change binding.
- Explicit flags on mutate that contradict bound state → **error**.
- Explicit **migration / rebind** commands = later feature.

### Notes

Operator accepted 2026-09-30 (+ migration later).

---

## Q15 — Identifier / URI forms

Status: locked

### Prompt

Canonical URI form; `yerk://` vs `yerk:`; how many spellings?

### Answer

**Two forms only:**

1. **Bare identifier** (CLI ergonomics): `personal/wumpus`,
   `personal/wumpus/main` (and unique short names where allowed).
2. **One canonical URI:** `yerk://personal/wumpus` /
   `yerk://personal/wumpus/main`.

No third form (`yerk:…` without `//`). Query/fragment rejected in v1.
Everything internal expands bare → canonical URI.

### Education (agent → operator)

In URI syntax (RFC 3986):

| Form | Shape | Meaning |
|------|--------|---------|
| `yerk://personal/wumpus/main` | scheme + `//` + **hierarchical** path | Standard “URL-like” URI. Path segments are `personal`, `wumpus`, `main`. Easy to join with normal path rules. |
| `yerk:personal/wumpus/main` | scheme + **opaque** rest (no `//`) | Not hierarchical to generic URI parsers; the part after `:` is one opaque string. Different libraries treat it differently than `//` form. |

So `yerk://…` is the usual choice for path-like resource ids. `yerk:…` would be
a second, easier-to-get-wrong spelling — you correctly prefer **not** to have it.

Bare `personal/wumpus/main` is **not** a URI until expanded; it is a yerk
**shortcut** for humans/CLIs.

### Notes

Operator 2026-09-30: prefer protocol-less ids + one canonical URI. Proposal
trimmed accordingly.

---

## Q16 — Domain required

Status: locked

### Prompt

Projects with empty/missing domain? Optional domains?

### Answer

**Domain is required** on every project. Canonical ids/URIs **always** include
domain. Empty domain → **loud error** at load or expand time. No implicit
default domain.

### Notes

Operator 2026-09-30. Catalog validation should enforce non-empty `domain` (or
equivalent in `id` — Q30). Interacts with **Q29** (domain no longer “the”
filesystem root key for placement, but still mandatory namespace).

---

## Q17 — Ambiguous short name

Status: locked

### Prompt

Two catalog projects named `wumpus` in different domains; user passes `wumpus`.

### Answer

**Error** listing candidates (`personal/wumpus`, `work/wumpus`). Optional later
disambiguators.

### Notes

Operator accepted 2026-09-30.

---

## Q18 — Project get vs default replica

Status: locked

### Prompt

Does `yerk get personal/wumpus` mean project resource only, or default replica?

### Answer

**Project resource** only. Replica needs explicit third segment or `replica get`.
No auto-expand to main.

### Notes

Operator accepted 2026-09-30.

---

## Q19 — State directory key

Status: locked

### Prompt

How to encode project identity in state path (with Q7)?

### Answer

```text
$XDG_STATE_HOME/yerk/projects/<domain>/<project>/state.json
```

Nested segments matching URI path; filesystem-safe names via catalog validation.

### Notes

Aligned with Q7 (`state.json`). Operator’s earlier nested TOML sample superseded
for file format only.

---

## Q20 — Get / lookup command tree

Status: locked

### Prompt

Ship noun commands, universal get, or both?

### Answer

**Both:** `yerk get <id>`; `yerk project|replica get`; `yerk project|replica
lookup` and/or `yerk lookup <path>` that dispatches.

### Notes

Operator accepted 2026-09-30.

---

## Q21 — Lookup path depth

Status: locked

### Prompt

Exact root only vs any subdirectory?

### Answer

Accept **any path under** a known replica or workspace (walk up to match).

### Notes

Operator accepted 2026-09-30.

---

## Q22 — Default output for get/lookup

Status: locked

### Prompt

Human text/table default vs JSON-first?

### Answer

Human-friendly default on TTY; **`--output json`** for agents as soon as
practical. Non-TTY → json later optional. Future `TTY_*` standards out of scope
until they exist.

### Notes

Operator accepted 2026-09-30.

---

## Q23 — Config resolve dump depth

Status: locked

### Prompt

File list only vs per-key provenance vs full merged doc?

### Answer

v1 = **ordered file list** + **effective key snapshot** for placement-related
keys. Full provenance matrix later if needed.

### Notes

Operator accepted 2026-09-30. Command name **Q24**.

---

## Q24 — Config resolve command name

Status: locked

### Prompt

`yerk resolve config` vs other?

### Answer

**`yerk config resolve`** (config parent, resolve subcommand). Target-scoped
contribution stack. Host-global dump stays under `yerk config show` (or
equivalent).

### Notes

Operator 2026-09-30. Update plan CLI sketches accordingly.

---

## Q34 — Explicit state rebind command name

Status: locked

### Prompt

Name for “regenerate / refresh host project bindings from current ambient”?
Candidates included lock-like verbs.

### Answer

**`yerk state update [project…|--all]`** with optional `--workspace-style`.

- Parent noun **`state`** matches XDG state / `state.json`.
- Verb **`update`** = create or overwrite binding from ambient (or explicit style).
- Reject **lock** (overloaded: VCS locks, package locks, freeze semantics).
- First bind remains ensure/materialize no-op-if-exists; update is the explicit
  rebind path (ADR 013 amendment).

### Notes

Operator 2026-10-01. Also: `config resolve` must not list missing state paths
under `files` (contribution note only).

---

## Q25 — Context dump command names

Status: locked

### Prompt

Naming for agent context surface?

### Answer

```text
yerk context
yerk context dir [path]
```

Revisit later if needed; fine for now.

### Notes

Operator 2026-09-30.

---

## Q26 — Agent JSON stability (v0)

Status: locked

### Prompt

How hard freeze JSON keys for external agents?

### Answer

Best-effort under `apiVersion: yerk/v1`; breaks bump version or release notes.
No iron-clad pre-1.0 claim; avoid casual renames.

### Notes

Operator accepted 2026-09-30.

---

## Q27 — XDG config vs state split

Status: locked

### Prompt

Confirm baseline XDG roles.

### Answer

| Tree | Role |
|------|------|
| `$XDG_CONFIG_HOME/yerk` | catalog + operator tool prefs |
| `$XDG_STATE_HOME/yerk` | tool-written project bindings/state |
| `$XDG_CACHE_HOME/yerk` | disposable caches (later) |
| `$XDG_DATA_HOME/yerk` | optional durable data (later) |

Dir-local `.local/yerk/` remains non-XDG tree config.

### Notes

Operator accepted 2026-09-30 → Phase 2 ADR.

---

## Q28 — Parallel change probes

Status: locked

### Prompt

Keep parallel probes out of this plan’s execute queue?

### Answer

**Yes — deferred.** Serial only until operator reprioritizes (owner backlog).

### Notes

Operator accepted 2026-09-30.

---

## Q29 — Path resolution without domain→directory mapping

Status: locked

### Prompt

How is **project workspace / replica filesystem path** chosen once domain is
only a logical namespace (Q4), not ADR 008’s `domains[domain] + relative path`?

### Answer

Synthesized from operator **Q4** + **Q8/Q9/Q32/Q33** (operator noted the path
model was restated there, not as a separate Q29 essay).

**Core rule:** placement paths are **host-local configuration**, not derived
from domain identity.

1. **Domain** (`personal`, …) appears only in **ids/URIs** and catalog identity
   (Q16/Q30). It does **not** select a filesystem root and is **not** joined to
   catalog `path` for resolve (ADR 008 domain-root join is **superseded**).

2. **`[domains]` map in `config.toml`** (if kept at all) is **not** the path
   engine for workspace style. Prefer removing path-resolution dependency on it
   in the successor ADR; dogfood migrates off relative-to-domain-root paths.

3. **Where paths come from (host / dir-local / project row):**
   - Effective **`workspace_style`** (string or **inline table**, Q32).
   - Style parameters supply directories as **absolute or `~/…` host paths**,
     e.g. operator mock:

     ```toml
     workspace_style = {
       style = "name-tags",
       main_dir = "~/.bimker",
       replica_dir = "~/tree/personal/devel/bimker",
     }
     ```

   - For plain styles without split dirs, the same idea: base directories are
     configured on the host (global default, dir-local, or per-project inline),
     not via `domain → root`.

4. **`name-tags` path math** (Q8/Q33):
   - Workspace (ensure/status/`path` project) = **container** dir (e.g.
     `replica_dir` or configured parent such as `…/projects`).
   - Main replica = container + project **name** as bare directory
     (`…/projects/wumpus`), unless `main_dir` overrides the main location
     (bimker-style).
   - Other replicas = beside main with `__` tag (`wumpus__feat1`), under the
     replica container (`replica_dir` when set).

5. **`workspace-dir` / `project-dir`:** keep existing *relative* layout math
   once a **host-resolved workspace base** exists; that base is **not**
   `domains[domain]+path`. Successor ADR should define the base as: explicit
   absolute/`~/` workspace path on the project or from style/dir-local host
   config (details in placement ADR—still “host path”, never domain join).

6. **Catalog `path` field:** stop treating it as “relative to domain root.”
   Prefer host-absolute or `~/` workspace locations (and/or fold location into
   style params). Relative-only portable catalog rows need a **non-domain**
   host mechanism later if desired—not domain roots.

7. **Omit path until init?** Not required for v1 of this lock; if omitted,
   ensure/materialize must receive enough style/base config to compute paths or
   error clearly. (No silent domain default.)

8. **State vs catalog path:** bound **style** (and style params) follow Q1
   (state wins + warn). Desired workspace location should be consistent with
   that binding; do not reintroduce domain-root relocation. Exact “catalog path
   string vs state snapshot” can be refined in the ADR without reviving ADR 008.

### Notes

Operator 2026-09-30: “I restated how to handle paths … in the last edits”
→ Q4 (no domain→directory) + Q9 `config.toml` inline dirs. Agent locked Q29
from that synthesis. **Promote to ADR** replacing/soft-deprecating ADR 008 path
join.

**Operator correction 2026-10-01:** most projects should **not** need a host
path row. Restore optional **`[domains]`** as a host convenience default:
workspace = `<domains[domain]>/<name>` when no `[[projects]]` path is set.
Full absolute/`~/` path (or relative under the root) only for exceptions.
Catalog still has **no** `path`. ADR 014 documents this hybrid; bullets 2/5/6
above are softened accordingly (domain root is optional default base, not
identity and not catalog join).

---

## Q30 — Catalog project identity fields

Status: locked

### Prompt

Q9 mock used `id = "personal/wumpus"`. Today catalog has `name` + `domain`.

Pick identity shape for catalog rows:

| Opt | Shape |
|-----|--------|
| A | Keep `name` + required `domain`; derived id `domain/name` |
| B | Single required `id = "domain/name"`; split for display |
| C | Allow either; normalize to URI |

### Answer

**A** — keep `name` + required `domain`; derived bare id `domain/name` and URI
`yerk://domain/name`.

### Notes

Operator 2026-09-30. Q9 `id = …` mock was illustrative only.

---

## Q31 — Rename candidates for today’s `yerk clone`

Status: locked

### Prompt

Choose a name for “materialize first/main replica from **remote**”.

### Answer

Merged into **Q12**: product command **`yerk materialize`**; no `clone` alias.

### Notes

Operator 2026-09-30.

---

## Q32 — Style derivations / inline parameters syntax

Status: locked

### Prompt

Named derivations vs inline table for parameterized styles?

### Answer

- **Inline table only** (or plain string style name when no params).
- **No** named derivation registry (`[workspace.styles.bimker]`).
- Example shape (field names can be fixed in ADR):

  ```toml
  workspace_style = "name-tags"

  workspace_style = { style = "name-tags", main_dir = "~/.bimker", replica_dir = "…" }
  ```

- TOML type: **string or table**. Inner key for the base style: prefer **`style`**
  (not `name`) unless operator objects later.

### Notes

Operator 2026-09-30: named derivations too complex / no good use case.
String-only styles still valid for `workspace-dir` / `project-dir` / bare
`name-tags`.

---

## Q33 — Project workspace under `name-tags`

Status: locked

### Prompt

Under `name-tags`, what is project workspace vs main replica path?

### Answer

- **Workspace path** = directory that **contains** the bare main checkout and
  sibling tagged replicas (e.g. `…/projects` if main is `…/projects/wumpus`).
- **`yerk path` / `workspace ensure` / status project PATH** → that **container**.
- **`path <project> <main-or-default>`** → bare main dir (`…/projects/wumpus`).
- Main is **not** a child folder literally named `main` under this style.

### Notes

Operator accepted proposal 2026-09-30.

---

## Changelog

- 2026-09-30: Created plan folder; Q1–Q28 opened/proposed.
- 2026-09-30: Operator answers applied (round 1). Follow-ups Q29–Q33 added.
- 2026-09-30: Operator round 2 — locked Q12/Q31 (`materialize`), Q30 (A),
  Q32 (inline only), Q33 (workspace = parent).
- 2026-09-30: **Q29 locked** by synthesizing Q4 + Q9 path/config.toml model
  (host-local style dirs; no domain-root path join; ADR 008 superseded for
  resolve). Open queue clear for Phase 0.
