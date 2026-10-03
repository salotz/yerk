# Decisions inbox — install-distribution

Plan-local Q&A only. Promote durable choices to `design/decisions/` ADRs.
Do **not** cite `Q*` in product code or operator docs.

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q1 | open | Install channels (mise backend: ubi / aqua / go / asdf plugin / other) |
| Q2 | open | Artifact shape (single OS first vs multi-OS + checksums) |
| Q3 | open | Version source (git tag, VERSION file, both) and ldflags package path |
| Q4 | open | Release automation (manual tag vs CI release job; soft dep ci-pipelines) |
| Q5 | open | Module `go install` support as first-class vs binary-only |

## Locked

| ID | Decision |
|----|----------|
| Q3 (scheme) | **Growth Versioning** B.R.G per product [ADR 023](../../../../design/decisions/023-growth-versioning.md) (salotz RFC 002). Tags `vB.R.G`; ldflags stamp `B.R.G` into `github.com/salotz/yerk/internal/version.Version`. |
| Q3 (source) | **Git tag is canonical** for releases. Optional single in-tree `VERSION` only if a tool cannot read git; must agree with tag at cut. No multi-file hand bumps. Dev/default remains `0.0.0-dev` (or dirty derived form) until stamped. |

Still open under Q3 implementation detail: whether release builds always require an exact tag vs allow `git describe`-style dirty strings for non-release artifacts.

---

## Notes

Operator intent: install via tools like **mise**, not only `go build` from a
clone. Keep ADR 007 (no host-private release fixtures in-repo).

`internal/version` already exists for identity strings; this plan wires
production stamping and distribution. Product version **meanings** are ADR 023
(not classic SemVer MAJOR.MINOR.PATCH).
