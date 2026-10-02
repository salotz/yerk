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

_(none yet)_

---

## Notes

Operator intent: install via tools like **mise**, not only `go build` from a
clone. Keep ADR 007 (no host-private release fixtures in-repo).

`internal/version` already exists for identity strings; this plan wires
production stamping and distribution.
