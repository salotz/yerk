# Decisions inbox — install-distribution

Plan-local Q&A only. Promote durable choices to `design/decisions/` ADRs.
Do **not** cite `Q*` in product code or operator docs.

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q1 | open | Install channels (mise backend: ubi / aqua / go / asdf plugin / other) |
| Q2 | open | Artifact shape beyond current GH Release linux/amd64 (multi-OS matrix) |
| Q5 | open | Module `go install` support as first-class vs binary-only |

## Locked

| ID | Decision |
|----|----------|
| Q3 (scheme) | **Growth Versioning** B.R.G per product [ADR 023](../../../../design/decisions/023-growth-versioning.md) (salotz RFC 002). Tags `vB.R.G`; ldflags stamp `B.R.G` into `github.com/salotz/yerk/internal/version.Version`. |
| Q3 (source) | **Git tag is canonical** for releases. Helpers: `.tasks/go-ldflags`, `.tasks/build-yerk` (`mise run build`). Untagged → `0.0.0-dev`. Exact tag on HEAD or `YERK_VERSION` for release CI. No multi-file hand bumps. |
| Q4 | **GitHub Actions** [`.github/workflows/release.yml`](../../../../.github/workflows/release.yml) on push tags `vB.R.G`: test + vet + linux/amd64 binary + `.tar.gz` + `SHA256SUMS` + GitHub Release. Cut tags with `mise run version-bump -- <part> --tag` (+ push). |

---

## Notes

Operator intent: install via tools like **mise**, not only `go build` from a
clone. Keep ADR 007 (no host-private release fixtures in-repo).

`internal/version` already exists for identity strings; this plan wires
production stamping and distribution. Product version **meanings** are ADR 023
(not classic SemVer MAJOR.MINOR.PATCH).

GH Release linux/amd64 is the first artifact channel; mise/ubi registration is
still Q1.
