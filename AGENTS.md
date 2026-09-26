# Agent guidelines

This project follows the guidelines at https://github.com/salotz/agent-guidelines
(local checkout often `~/tree/personal/devel/agent-guidelines`).

- Read `content/shared/generic-agent-guidelines.md` (or equivalent) for agent-assisted work.
- Prefer task-specific tools over shell.
- Follow salotz RFC 22 (project layout), RFC 23/24 (host context), RFC 28 (PRJX).
- Cache remote resources locally when possible.
- Use compacted inlining for referenced standards.
- Check `content/shared/summaries/` before fetching full external standards.

See agent-guidelines `content/shared/glossary.md` for term definitions.

## This repository

FQ name `salotz.yerk`.
Go CLI shipped as a standalone binary named `yerk`.

Host multi-project manager: register many software projects on one machine,
check out replicas into workspace layouts, status across repos, and (planned)
stage host-local configuration. Speaks PRJX vocabulary.

### Pointers

- [README.md](./README.md): human hub
- [design/goals.md](./design/goals.md): product goals from the idea note
- [design/decisions/](./design/decisions/): ADRs
- [contributing/development.md](./contributing/development.md): bootstrap → build → test
- [`.agents/`](./.agents/): plans and context
- `.prjx-root`, [`.config/`](./.config/): PRJX root + portable metadata
- `mise.toml`, `.editorconfig`, `.bootstrap/`: tooling

### Tooling entrypoints

```sh
sh .bootstrap/host-tool-check
mise install && mise run preload
mise run build
mise run test
mise run check
```

Prefer `mise run` / `mise exec` so pins and `.local` Go caches apply without
shell activation.

### Implementation notes for agents

- Language: Go. Module path: `github.com/salotz/yerk`.
- Binary entrypoint: `cmd/yerk`.
- Keep packages small under `internal/…`.
- Config is TOML under `~/.config/yerk` with `YERK__…` overrides.
- Do not invent a second product root; code lives at repo root Go layout,
  design/docs beside it.
- MVP stubs already exist for `register`, `clone`, `pull`, `push`.
  Prefer implementing those before new surface area.
- Design source note (operator org):
  `notes/todo/ideas/20260925T115055--software-project-management-tool__dev_software_todo.org`
