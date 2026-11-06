# Handoff: domspec (prototype)

**Created:** 2026-11-03T15:43-05:00  
**From:** goose session on bimker replica `emacs-yerk-jump`  
**Status:** open — name not locked; schemas empty; still nested in bimker

Treat **this directory** (`design/external/domspec/`) as its own project.
Move it out of bimker once the prototype is useful. Do not treat bimker
ADRs or yerk Go as this repo’s source of truth.

## What this is

**Codename: `domspec`.** A small CUE-based framework for writing **checkable
domain models** (types now; functions, examples, traces, and test tooling
later). Product repos import it instead of duplicating envelope conventions.

It is **not** a new modelling language. CUE is the notation. This project
only adds shared schemas and, later, tools that sit on those schemas.

Informal spec for now: humans and agents read the CUE; implementations stay
hand-written. Codegen from CUE (including Go) is an eventual goal, not this
prototype.

**Name is not locked.** `domspec` is a working name so the module path has
somewhere to live. Rename before publish if a better name wins.

## Where it lives

| Path | Role |
|------|------|
| This tree | Prototype checkout (nested under bimker `design/external/`) |
| `cue.mod/module.cue` | CUE module identity |
| `README.md` | One-line stub |
| Parent bimker replica | `~/tree/personal/devel/bimker/bimker__emacs-yerk-jump` |
| First consumer | that replica’s `design/models/` (yerk types, **not** this repo) |

On extract: this folder becomes its own git project. Import paths below are
already the intended published names so they need not change.

## Module paths (locked for the prototype)

CUE module paths are slash paths whose first component is a domain we
control — not `github.com/…`, not a URL with a scheme.

| Module | Meaning |
|--------|---------|
| `salotz.info/domspec/core` | **This** project: shared modelling schemas |
| `salotz.info/domspec/…` | Room for sibling toolkit modules later (`cli`, `testkit`, …) |
| `salotz.info/yerk/domspec` | Yerk’s domain-spec installable (consumer, other repo) |
| `salotz.info/yerk/…` | Room for other yerk CUE modules |

Do **not** also publish parent modules `salotz.info/domspec` or
`salotz.info/yerk` — nested module paths must not overlap. Every published
module is a leaf.

`design/models/` in a product repo is the RFC 22 **folder**. It does not
have to match the module string (`…/yerk/domspec`).

One CUE **module** is one installable (Python: one PyPI distribution).
Packages are directories inside it (`domain/v1alpha1`, later `examples/`,
`functions/`). Do not put a `cue.mod` in each of those folders.

## What belongs here vs in a product

| Here (`salotz.info/domspec/core`) | Product (`salotz.info/<app>/domspec`) |
|----------------------------------|----------------------------------------|
| Envelope / TypeMeta, later call-records (`in`/`out`), example/trace conventions | Nouns of that product (`#Project`, `#Presence`, …) |
| Versioning and file-layout conventions for models | API generation folders (`domain/v1alpha1`, …) |
| Later: vet helpers, test runners, frontends | Examples and function specs **of that product** |

Yerk kinds stay in yerk’s module. Do not copy `#Project` into this tree.

The first consumer (still in the parent replica, not extracted):

```text
design/models/                         module: salotz.info/yerk/domspec
  cue.mod/
  domain/v1alpha1/                     package still named `models` (should become `v1alpha1`)
    core.cue                           Presence, Kind, APIVersion, ResourceMeta
    catalog.cue, info.cue, status.cue, …
    _package.cue                       #Resource union — deferred, candidate to delete
```

That consumer does **not** import this module yet. This tree has no schema
files, only `cue.mod`. Wiring the import is future work.

## Locked decisions

- **Notation:** CUE. Not TypeScript, not a custom sexp (that thread pivoted
  here). Field names are JSON (`defaultReplica`), not Go.
- **Check:** `cue vet` on the package. Product repos pin `cue` via mise
  (parent replica: `0.15.4`, language `v0.15.0`).
- **No codegen** until asked.
- **API version ≠ product version.** Product uses RFC 002 `B.R.G`. Wire
  `apiVersion` is a schema generation only.
- **API version form:** Kubernetes-style `vN` / `vNalphaM` / `vNbetaM`.
  Compatible growth (new optional fields) stays in the same generation;
  “need newer software” is the **product** bump. Incompatible schema →
  `v1alpha1` → `v1alpha2` (or later `v1beta1` / `v1`).
- **Do not use `v0` on the wire.** Outsiders do not read it as alpha.
  Current yerk draft: **`yerk/v1alpha1`**.
- **Do not put full `B.R.G` in `apiVersion`.**
- **Content vs index files (product packages):** ordinary names for nouns
  (`catalog.cue`); `core.cue` for small shared types. A package-index file
  (`_package.cue` / `resources.cue`) was proposed then deferred.

## Deferred

- Real schemas in this module (`#TypeMeta` / envelope). Yerk still defines
  `#ResourceMeta` locally.
- Call-records, examples, traces, layout packages.
- Software tools (vet orchestration, function tests, frontends).
- `#Resource` tagged union (`_package.cue` in the yerk consumer): a sum of
  every kind for a generic decoder. Nothing consumes it yet. Prefer
  **delete** over maintaining a parallel catalog; revisit when a generic
  read path exists.
- Publishing to a CUE registry; `cue.mod` deps / local replace (CUE 0.15
  has no Go-style `replace`).
- Final project name.
- JSON Schema / OpenAPI generation.

## Suggested cadence (this project)

1. Keep the tree extractable: no bimker-only paths inside schemas.
2. Add the smallest shared CUE (`#TypeMeta`: `apiVersion` + `kind` as
   strings) when the yerk consumer is ready to import it.
3. Point yerk’s `domain/v1alpha1` at this module; drop the local duplicate.
4. Extract to its own git repo; keep module path `salotz.info/domspec/core`.
5. Grow tools and extra modules under `salotz.info/domspec/…` only when a
   concrete need shows up.

## Do not

- Rename or publish without an explicit name lock.
- Put product nouns (yerk `Project`, …) in this module.
- Publish overlapping parent module paths.
- Implement yerk Go, Emacs, or codegen from a domspec thread unless asked.
- Land this on `~/.bimker` / yerk `main` from the parent replica until asked.
- Commit unless asked.

## Pointers

| What | Where |
|------|--------|
| This prototype | `~/tree/personal/devel/yerk/main/design/external/domspec/` |
| First consumer types | `~/tree/personal/devel/yerk/main/design/models/domain/v1alpha1/` |
| Consumer module | `salotz.info/yerk/domspec` |
| Notation origin | `~/.local/xagents/handoffs/2026-11-02--yerk.domain-type-notation.md` (superseded: CUE won) |
| CUE decision notes | `yerk://personal/salotz-org/master/notes/threads/20261103T123800--domain-model-types-contracts-and-cue__thread_devel.org` |
| Yerk product API (today’s wire, not the target redraw) | `~/tree/personal/devel/yerk/main/internal/api/types.go` |
| RFC 002 (product versioning only) | agent-guidelines summaries / yerk ADR 023 |
