# 024. Packslip release manifests

## Status

Accepted (2026-10-02)

## Context

yerk ships a standalone Go binary (ADR 002) with Growth Versioning tags
(ADR 023). Operators want install via **mise** (and similar) with verified
artifacts, not only a raw GitHub Release download. salotz RFC 030 already
treats [packslip](https://packslip.dev/) as the signed release / install
metadata layer complementary to `.appinfo` usage context.

## Decision

- Publish a **packslip** bundle on every product release tag `vB.R.G`.
- Implementation: GitHub Actions job in
  [`.github/workflows/release.yml`](../../.github/workflows/release.yml) after
  release assets are uploaded, using [`jdx/packslip@v1`](https://github.com/jdx/packslip)
  (OIDC / keyless; no long-lived signing key in-repo).
- **Project name:** `github.com/salotz/yerk` (matches module / default packslip
  GitHub project form).
- **Primary install artifact:** `yerk-<B.R.G>-linux-x64.tar.gz` containing a
  single executable `yerk` at archive root (`CGO_ENABLED=0`). Packslip
  `bin: yerk` only describes this tarball (not the bare `.bin` or checksum
  file).
- **Permissions on the release job:** `contents: write`, `id-token: write`,
  `attestations: write` (attest + upload bundle in-job). A split
  read-only-signer job remains an optional hardening later; not required now.
- **Consumer install (documented):** `mise use … packslip:github.com/salotz/yerk`
  (mise 2026.9.2+). After the first release, publish `packslip pin` fingerprint
  in operator docs when available.
- **Workflow path stability:** keep packslip signing in
  `.github/workflows/release.yml` only. Moving the signing step to another
  workflow file changes the signer identity consumers remember.

## Non-goals (this ADR)

- Multi-OS/arch packslip artifacts (may extend the same workflow later).
- Completions / man / skills resources in the packslip (add when CLI supports
  them cleanly).
- Domain-hosted project name or supplementary release lists.
- Replacing ADR 023 tag/`B.R.G` rules.

## Consequences

- Release job is slightly privileged (write + OIDC + attestations).
- First green tagged release is required before mise packslip install works.
- install-distribution “channel” defaults to packslip + GH Release assets;
  ubi/aqua registry entries are optional extras, not the primary path.
- RFC 030 `.appinfo` remains usage/env context; packslip remains release
  identity—do not merge the two files.
