# How to use a fixture config directory

Status: **stub**

## Goal

Run `yerk` against a sample host config/catalog (for example
`fixtures/operator-host`) without writing `~/.config/yerk`.

## Prerequisites

- Built `yerk` binary (see `contributing/development.md`)
- Checkout of this repository (for in-tree fixtures)

## Steps (to write)

1. Set `YERK__CONFIG_DIR` to the fixture directory
2. Run `yerk status` / `yerk path` / optional `mise run host-status`
3. Unset or override when returning to real XDG config

## See also

- [Configuration reference](../reference/configuration.md)
- [Environment variables reference](../reference/envvars.md)
- Root [README](../../README.md#try-against-this-machines-fixtures)
