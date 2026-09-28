# 006. Ad hoc plain Markdown docs (Diátaxis), publication later

## Status

Accepted (2026-09-28)

## Context

`yerk` needs operator- and user-facing documentation beyond the root README,
design notes, and contributor tooling guides.

Those audiences and jobs differ:

| Tree | Job |
| --- | --- |
| `docs/` | Use and operate the shipped CLI |
| `design/` | Goals, domain language, ADRs |
| `contributing/` | Bootstrap, build, test, project tooling |

We do not yet want to pick and maintain a documentation **publication** stack
(static site generator, theme, hosting, search, versioned multi-format
output). Doing that early would slow writing real content and couple IA to a
tool choice we have not earned.

Content still needs a durable **information architecture** so pages do not
collapse into one long README. Project guidelines default software docs to
[Diátaxis](https://diataxis.fr/) (tutorials, how-to guides, reference,
explanation).

CLI help remains a separate surface (ADR 005): flags and env vars are
authoritative in the binary (`--help`, `yerk help envvars`, `yerk envvars`).
Repo docs should not become a second hand-maintained dump of that surface
without a clear reference strategy.

## Decision

1. **Location.** User/operator docs live under `docs/` at the repo root.
2. **Format (near term).** **Ad hoc plain Markdown** only—edit files in git;
   no required doc generator, theme, or publish pipeline for this stage.
3. **Structure.** Organize `docs/` by **Diátaxis** types:
   - `docs/tutorials/` — learning-oriented lessons
   - `docs/how-to/` — task-oriented guides
   - `docs/reference/` — information-oriented lookup
   - `docs/explanation/` — understanding-oriented background
4. **Stubs first.** Type hubs (and later pages) may start as stubs; fill
   content as features stabilize. Prefer correct type over premature polish.
5. **Publication later.** A better system for **actual publication**
   (site build, navigation chrome, hosting, possibly alternate formats) is
   **explicitly deferred**. Choosing MkDocs, mdBook, Sphinx, Hugo, Docusaurus,
   or similar is out of scope for this ADR. When adopted, it should consume
   or map this Diátaxis tree rather than invent a parallel IA.
6. **Non-goals of `docs/`.** Design history stays in `design/`; contributor
   bootstrap stays in `contributing/`. Cross-link; do not duplicate.

## Consequences

- Authors write Markdown under the four type directories; the `docs/README.md`
  hub points at Diátaxis and this ADR.
- Reviewers can reject mixed-type pages or “everything in README” dumps in
  favor of the matching stub/type.
- CI need not build a doc site until a future ADR selects publication tooling.
- Moving to a generator later should be mostly path and nav wiring if content
  already respects Diátaxis boundaries.
- Root README stays a short hub (quick start + pointers), not the full manual.

## References

- [Diátaxis](https://diataxis.fr/)
- Agent-guidelines application: software documentation (Diátaxis default)
- [005](./005-cli-help-and-envvars.md) — CLI help and env var docs
