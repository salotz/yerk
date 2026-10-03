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

