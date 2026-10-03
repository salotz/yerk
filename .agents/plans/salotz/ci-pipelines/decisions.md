# Decisions inbox — ci-pipelines

Plan-local Q&A only. Promote durable choices to `design/decisions/` if needed.
Do **not** cite `Q*` in product code or operator docs.

## Open queue

_(none — Q1–Q5 locked)_

## Locked

| ID | Decision |
|----|----------|
| Q1 | **GitHub Actions** (`.github/workflows/`) |
| Q2 | **PR + push** to `main`/`master` for check CI; **not** tag-only |
| Q3 | Jobs: `go test ./...`, `go vet ./...`, stamped `build` via `.tasks/build-yerk`. **No** fmt gate in CI for now |
| Q4 | Go **1.27.1** in workflows (must match `mise.toml` `[tools].go`); document dual-maintain when bumping |
| Q5 | **linux/amd64 only** (`ubuntu-latest`) for check CI |

## Related (release, not this plan’s check job)

Release publish on `vB.R.G` tags lives in `.github/workflows/release.yml` (install-distribution + ADR 023). Check workflow: `ci.yml`.
