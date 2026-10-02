# 021. MIT license

## Status

Accepted (2026-10-02)

## Context

The repository shipped without a root license file. Open-source distribution
(clones, forks, binary packaging, and agent/tool reuse of the tree) needs a
clear grant of rights. Operator preference is a short, permissive license
common for Go CLI tools rather than copyleft or a multi-clause corporate form.

## Decision

- License the project under the **MIT License**.
- Canonical text lives at the repo root as [`LICENSE`](../../LICENSE).
- Copyright line: **Copyright (c) 2026 salotz** (update year or holder only
  when intentionally changing ownership notice).
- Do not dual-license or add a second product license without a new ADR.

## Consequences

- Redistributors must retain the copyright and permission notice.
- No patent grant beyond MIT’s usual text; no CLA required for this choice.
- README / packaging metadata may point at `LICENSE`; the file is authoritative.
- Third-party deps keep their own licenses; this ADR covers **yerk** sources
  only.
