# Domain models (CUE)

Informal, language-neutral **product types** for yerk. Go in `internal/api`
is still hand-written; nothing here generates code.

These files translate today’s `internal/api` kinds (JSON field names,
required vs optional from `json` tags). They are **not** a redesign of
Project / ProjectInfo / ProjectStatus. File schema (`internal/config` TOML)
is out of scope.

Functions and traces come later.

## Layout

- `cue.mod/`: CUE module `salotz.info/yerk/domspec`
- `cue.mod/module.cue`: pins `salotz.info/{git,std,unix,url}`, RFC 027/033, XDG basedir, PRJ Spec (RFC 027 transitive)
- `cue.mod/local-module.cue`: gitignored replaces into RFCs `external/` and `rfcs/`
- `domain/v1alpha1/`: wire generation `yerk/v1alpha1` (unstable)
- `domain/v1alpha1/core.cue`: shared scalars (identity, paths, `#WorkspaceStyle`, `#Presence`, …)
- `domain/v1alpha1/catalog.cue`: `Project`, `Catalog`
- `domain/v1alpha1/placement.cue`: placement snapshot, style spec
- `domain/v1alpha1/info.cue`: `ProjectInfo`, `ReplicaInfo`
- `domain/v1alpha1/status.cue`: replica/project status
- `domain/v1alpha1/config.cue`: tool config dump, `config resolve`
- `domain/v1alpha1/context.cue`: env dump, tool/dir agent context
- `domain/v1alpha1/layout.cue`: on-disk layouts (RFC 033 dirshape)
- `domain/v1alpha1/resources.cue`: `#Resource` tagged union (deferred; candidate to delete)

Package clause is still `models` (should become `v1alpha1`). Field names
match JSON (`defaultReplica`, not `DefaultReplica`).

## External and RFC types

Reuse rather than re-encode. Product grammar stays here.

| Product type | Source |
|--------------|--------|
| `#AbsUnixPath`, `#HostPath` | `salotz.info/unix/path` |
| `#GitRemote` | `salotz.info/git/url` (`#Remote`) |
| `#BranchName` | `salotz.info/git/refs` plus `"-"` as a product sentinel |
| `#EnvVarName` | `salotz.info/unix/envvar` (`#Name`; mixed-case live dumps) |
| `#ToolEnvVar` | RFC 027 `#EnvName` constrained to `YERK__*` |
| `#PlatformEnvVar` | XDG `#Var` ∪ POSIX `#Name`, closed to names yerk reads |
| host/dir-local/workspace-dir layouts | RFC 033 `#LayoutSpec` |

`#ProjectDomain` is an ADR 012 segment, not `net.FQDN` and not RFC 025
`#DomainName`. `#WorkspaceStyle` is `workspace-dir` | `project-dir` |
`name-tags`. Product `yerk://` URIs stay here; generic URL strings are
`salotz.info/url/core` (schemes in `…/common`). `#ProductVersion` is a
string, not RFC 002 `#Version` (untagged builds stamp `0.0.0-dev`).

Not imported (wrong grammar or out of scope):

- RFC 028 `#ProjectID` / PRJ Spec IDs — not `domain/name`
- RFC 030/031 appinfo env tables — `.appinfo/meta.toml`, not the wire dump
- RFC 032 type/policy names — Go registry only today (`internal/envvars`)
- RFC 023 agent paths — not ADR 017 `ToolContext` / `DirContext`

Those modules live in `yerk://personal/rfcs` (`external/` and `rfcs/`).
`cue vet` resolves them through `cue.mod/local-module.cue` (gitignored)
with relative `replaceWith` paths; this workspace must have the RFCs
checkout next to yerk (`devel/rfcs/master` beside `devel/yerk`). A
registry would drop the need for that file.

Shared modelling conventions (envelope, later call-records) belong in
[`../external/domspec/`](../external/domspec/) (`salotz.info/domspec/core`),
not here. This module does not import it yet.

## Check

```sh
mise install
mise run models:vet
mise run models:fmt
```
