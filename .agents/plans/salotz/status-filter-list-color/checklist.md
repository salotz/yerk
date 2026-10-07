# Checklist — status-filter-list-color

## Phase 0 — Decisions

- [ ] Q1–Q10 locked ([decisions.md](./decisions.md))

## Phase 1 — ADR

- [ ] Draft next status-UX ADR (not 025)
- [ ] Operator accept / revise
- [ ] Domain spine status scopes: pointer only after accept

## Phase 2 — Convenience filters

- [ ] Matcher package/functions on `api.ProjectStatus` / `ReplicaStatus`
- [ ] Flags on `yerk status` (locked set)
- [ ] AND combine; empty-ok; preamble
- [ ] `--presence-only` vs change flags error
- [ ] json/yaml filtered
- [ ] Tests (temp dirs)

## Phase 3 — Replica listing

- [ ] `yerk status --replicas` (or locked alternative)
- [ ] Flatten live replicas across selected projects
- [ ] Filters at replica grain
- [ ] Tests

## Phase 4 — Color

- [ ] Auto TTY + `NO_COLOR` + `--color`
- [ ] Token colors; tabwriter alignment
- [ ] No color on json/yaml
- [ ] Tests with `--color=never` / always

## Phase 5 — Docs + close

- [ ] status-model, check-status, commands, help
- [ ] `--status=` defer noted on tag-boolean-select / owner todo
- [ ] Success criteria; delete folder; update todo
