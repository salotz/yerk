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

Link-time version identity (optional):

```sh
go build -ldflags "-X github.com/salotz/yerk/internal/version.Version=0.1.0" -o yerk ./cmd/yerk
```

## Test and check

```sh
mise run test
mise run check          # tests + go vet
mise run fmt            # gofmt -w
```

## Config while developing

```sh
.local/bin/yerk config path
.local/bin/yerk config example
.local/bin/yerk config show
```

Point at a throwaway file:

```sh
YERK__CONFIG=/tmp/yerk-dev.toml .local/bin/yerk status
```

## Layout reminder

| Path | Role |
| --- | --- |
| `cmd/yerk` | main |
| `internal/cli` | commands |
| `internal/config` | load TOML + env |
| `internal/workspace` | path styles |
| `internal/version` | build stamps |
| `.local/` | caches, built binary (gitignored) |
