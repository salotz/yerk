# yerk

Host multi-project manager for humans, agents, and harnesses.

`yerk` keeps a **catalog** of projects, materializes **replicas** into a
**workspace** layout, resolves paths, and reports **presence** and git
**change** status. A replica is a checkout of a project (clone or worktree).

Written in Go as a single portable binary. Requires `git` on `PATH`.

## Declarations

Project declarations per
[RFC 014](https://github.com/salotz/rfcs/tree/master/rfcs/salotz.014_project-declarations.md).

| topic              | stance           |
|--------------------|------------------|
| lifecycle phase    | incubating       |
| maintenance intent | best-effort      |
| authorship         | ai-authored      |
| contributions      | limited          |
| support            | none             |
| license            | [MIT](./LICENSE) |

## Install

Download a release binary and put it on your `PATH`.
<!-- Binary release packaging is not shipped yet; use a local build until then. -->

## Configure

Two files under `~/.config/yerk` (or `$XDG_CONFIG_HOME/yerk`):

| File           | Role                                                   |
|----------------|--------------------------------------------------------|
| `config.toml`  | Host placement: workspace style, domain roots          |
| `catalog.toml` | Portable project registry (name, domain, remote, tags) |

Override paths with `YERK__CONFIG_DIR`, `YERK__CONFIG`, or `YERK__CATALOG`.

### Host config

```toml
# ~/.config/yerk/config.toml
[workspace]
style = "workspace-dir"

[domains]
personal = "~/personal"
```

With `workspace-dir`, each project’s workspace defaults to
`<domain-root>/<name>` (here `~/personal/yerk`). Replicas sit
under that directory (`…/yerk/main`, `…/yerk/feature-x`).

There are other options for workspaces that can be configured per-project as well.

### Catalog

```toml
# ~/.config/yerk/catalog.toml
tags = ["devel"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@github.com:salotz/yerk.git"
tags = ["devel"]
# default_replica = "main"   # optional; else remote HEAD / main
```

`domain` is required and should be thought of as akin to a domain name or namespace.

Paths stay out of the catalog;
the host config owns placement.

## Use

### 1. Materialize the main replica

```sh
yerk materialize yerk
# or: yerk materialize personal/yerk
# or: yerk materialize personal/yerk/main
# or: yerk materialize yerk://personal/yerk/main
```

Clones the default replica (usually `main`) into the workspace layout.
Already-present checkouts are left alone.

### 2. Status across projects

```sh
yerk status
yerk status yerk
```

Project rows summarize the default replica (presence + change).
Change probes (git status) run by default;
use `--presence-only` to skip git.

### 3. Create a session replica (worktree)

```sh
yerk replica create yerk feature-x
```

Default method is `git worktree` from the main replica (materialize main
first). Prints the absolute path of the new checkout.

```sh
cd "$(yerk path yerk feature-x)"
# work on the branch…
```

### 4. Change status on a replica

```sh
yerk status yerk feature-x
```

Change is a space-separated bag of flags from the git probe, for example:

| Situation               | Typical `change`                             |
|-------------------------|----------------------------------------------|
| In sync, clean tree     | `clean`                                      |
| Uncommitted edits       | `dirty`                                      |
| Untracked files         | `untracked` (often with `dirty`)             |
| Commits not on upstream | `clean ahead:N`                              |
| Dirty and unpushed      | `dirty ahead:N`                              |
| No upstream branch      | `clean no-upstream` (or `dirty no-upstream`) |
| Behind remote           | `clean behind:N` (can combine with `ahead:`) |

```sh
# dirty working tree
echo tweak >> README.md
yerk status yerk feature-x
# → … change: dirty …

# commit locally (still ahead of upstream)
git add README.md && git commit -m "tweak"
yerk status yerk feature-x
# → … change: clean ahead:1 …

# after push, clean and in sync
git push -u origin HEAD
yerk status yerk feature-x
# → … change: clean …
```

Project-level `yerk status yerk` rolls change across present replicas
(problem flags win over bare `clean`).

### 5. Bulk: all, domain, tags

In addition to providing a smooth surface for high-level operations on projects yerk makes it easy to manage many projects at once either by all projects, domains, or tags.
Tags are strings declared in your catalog and each project can have many tags.

Mutating commands take **one** selector: project names, `--all`, a single
`--tag`, or a single `--domain` (XOR). `status` defaults to the full catalog;
`--tag` / `--domain` filter it.

#### All projects

```sh
yerk status
# clones all default replicas
yerk materialize --all
```

#### One domain

`--domain` matches catalog `domain` (identity namespace), not host path roots:

```sh
yerk status --domain personal
yerk materialize --domain personal
yerk workspace ensure --domain personal
```

#### Tags

`--tag` is one declared catalog tag (must appear in the root `tags = […]`
vocabulary). For several tags, run once per tag (or name the projects):

```sh
yerk status --tag devel
yerk materialize --tag devel

yerk status --tag work
yerk materialize --tag infra
```

Combine tags in the catalog on each project (`tags = ["devel", "go"]`), then
pick the tag that defines the set you want to operate on.

### 6. Project introspection

`yerk` also provides utilities for introspection on projects.

This is useful as a tool for AI coding agents themselves and integration with
other tools such as AI coding session managers and ephemeral development
environments on branches.

For example your session manager may default to organizing sessions by the
folder name only, which can be ambiguous or verbose. Every resource in your
catalog and host has a unique URI with the prefix `yerk://` and path parts
`yerk://<domain>/<project>/<replica>`. On the CLI you can refer to projects
without the prefix.

By configuring project names, domains, and replica names you can always
unambiguously identify a project/replica from e.g. a filesystem path.

`yerk` has an internal API object model that can return structured output for
agents or tools (`--output json|yaml`, `apiVersion: yerk/v1`).

#### Path → project / replica

Map a filesystem path (or `.` for cwd) to a canonical URI. Walk-up matches the
longest workspace or replica root. Default output is the URI alone:

```sh
yerk lookup .
# → yerk://personal/yerk/main

yerk lookup /home/you/personal/yerk/main
yerk lookup /home/you/personal/yerk/feature-x/src

yerk project lookup /home/you/personal/yerk
# → yerk://personal/yerk

# expand when you need the full resource
yerk get "$(yerk lookup .)"
yerk lookup . --output json
```

#### Name → path

Resolve the on-disk workspace or replica path from an id (no side effects).
`path` and `resolve` are aliases:

```sh
yerk path yerk                  # project workspace
yerk path yerk main             # replica checkout
yerk path yerk/feature-x
yerk path personal/yerk/main
yerk path yerk://personal/yerk/main

cd "$(yerk path yerk feature-x)"
```

#### URI / id → JSON resource

Fetch one API resource by identifier. A bare project id is **project** info
(not the default replica); include the replica segment for checkout info:

```sh
yerk get yerk --output json
# → kind: ProjectInfo, uri: yerk://personal/yerk, …

yerk get yerk://personal/yerk --output json
yerk get personal/yerk/main --output json
yerk get yerk://personal/yerk/main --output json
# → kind: ReplicaInfo, uri: yerk://personal/yerk/main, path, presence, …

yerk get "$(yerk lookup .)" --output json
```

Same shapes via `yerk project get` and `yerk replica get`. For “where am I?”
plus placement and short status in one dump, see `yerk context dir`.
