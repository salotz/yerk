# Repo map

FQ name: `salotz.yerk`

## Control plane

- `AGENTS.md`, `README.md` — hubs
- `.appinfo/meta.toml` — application info + env registry (RFC 030/031)
- `docs/` — operator/user docs (Diátaxis; ad hoc Markdown until publication, ADR 006)
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

- **Workspace-local agent context (RFC 23):** parent of this replica is the
  yerk **project workspace** (`…/yerk/` → checkout `main/`). Catalog uses
  relative `path` under domain root (ADR 008), not the replica path. Host
  dogfood guidance for updating `~/.config/yerk/` lives in `…/yerk/.agents/`
  (not in this remote tree). Optional RFC 26/PRJX host-local project files:
  `…/yerk/.local/`.
- Design brainstorm lives in the personal org silo idea note:

`admin/org/notes/todo/ideas/20260925T115055--software-project-management-tool__dev_software_todo.org`
