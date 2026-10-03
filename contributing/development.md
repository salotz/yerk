# Development

## Bootstrap (host)

Check-only (no installs):

```sh
./.bootstrap/host-tool-check
```

Required host tools: `git`, `mise` (see `.bootstrap/host-tools.conf`).

## Install project tools

```sh
mise install
mise run preload
```

`preload` downloads Go module dependencies.
Go itself is pinned in `mise.toml`.

Module and build caches are directed under `.local/` (gitignored) so sandboxed
or multi-clone work does not need a writable global `GOPATH`.

## Build and run

```sh
mise run build          # → .local/bin/yerk
.local/bin/yerk version
.local/bin/yerk --help
```

Passthrough:

```sh
mise run build
.local/bin/yerk status
```

Standalone (any Go toolchain):

```sh
go build -o yerk ./cmd/yerk
```

### Product version (Growth Versioning)

Product versions use **Growth Versioning** `B.R.G`
([ADR 023](../design/decisions/023-growth-versioning.md), salotz RFC 002)—not
classic SemVer meanings. Release identity is a **git tag** `vB.R.G`; the binary
stamp is `B.R.G` without the `v`.

`mise run build` stamps via [`.tasks/build-yerk`](../.tasks/build-yerk) /
[`.tasks/go-ldflags`](../.tasks/go-ldflags):

| Situation | `Version` stamp |
| --- | --- |
| Untagged HEAD | `0.0.0-dev` |
| HEAD is exact tag `vB.R.G` | `B.R.G` |
| `YERK_VERSION=B.R.G` in env | that value (release CI) |

**Bump in practice:**

```sh
mise run version-show
# dry-run next tag (no git writes):
mise run version-bump -- growth        # G+1  (aliases: g, patch)
mise run version-bump -- regression    # R+1  (aliases: r, minor)
mise run version-bump -- breakage      # B+1, G=0  (aliases: b, major)

# cut a release on a clean HEAD, then push the tag:
mise run version-bump -- growth --tag
git push origin v0.0.1                 # or: … --tag --push
```

Without `--tag`, the task only prints current → next. With `--tag` it creates
an **annotated** tag on `HEAD` (refuses a dirty tree unless `--force`).

### Continuous integration (GitHub Actions)

| Workflow | Trigger | Does |
| --- | --- | --- |
| [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) | pull request; push to `main`/`master` | `go test`, `go vet`, stamped build |
| [`.github/workflows/release.yml`](../.github/workflows/release.yml) | push tag `vB.R.G` | test + vet + linux/x64 archive + checksums → GitHub Release, then **packslip** sign + upload `packslip.sigstore.json` |

Local equivalents:

```sh
mise run check          # test + vet
mise run build          # → .local/bin/yerk with ldflags
```

Go pin in workflows must match `mise.toml` (`1.27.1`). Bump both together.

### Release cut (tag → assets → packslip)

