# How to use example config files

Status: **stub**

## Goal

Run `yerk` against a disposable config/catalog without writing `~/.config/yerk`,
using the portable samples under [`examples/`](../../examples/).

## Prerequisites

- Built `yerk` binary (see `contributing/development.md`)
- Checkout of this repository (for `examples/`)

## Steps

1. Create a throwaway directory and copy the samples:

   ```sh
   mkdir -p /tmp/yerk-dev
   cp examples/config.toml /tmp/yerk-dev/config.toml
   cp examples/catalog.toml /tmp/yerk-dev/catalog.toml
   ```

2. Edit **`config.toml`**: set `[domains]` roots for **your** host (e.g.
   `personal = "~/tree/personal"`). Edit **`catalog.toml`**: keep `path`
   relative to that domain (e.g. `devel/example`) and fix remotes. Style
   `workspace-dir` places the replica at `<workspace>/<replica>`. Do not
   commit host-private absolute paths into `examples/` (ADR 007 / 008).

3. Point yerk at the directory:

   ```sh
   export YERK__CONFIG_DIR=/tmp/yerk-dev
   yerk config show
   yerk catalog show
   yerk status
   ```

4. Unset or override `YERK__CONFIG_DIR` when returning to real XDG config.

## See also

- [Configuration reference](../reference/configuration.md)
- [Environment variables reference](../reference/envvars.md)
- [ADR 007](../../design/decisions/007-examples-and-host-local-data.md), [ADR 008](../../design/decisions/008-domain-roots-and-relative-catalog-paths.md)
- Root [README](../../README.md#quick-start)
