# How-to guides

**Task-oriented** (Diátaxis).

Concrete goals for someone who already knows the basics.
Name the problem; give practical steps; do not scaffold like a first lesson
or dump every flag.

Publication remains ad hoc plain Markdown
([ADR 006](../../design/decisions/006-ad-hoc-docs-diataxis.md)).

## Pages

| Page | Status | Intent |
| --- | --- | --- |
| [Add a project to the catalog](./add-a-project.md) | Stub | Register a project by editing `catalog.toml` |
| [Materialize a default replica](./materialize-a-replica.md) | Stub | Materialize a checkout with `yerk materialize` |
| [Create a session replica](./create-a-replica.md) | Stub | `yerk replica create` (worktree \| clone) |
| [Check presence and change status](./check-status.md) | Stub | `yerk status` (change on; `--presence-only`) |
| [Get project or replica info](./get-project-info.md) | Draft | `yerk get` / `lookup` (+ noun forms); `--output json` |
| [Explain placement](./explain-placement.md) | Draft | `yerk config resolve` — why is my style X? |
| [Refresh host project state](./update-project-state.md) | Draft | `yerk state update` — rebind from ambient |
| [Use example config files](./use-example-config.md) | Stub | Copy `examples/` into a throwaway `YERK__CONFIG_DIR` |
