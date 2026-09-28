# Reference

**Information-oriented** (Diátaxis).

Accurate, lookup-friendly description of the product as it is.
Prefer structure that mirrors the CLI, config files, and env registry.
Keep narrative and tutorials out of these pages.

Publication remains ad hoc plain Markdown
([ADR 006](../../design/decisions/006-ad-hoc-docs-diataxis.md)).

Static env declarations live in
[`.appinfo/meta.toml`](../../.appinfo/meta.toml) (RFC 030/031).
CLI built-in help is the runtime view of the same knobs
([ADR 005](../../design/decisions/005-cli-help-and-envvars.md)):

- `yerk <command> --help`
- `yerk help envvars`
- `yerk envvars` (live values)

## Pages

| Page | Status | Intent |
| --- | --- | --- |
| [Commands](./commands.md) | Stub | Command surface and primary flags |
| [Configuration](./configuration.md) | Stub | `config.toml`, paths, XDG layout |
| [Catalog](./catalog.md) | Stub | `catalog.toml` fields |
| [Environment variables](./envvars.md) | Stub | Pointer + summary; full list via CLI |
