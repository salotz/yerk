# Explanation: Status model

Status: **living** (aligned with domain design + current probe implementation)

## Intent

Explain why **presence** (disk vs expected path) and **change** (git probe)
are separate, how to read `yerk status` columns, what each change flag means,
and how that compares to tools like **gita**.

Authority split:

| Layer | Role |
| --- | --- |
| [design/domain-and-near-term.md](../../design/domain-and-near-term.md) | Product vocabulary and intended flag set |
| This page | Operator-facing explanation + probe behavior + known gaps |
| `internal/gitcmd` | Actual probe implementation (`Probe` / `Flags`) |

---

## Two jobs, two columns

| Concern | Question | Column | When empty / skipped |
| --- | --- | --- | --- |
| **Presence** | Is the expected path a usable checkout? | `PRESENCE` | Always filled for the scoped path |
| **Change** | What is dirty / untracked / vs remote *inside* that checkout? | `CHANGE` | `-` if not probed (`--presence-only`, or presence ≠ `present`) |

Do not treat `missing` as “clean”, or `dirty` as “missing”. Change is only
defined when presence is `present`.

---

## Presence

Applies to a **replica checkout path** (or, for project workspace dirs, a
simpler exists-as-directory check — not a git classifier).

| Value | Meaning |
| --- | --- |
| `missing` | Expected path does not exist |
| `present` | Path is a usable git checkout (`.git` dir, file, or symlink — covers worktrees) |
| `invalid` | Path exists but is not a usable checkout |

---

## Change: model (flags, not one enum)

Change is a **bag of orthogonal flags**, space-separated in the table. That
matches how operators think (“dirty *and* ahead”) better than a single state
machine.

### Axes

```text
LOCAL (worktree vs HEAD)          SYNC (HEAD vs remote reference)
────────────────────────          ────────────────────────────────
clean | dirty | untracked         (see sync situations below)
(+ future: staged split, stash)   + probe error / unknown
```

### Local flags (worktree)

| Flag | Meaning | How probed today |
| --- | --- | --- |
| `clean` | No staged/unstaged tracked changes and no untracked files | Porcelain empty → clean |
| `dirty` | Staged and/or unstaged modifications to **tracked** paths | Porcelain line not `??` / `!!` |
| `untracked` | Untracked (or ignored-listed as `!!`) paths present | Porcelain `??` / `!!` |

Notes:

- `clean` is shown only when neither `dirty` nor `untracked` applies (and no
  hard probe error that replaces the whole string with `error`).
- Staged vs unstaged are **not** split yet (both collapse into `dirty`).
- Stash is **not** probed yet.

### Sync flags (local branch vs remote)

| Flag / form | Intended meaning | How probed today |
| --- | --- | --- |
| `ahead:N` | Local has N commits not in the comparison ref | `rev-list --left-right --count HEAD...@{upstream}` left count |
| `behind:N` | Comparison ref has N commits not in local | same, right count |
| `no-upstream` | No **branch upstream** configured (`branch.<name>.merge` / `@{upstream}`) | `rev-parse @{upstream}` fails |
| `sync-unknown` | Upstream exists (or should) but counts could not be computed | `rev-list` failed / parse failed |
| *(none of the above)* | In sync with comparison ref (ahead=0, behind=0) | counts both zero — **no extra token** today |

Counts are against **local remote-tracking refs** (whatever `@{upstream}`
points at after the last fetch). Default `yerk status` does **not** fetch;
stale remote-tracking refs mean “ahead/behind vs last known remote”, not
necessarily live network truth. A future `status --fetch` would refresh first.

### Probe-level

| Display | Meaning |
| --- | --- |
| `-` | Change not collected (presence-only mode, or not present) |
| `error` | Git probe failed unexpectedly (alone or as an extra flag) |

---

## Display examples

| Worktree | Tracking | Typical `CHANGE` |
| --- | --- | --- |
| clean | tracking, equal | `clean` |
| clean | tracking, 2 local commits not on remote-tracking branch | `clean ahead:2` |
| clean | tracking, remote has 1 not in local | `clean behind:1` |
| clean | tracking, both sides moved | `clean ahead:1 behind:1` |
| dirty + untracked | tracking, equal | `dirty untracked` |
| clean | **no** `@{upstream}` | `clean no-upstream` |
| anything | probe blew up | `error` or `… error` |

---

## Known gap: `no-upstream` vs “unpushed”

### What operators often mean by “unpushed”

1. **Branch has upstream, local is ahead** → commits not pushed → should show
   `ahead:N` (this is the clear “unpushed” case).
2. **Branch has no upstream configured**, but a remote named `origin` (or the
   catalog remote) exists and the branch name exists only locally, or local is
   ahead of `origin/<branch>` without tracking set → still “I haven’t got
   these commits onto the server”, but git itself has no `@{upstream}`.
3. **Truly no remote** (or no comparable remote branch) → “no remote” / nothing
   to push to.

### What yerk does today

Probe picks a **comparison ref**, then counts ahead/behind:

1. `@{upstream}` if configured  
2. else `origin/<branch>` if that ref exists  
3. else `no-upstream`

So **agent-guidelines**-style checkouts (remote + `origin/main`, no branch
upstream) show **`ahead:N`** when local is ahead of `origin/main`, not a bare
`no-upstream`.

