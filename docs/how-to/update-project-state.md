# How to refresh host project state

Status: **draft**

## When

`yerk workspace ensure` / `materialize` **bind once**. If ambient placement
later changes (host config, dir-local, env) and you want the **binding** to
match again, run an explicit refresh — ambient drift alone only **warns**.

## Command name

**`yerk state update`**

Avoided **lock** (overloaded with VCS/package locks and implies freeze without
saying “recompute from ambient”). **update** means rewrite the tool-written
binding from current ambient (or an explicit style flag).

## Usage

```sh
yerk state update personal/yerk
yerk state update --all
yerk state update yerk --workspace-style project-dir
```

Then verify:

```sh
yerk config resolve personal/yerk
```

- Missing bindings are **created**.
- Existing bindings are **overwritten** (BoundAt kept; UpdatedAt refreshed).
- Does **not** mkdir workspaces or clone replicas.

## See also

- [Explain placement](./explain-placement.md)
- [ADR 013](../../design/decisions/013-placement-policy-and-host-state.md)
