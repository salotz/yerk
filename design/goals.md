# Goals

Product intent distilled from the operator idea note
(`software-project-management-tool`, 2026-09-25).

## Main goal

Manage many software projects on one host machine.

Developers (and coding agents) multitask across projects. Tracking remote and
local state by hand does not scale:

- Were latest changes pushed everywhere?
- Are there dirty or untracked trees?
- Can I pull everything in one domain?

Related prior art: `gita` (register many git repos, status). `yerk` is a
personal tool with a broader host-ops surface.

## Name

- Binary / package / repo: **yerk**
- Config: `~/.config/yerk`
- Env: `YERK__…` for tool knobs; discover `.prjx-root` / `PRJX__…` for PRJX
- One-liner: host multi-project manager (replicas, locals, status); speaks PRJX

## Feature directions (beyond MVP)

- Clone into workspaces optimized for agentic workflows
- Apply local (untracked) configuration into new checkouts
  (dotfiles, domain/project `.local`, tool-managed locals)
- Index + optional FS watch for efficient change collection
- Bulk checkout into host structure (e.g. bimtree domains)
- Domain language: project, workspace, replica, worktree
- Monorepo-aware checkout presets; PRJX monorepo integration
- Query by project/replica name rather than raw paths
- Host discoverability for agents and editors (e.g. Emacs)
- Standard ignore/clean tasks (`node_modules`, prune worktrees)
- Track host resources tied to projects (devpod, containers, local k8s)
- Manage port ranges and other global resource spaces per project
- Integrate file sync (e.g. mutagen) for split edit/run hosts

## Motivating use cases

### Agent session context

Agent shell switchers often show a short session/project label (`main`) that
collides across repos. `yerk` should supply columns for domain, project name,
and replica name so session UIs can render unambiguous rows. Session/agent
metadata stays with the session manager; project columns come from `yerk`.

### Staging local configuration

Example: DVC `config.local` overrides that must not be committed, staged into
each replica on checkout from:

- project workspace `.local` trees (RFC 26 style), and/or
- `~/.config/yerk/projects/<project>/locals/…`

## MVP (first milestone)

1. Config file: register projects by name, remote URI, destination, tags
2. CLI: clone, push, pull, status
3. Ship two workspace styles only:
   - workspace dir: `projects/<project>/<replica>`
   - project dir + replica suffix: `projects/<project>__<replica>`