`no-upstream` still means “nothing to compare to” (no tracking and no
`origin/<branch>`), not “nothing to push” in the abstract.

Still MVP-limited: remote name is only `origin`; no `@{push}` / triangular
workflow; no catalog-remote override; detached HEAD has no branch fallback.

### Better vocabulary (proposal; not all implemented)

Prefer splitting “no tracking” from “has remote but unpushed”:

| Situation | Suggested flag(s) | Notes |
| --- | --- | --- |
| No remotes at all | `no-remote` | Stronger than no-upstream |
| Remotes exist, no `@{upstream}` | `no-upstream` (keep) | Config gap; still actionable (`git branch -u`) |
| Optional fallback: compare `HEAD` to `origin/<branch>` or remote HEAD without tracking | `ahead:N` / `behind:N` **or** `unpushed` | Needs a defined default remote (catalog `remote` / `origin`) |
| Tracking set, ahead | `ahead:N` | Prefer counts over the vague word `unpushed` |
| Tracking set, equal | *(omit sync token)* or `synced` | Today: omit |

**Recommendation:** keep combinatorial flags; use **`ahead:N`** as the primary
“unpushed commits” signal when a comparison ref exists. Add a **fallback
comparison** when `@{upstream}` is missing but `origin/<current-branch>` (or
catalog-derived remote branch) exists, so those repos show `ahead:N` instead of
only `no-upstream`. Reserve a bare `unpushed` synonym only if we want gita-like
brevity without counts.

---

## Comparison: gita

[gita](https://github.com/nosarthur/gita) (`gita ll`) is a multi-repo status +
command delegate tool. Useful reference; yerk is not a clone of its UX.

### gita local symbols (default)

| Symbol | Meaning |
| --- | --- |
| `*` | dirty (unstaged) |
| `+` | staged |
| `?` | untracked |
| `$` | stashed |

### gita remote *situation* (exactly one; colors the branch)

| Situation | Meaning (approx.) | Default color |
| --- | --- | --- |
| `no_remote` | `git diff @{u} @{0}` exit 128 — no upstream | white |
| `in_sync` | diff quiet vs upstream | green |
| `local_ahead` | local has commits upstream lacks (“good for push”) | purple |
| `remote_ahead` | upstream has commits local lacks | yellow |
| `diverged` | both sides differ from merge-base | red |

Symbols for situations (customizable): e.g. `↑` local ahead, `↓` remote ahead,
`⇕` diverged, `∅` no remote, empty when in sync.

### How gita decides ahead/behind

Not `rev-list` counts. It:

1. `git diff --quiet @{u} @{0}` → no upstream (128), in sync (0), or differ.
2. If differ: `merge-base` + further diffs to classify **local_ahead** vs
   **remote_ahead** vs **diverged**.

No numeric ahead/behind in the default UI; **one** remote situation, plus
independent dirty/staged/untracked/stash marks.

### Side-by-side

| Topic | gita | yerk (today) |
| --- | --- | --- |
| Unit of list | Registered repo paths | Catalog **projects** / **replicas** (PRJX) |
| Local dirt | staged / unstaged / untracked / stash symbols | `dirty` + `untracked` (staged merged into dirty; no stash) |
| Remote model | Single situation + color | Orthogonal flags; counts `ahead:N` `behind:N` |
| No upstream | `no_remote` / `∅` | `no-upstream` |
| Unpushed with tracking | `local_ahead` (↑) | `ahead:N` |
| Unpushed without tracking | still `no_remote` | still `no-upstream` (**same class of blind spot**) |
| Fetch | separate `gita fetch` | no auto-fetch; optional `--network` only for default-branch name |
| Output | compact symbols + color | explicit words for agents/humans/scripts |

**Takeaway:** gita’s purple “ahead” is the UX people want for “unpushed.” Yerk
already has the stronger form (`ahead:N`) **when tracking exists**. Both tools
under-inform when tracking is unset; fixing yerk means a deliberate fallback
ref, not renaming `no-upstream` to `unpushed` alone.

---

## Status scopes (CLI)

| Invocation | View |
| --- | --- |
| `yerk status` / `yerk status <project>` | **ProjectStatus** — workspace path + default-replica presence/change summary |
| `yerk status <project> <replica>` | **ReplicaStatus** — checkout path, presence, change, branch |
| `yerk status --tag <name>` | Project list filtered by declared catalog tag |
| `yerk status --presence-only` | Skip change probes |

Project table: no `REPLICA` column; `PATH` is the **project workspace**.
Replica table: full checkout path and replica name.

---

## Implementation map

```text
presence.Classify(path)     → missing | present | invalid
gitcmd.Probe(repo)          → ProbeResult { Dirty, Untracked, Clean, Ahead, Behind, NoUpstream, SyncUnknown, Branch, Error }
ProbeResult.Flags()         → space-separated CHANGE cell
project.Resolver            → fills api.ReplicaStatus / ProjectStatus.DefaultReplica
cli printers                → human tables
```

Resources: [ADR 011](../../design/decisions/011-api-resources.md).

---

## See also

- [How to check status](../how-to/check-status.md)
- [Commands reference](../reference/commands.md)
- [design/domain-and-near-term.md](../../design/domain-and-near-term.md) (status section)
- [ADR 011](../../design/decisions/011-api-resources.md)
- gita status symbols / colors: upstream README (`gita ll`)
