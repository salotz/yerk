# Plan: parallel change probes

## Context

`yerk status` probes replicas **serially** today. That is correct and simple;
large catalogs feel slow. Parallelize behind the existing `gitcmd.Runner`
interface with a concurrency limit and **stable** row ordering.

Soft dep: serial path and `api` status resources already on main.

---

## Goals

1. Faster multi-replica / multi-project status when change probes run.
2. Deterministic table order (not completion order).
3. Tests that do not require a wall-clock flake harness.

## Non-goals

- Rewriting git to go-git
- Parallelizing unrelated commands without need
- Distributed/remote status agents

---

## Phases

0. Lock limit, ordering, error policy, env knob.
1. Implement parallel collection in `internal/project` (or probe helper).
2. Expand fake git / unit tests.
3. Short doc note; dogfood on host catalog.

## Success criteria

- [ ] Measurable or clearly concurrent probes under load.
- [ ] Output order matches pre-parallel expectations.
- [ ] Tests green; plan removable at close.

## Immediate next step

When In Progress: Phase 0 decisions with operator.
