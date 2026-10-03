# 023. Growth Versioning (salotz RFC 002)

## Status

Accepted (2026-10-02)

## Context

yerk needs a **product version** for binaries (`yerk version`), release tags,
and install channels. Today:

- `internal/version.Version` defaults to `0.0.0-dev` and can be set via
  `-ldflags` (see [contributing/development.md](../../contributing/development.md)).
- No tag scheme, `VERSION` file, or changelog process is locked.
- [install-distribution](../../.agents/plans/salotz/install-distribution/) and
  CI release work need a scheme before stamping artifacts.

Operator preference is **salotz RFC 002** (*Semantic Changelog* / Growth
Versioning): familiar three-part numbers with **different meanings** than
classic SemVer MAJOR.MINOR.PATCH.

This ADR chooses the **scheme and product rules**. Mechanics of stamping
builds and publishing install assets remain with install-distribution (and
CI), constrained by this decision.

## Decision

### Scheme

yerk **product versions** use RFC 002 **Growth Versioning**:

```text
B.R.G
```

| Part | Name | Increments when |
| --- | --- | --- |
| **B** | Breakage | Any **breakage** change for end users (stricter, stingier, replaced, rename, removal, …) |
| **R** | Regression | Noteworthy **regression** (hamstring, deprecation, pollution, noisier, …). Optional in practice; still part of the tuple. |
| **G** | Growth | Noteworthy **growth** (feature, relaxed, repair, performance, clarity, robustness, …) |

Rules (from RFC 002, applied here):

- Format is still three non-negative integers separated by dots (SemVer-*shaped*).
- Incrementing **B** resets **G** to `0` (new effective product surface).
- Incrementing **R** does **not** reset **G**.
- While **B is 0** (pre-1.0 product), interface breakages are recorded as
  **regressions** (bump **R**, not **B**), matching the usual “no stable
  contract until 1.0” pattern. First intentional stable public contract ships
  as **B ≥ 1**.
- Users should be able to take **G**-only bumps without fear of user-facing
  breakage or regression intent.

### Tag and string form

- Git tags for releases: **`vB.R.G`** (leading `v`, e.g. `v0.1.3`).
- The string stamped into `internal/version.Version` for a release build is
  **`B.R.G` without the `v`** (e.g. `0.1.3`), unless a later ADR changes the
  CLI contract.
- Pre-release / dirty / untagged developer builds keep a clear marker such as
  the current default `0.0.0-dev` or a derived form like
  `B.R.G-dev+<git>` — exact dirty formula is an implementation detail of
  install-distribution, not a second product scheme.

### What this is *not*

| Surface | Scheme | Notes |
| --- | --- | --- |
| Product / binary version | **B.R.G** (this ADR) | `yerk version`, release tags, install metadata |
| API resources `apiVersion` | `yerk/v1` (ADR 011) | Wire format generation; bump only on resource schema breaks |
| `.appinfo/meta.toml` `version` | RFC 030 appinfo doc version | Not the CLI product version |

Do not conflate these three.

### Changelog and commits (intent, not tooling yet)

- **End-user changelog** (when introduced) follows RFC 002 section names:
  Improvements / Breakages / Regressions (Keep a Changelog structure with
  renamed sections), audience = users.
- **Commit trailers** (`Change`, `Change-Full`, `Audience`) are the preferred
  structured metadata when authors bother; they are **not** required on every
  commit for day-to-day work. No Conventional Commits-style subject prefixes
  for change kind (RFC 002).
- Automating changelog generation from trailers is optional later work; this
  ADR does not mandate a generator.

### Version source of truth (product identity)

For **released** product identity:

1. **Git tag `vB.R.G`** is the canonical release name.
2. The binary stamp **must match** that release’s `B.R.G`.

For **day-to-day development**, do **not** require bumping a version on every
merge. Unreleased trees may stay on the dev default until a release is cut.

**Operator bump path:** `mise run version-show` / `mise run version-bump --
<growth|regression|breakage> [--tag] [--push]` (script
[`.tasks/version-bump`](../../.tasks/version-bump)). Dry-run is default; `--tag`
writes the annotated tag on `HEAD`. SemVer-shaped aliases `patch`/`minor`/`major`
map by tuple position only (G/R/B), not classic SemVer semantics.

Optional in-tree `VERSION` (or equivalent) may mirror the **next** or **last**
release for tools that cannot read git; if both tag and file exist for a
release cut, they **must agree**. Prefer **not** hand-editing version in many
places—derive stamp from tag (or one file) at build time.

**CI:** Tag push does not imply automation until a workflow is added. Intended
split (plans): check pipeline on PR/default-branch push; release/publish on
`v*` tags (install-distribution), stamping `Version` from the tag.

## Consequences

- install-distribution Phase 1 (ldflags / tags) and any release CI **must** use
  B.R.G + `vB.R.G` tags; plan Q3 “version scheme” is answered by this ADR.
- Docs and agents should say **Growth Versioning / B.R.G (RFC 002)**, not
  “SemVer”, when explaining product bumps—even though the wire shape matches
  `X.Y.Z`.
- Pre-1.0 (`0.R.G`): breaking CLI/catalog/config changes bump **R** (as
  regression), not **B**, until the project deliberately declares 1.0.0.
- Classic SemVer tooling (libraries that assume MAJOR = breakage only in the
  SemVer sense) remains usable for **ordering** and most constraint parsers;
  human/release notes must still use RFC 002 meanings for **R** vs **G**.
- No requirement to adopt commit-trailer discipline before first release; low
  friction wins until automation exists.

## References

- salotz RFC 002 — Semantic Changelog / Growth Versioning  
  (`rfcs/salotz.002_semantic-changelog` in the salotz/rfcs tree)
- [ADR 011](./011-api-resources.md) — `apiVersion` on resources
- [ADR 002](./002-go-standalone-binary.md) — standalone binary product
- `internal/version` — link-time identity hooks
