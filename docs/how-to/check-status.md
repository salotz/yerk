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

1. **All projects** (project-scoped table with **overall** rollup):

   ```sh
   yerk status
   ```

   Columns: name, domain, **overall presence**, **overall change**, replica
   **count**, and **project workspace** path. Overall looks across every
   **live** replica under the workspace (plus default/main even if missing).
   Change prefers divergence (dirty, untracked, ahead/behind, …) over clean.

2. **One project** — summary + **all replicas**:

   ```sh
   yerk status yerk
   yerk status personal/yerk
   ```

   Prints overall presence/change, then a replica table (each live checkout
   probed separately). Session worktrees from `replica create` show up here.

3. **One replica** (replica-scoped table only):

   ```sh
   yerk status yerk main
   yerk status personal/yerk/feat
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

6. **Structured output** (agents; ADR 019):

   ```sh
   yerk status yerk --output json
   yerk status --tag devel --output yaml
   ```

   Single project → one `ProjectStatus` document; multi-project → a list.
   `--output table` keeps the human table/detail layout.

7. Read **presence** vs **change** without conflating them: change is only
   meaningful when presence is `present` (or overall has present replicas).

8. Read change flags as a **bag**, not a single enum. Common patterns:

   | CHANGE | Rough meaning |
   | --- | --- |
   | `clean` | Worktree clean; if tracking and equal, no sync token |
   | `clean ahead:2` | Clean worktree; **2 commits not on upstream** |
   | `dirty untracked` | Local modifications + untracked files |
   | overall `dirty ahead:1` | At least one replica dirty and/or ahead (clean dropped) |

   Full flag dictionary and gita comparison:
   [Status model (explanation)](../explanation/status-model.md).

## See also

- [Status model (explanation)](../explanation/status-model.md)
- [Commands reference](../reference/commands.md)
- [Create a session replica](./create-a-replica.md)
