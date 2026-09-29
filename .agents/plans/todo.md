# Plans todo

Near-term spine: [near-term.md](./near-term.md) and
[design/domain-and-near-term.md](../../design/domain-and-near-term.md).

## Near-term (active)

- [x] P0: config / catalog split
- [x] P1: `yerk path` + presence `yerk status` (baseline)
- [x] P2: git adapter + change probe + default branch
- [x] P3: `yerk workspace ensure` + `yerk clone`
- [x] ADR 010 closed tags + `status --tag`
- [ ] **P5: Explicit API-style resources** (`internal/api` or similar) + ADR
- [ ] **P6: Status UX + change-on-disk (serial)**
  - [ ] Project list: workspace path; no REPLICA column by default
  - [ ] `yerk status [project [replica]]` scopes
  - [ ] Change status on by default; flag to disable
- [ ] **P7: Parallel change probes** (after P6)
- [ ] **P8: `--tag` on ensure/clone** (+ XOR with names/`--all`)
- [ ] P5 follow-ons: `--output json|yaml`, schemas, `yerk get`
- [ ] **P9: push/pull semantics design** (before any sync CLI)
- [ ] Optional: document copy of `examples/` → `~/.config/yerk` when XDG writable

## Later

- [ ] Implement `yerk register` (catalog mutation)
- [ ] Implement `yerk pull` / `yerk push` (after P9; single + tag)
- [ ] Local config staging on materialize (`.local` + config-dir locals)
- [ ] PRJX / bimtree domain-aware roots
- [ ] Multi-replica listing / discovery under a project workspace
- [ ] Resource backends beyond TOML files (DB, etc.)
