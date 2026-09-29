# How to check presence and change status

Status: **stub**

## Goal

List cataloged projects with **presence** on disk and git **change** flags via
`yerk status`. Change probes run by default; use `--presence-only` for a fast
presence scan.

## Prerequisites

- Non-empty catalog (or acceptance of empty output)
- For change columns: present replicas and `git` on `PATH`

## Steps

1. **All projects** (project-scoped table):

   ```sh
   yerk status
   ```

   Columns: name, domain, default-replica **presence**, **change**, and
   **project workspace** path (not the replica checkout path). No REPLICA
   column on this view.

2. **One project:**

   ```sh
   yerk status yerk
   ```

3. **One replica** (replica-scoped table: path, presence, change, branch):

   ```sh
   yerk status yerk main
   ```

4. **Select by tag:**

   ```sh
   yerk status --tag devel
   ```

   - `<name>` must be in the catalog root `tags` list (unknown tag → error)
   - Output includes `filter.tag=<name>` when filtering
   - Declared tag with no projects → “No projects matched tag …”
   - Do not combine `--tag` with project name args

5. **Presence only** (skip git change probes):

   ```sh
   yerk status --presence-only
   ```

6. Read **presence** vs **change** without conflating them: change is only
   meaningful when presence is `present`.

7. Read change flags as a **bag**, not a single enum. Common patterns:

   | CHANGE | Rough meaning |
   | --- | --- |
   | `clean` | Worktree clean; if tracking and equal, no sync token |
   | `clean ahead:2` | Clean worktree; **2 commits not on upstream** (unpushed vs tracking) |
   | `clean no-upstream` | Clean worktree; **no branch upstream configured** (not proof that a remote has your commits) |
   | `dirty untracked` | Local modifications + untracked files |

   Full flag dictionary and gita comparison:
   [Status model (explanation)](../explanation/status-model.md).

## See also

- [Status model (explanation)](../explanation/status-model.md)
- [Commands reference](../reference/commands.md)
