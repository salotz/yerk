# Decisions inbox — install-distribution

Plan-local Q&A only. Promote durable choices to `design/decisions/` ADRs.
Do **not** cite `Q*` in product code or operator docs.

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q2 | open | Multi-OS/arch beyond linux-x64 packslip artifact |
| Q5 | open | Module `go install` support as first-class vs binary-only |

## Locked

| ID | Decision |
|----|----------|
| Q1 | **Primary channel: packslip** via GitHub Release + `jdx/packslip@v1` ([ADR 024](../../../../design/decisions/024-packslip-releases.md)). Consumer: `mise use packslip:github.com/salotz/yerk`. Optional ubi/aqua/registry later—not required for MVP. |
| Q2 (mvp) | **linux-x64 only** for now (`yerk-<ver>-linux-x64.tar.gz`); expand matrix later under remaining Q2. |
| Q3 (scheme) | **Growth Versioning** B.R.G per [ADR 023](../../../../design/decisions/023-growth-versioning.md). Tags `vB.R.G`; ldflags stamp `B.R.G`. |
| Q3 (source) | **Git tag canonical**. Helpers: `.tasks/go-ldflags`, `.tasks/build-yerk`. Untagged → `0.0.0-dev`. Release CI sets `YERK_VERSION`. |
| Q4 | **GitHub Actions** [`.github/workflows/release.yml`](../../../../.github/workflows/release.yml): test + vet + archive + GH Release + **packslip sign/upload**. Cut tags with `mise run version-bump -- <part> --tag` (+ push). |

---

## Notes

ADR 007: no host-private release fixtures in-repo. Signer is workflow OIDC,
not a committed key.

After first remote release: run `packslip pin` on the bundle and paste the
`ps1_…` fingerprint into README / install how-to.
