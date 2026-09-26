# Repo map

FQ name: `salotz.yerk`

## Control plane

- `AGENTS.md`, `README.md` — hubs
- `design/` — goals + ADRs
- `contributing/` — human/agent how-to for tooling
- `.bootstrap/` — host tool check-only
- `.config/_project-meta.toml`, `.prjx-root` — PRJX
- `mise.toml` — Go pin + tasks
- `.agents/` — this tree

## Product code

- `cmd/yerk` — CLI main
- `internal/cli` — cobra command tree
- `internal/config` — XDG TOML catalog
- `internal/workspace` — replica path styles (MVP)
- `internal/version` — ldflags-friendly identity

## Out of tree (operator notes)

Design brainstorm lives in the personal org silo idea note:

`admin/org/notes/todo/ideas/20260925T115055--software-project-management-tool__dev_software_todo.org`
