# Plan: docs deadwood cleanup (+ agent role)

## Context

Docs grew through the first feature series. Expect drift: old `clone` names,
empty stubs, links into deleted plan paths, help text vs how-to mismatch.

## Goals

1. Inventory of stale/dead docs and link rot.
2. Operator-approved cleanup pass.
3. Small reusable **agent checklist/role** for future cleanups (link crawl,
   `yerk --help` vs docs, ADR status vs prose).

## Non-goals

- Full Diátaxis IA redesign (ADR 006 stays ad hoc Markdown)
- Host dogfood notes (workspace-local `.agents/`)

## Phases

0. Lock delete policy + role home.  
1. Inventory (read-only report).  
2. Apply after accept.  
3. Check in role/checklist; close-out.

## Success criteria

- [ ] No known false command names in how-tos.
- [ ] Broken internal links fixed or removed.
- [ ] Reusable cleanup checklist exists.
- [ ] Plan removable at close.