1. Clean tree on the commit you want to ship; `mise run check`.
2. `mise run version-bump -- growth --tag` (or `regression` / `breakage`).
3. `git push origin main` (if needed) and `git push origin vB.R.G`.
4. Workflow **release**:
   - Builds `yerk-<ver>-linux-x64` (CGO off), packs `yerk-<ver>-linux-x64.tar.gz`
     (archive member `yerk`), writes `SHA256SUMS`.
   - Creates/updates the GitHub Release with those assets.
   - Runs [`jdx/packslip@v1`](https://packslip.dev/docs/publishing/) on the
     **tarball only** (`bin: yerk`, project `github.com/salotz/yerk`), attests
     files, signs with the workflow OIDC identity, uploads
     `packslip.sigstore.json`.
5. **After the first successful release**, download the bundle and record the
   signer pin for install docs / README:

   ```sh
   curl -fsSLO "https://github.com/salotz/yerk/releases/download/vB.R.G/packslip.sigstore.json"
   packslip pin packslip.sigstore.json   # → ps1_… ; publish beside install instructions
   ```

6. Operators install with mise:

   ```sh
   mise use -g packslip:github.com/salotz/yerk
   ```

**Do not** move packslip signing to a different workflow file without planning
a consumer trust reset (packslip pins the **workflow path**
`.github/workflows/release.yml`).

## Test and check

```sh
mise run test
mise run check          # tests + go vet
mise run fmt            # gofmt -w
```

## Config while developing

```sh
.local/bin/yerk config path
.local/bin/yerk config show
.local/bin/yerk catalog path
.local/bin/yerk catalog show
```

Point at a throwaway config **directory** (both `config.toml` and `catalog.toml`),
seeded from portable examples (ADR 007 — do **not** commit host-private trees):

```sh
mkdir -p /tmp/yerk-dev
cp examples/config.toml /tmp/yerk-dev/config.toml
cp examples/catalog.toml /tmp/yerk-dev/catalog.toml
# edit [domains] roots (default <root>/<name>); catalog has no path (ADR 014)

export YERK__CONFIG_DIR=/tmp/yerk-dev
.local/bin/yerk status
.local/bin/yerk catalog show
```

Unit tests: `mise run test` (temp dirs + fake git where needed; real `git` for
`internal/gitcmd`).

## Adding an environment variable

Tool-owned names use the `YERK__` prefix (RFC 027 field style, ADR 003).
Two places must stay aligned so CLI help and static discovery say the same
thing (ADR 005, RFC 030/031):

| Layer | Path | Role |
| --- | --- | --- |
| Runtime registry | `internal/envvars` | Drives `yerk --help`, `yerk help envvars`, `yerk envvars`, and command `--help` |
| Static declarations | `.appinfo/meta.toml` | RFC 030/031 product + env registry for agents/tools without running the binary |

### Steps

1. **Name and ownership**
   - Product knob → `YERK__SOME_LEAF` (double underscore after the prefix).
   - Platform name yerk only *reads* → declare with `external = true` / `External: true` (e.g. `XDG_CONFIG_HOME`).
   - Do **not** put `YERK__*` under PRJX `[project.env-vars]` in
     `.config/_project-meta.toml` (wrong prefix; RFC 28 leaves are `PRJX__*`).
   - Do **not** add full `PRJX_*` catalogs to `.appinfo` or the help registry;
     point at RFC 28 / project metadata instead.

2. **`internal/envvars`** — add a `Var` in `All()` with:
   - `Name`, `Scope` (`tool` or `platform`)
   - `Summary` (one line) and optional `Long`
   - `Type` (`string`, `enum`, `boolean`, … — RFC 032 names only)
   - `Policy` (`warn` default; use `required` only when unset is an error — then **no** `Default`)
   - `Default` prose matching the real resolution story (not a fake literal path if the process derives it)
   - `Values` when `Type` is `enum`
   - `External` for non-owned names
   - `Global` / `Commands` (which cobra paths read it)
   - `HelpPrimary` (root + matching command `--help`)

3. **`.appinfo/meta.toml`** — add a matching `[env.vars.YERK__SOME_LEAF]`
   (or platform name) subtable with the **same** `type`, `policy`, default
   story (`default` scalar or `{ resolution = "…" }`), `values` / `aliases`
   if any, `external`, and `description` aligned with `Summary`.

4. **Read the variable in code** where needed (`os.Getenv` / config load).
   Prefer centralizing load logic in `internal/config` (or one package) rather
   than scattering getenv calls.

5. **Tests**
   - Registry/help: extend `internal/envvars` tests if the new name is
     primary or changes full-reference shape.
   - Behavior: unit-test the code path with `t.Setenv`.
   - Run `mise run test` and skim:
     - `yerk help envvars` — type, policy, default, values
     - `yerk <cmd> --help` — primary subset if `HelpPrimary`
     - `yerk envvars` — live row appears

6. **Docs (optional but preferred)**
   - Primary knobs: root [README](../README.md) configuration env table.
   - [docs/reference/envvars.md](../docs/reference/envvars.md) summary table.

### Checklist

- [ ] `internal/envvars` entry (type, policy, default, commands, primary)
- [ ] `.appinfo/meta.toml` row (same facts)
- [ ] Code reads the var
- [ ] Tests + `mise run test`
- [ ] Help shows type/policy/default (and enum values if any)
- [ ] README / docs reference if it is a primary operator knob

### Verify

```sh
mise run test
mise run build
.local/bin/yerk help envvars
.local/bin/yerk envvars
.local/bin/yerk status --help   # if primary for status
```

### References

- [ADR 003](../design/decisions/003-config-xdg-and-env.md) — XDG + `YERK__`
- [ADR 005](../design/decisions/005-cli-help-and-envvars.md) — help surfaces
- [`.appinfo/meta.toml`](../.appinfo/meta.toml) — static registry
- RFC 027 (names), 030/031 (appinfo + env tables), 032 (types/policies)

## Layout reminder

| Path | Role |
| --- | --- |
| `cmd/yerk` | main |
| `internal/cli` | commands |
| `internal/config` | tool config + catalog load |
| `internal/envvars` | env registry + help/live formatters |
| `internal/workspace` | path styles + workspace ensure |
| `internal/presence` | disk presence |
| `internal/gitcmd` | git adapter |
| `internal/project` | resolve / status |
| `internal/version` | build stamps |
| `.appinfo/meta.toml` | static product + env declarations (RFC 030/031) |
| `examples/` | portable config/catalog samples (ADR 007) |
| `.local/` | caches, built binary (gitignored) |
