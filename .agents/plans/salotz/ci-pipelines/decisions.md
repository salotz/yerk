# Decisions inbox — ci-pipelines

Plan-local Q&A only. Promote durable choices to `design/decisions/` if needed.
Do **not** cite `Q*` in product code or operator docs.

## Open queue

| ID | Status | Topic |
|----|--------|--------|
| Q1 | open | CI host (GitHub Actions vs Forgejo/other) |
| Q2 | open | Trigger policy (PR + main push vs main-only) |
| Q3 | open | Job set: test + vet + build; fmt check yes/no |
| Q4 | open | Go pin source (`mise.toml` version vs workflow-only) |
| Q5 | open | OS matrix now (linux only vs multi-OS CI) |

## Locked

_(none yet)_

---

## Defaults if operator is silent on go

- GitHub Actions (repo already GitHub-shaped module path)
- PR + push to default branch
- `go test ./...`, `go vet ./...`, `go build ./cmd/yerk`
- Go version aligned with `mise.toml` (`1.27.1` at plan spawn)
- Linux amd64 only for CI (multi-OS belongs to install-distribution if needed)
