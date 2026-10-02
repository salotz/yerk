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
   `personal = "~/tree/personal/devel"`). Catalog projects then default to
   `<root>/<name>`. Add `[[projects]]` rows only for path overrides. Edit
   **`catalog.toml`**: fix remotes; keep it free of host paths (ADR 014).
   Style `workspace-dir` places the replica at `<workspace>/<replica>`. Do not
   commit host-private absolute paths into `examples/` (ADR 007).

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
- [ADR 007](../../design/decisions/007-examples-and-host-local-data.md),
  [ADR 014](../../design/decisions/014-host-local-path-model.md)
- Root [README](../../README.md#quick-start)
