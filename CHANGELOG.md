# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
with modifications from [salotz RFC 002](https://github.com/salotz/rfcs/tree/master/rfcs/salotz.002_semantic-changelog)
(Growth Versioning `B.R.G`; [ADR 023](./design/decisions/023-growth-versioning.md)).
Audience is end users.

## [Unreleased]

## [0.0.2] - 2026-11-01

Inspect dumps that were human-only now share the `--output` surface.

### Improvements

- `catalog show --output json|yaml` emits a `Catalog` resource.
- `config show --output json|yaml` emits a `ToolConfig` resource (host file, workspace style, domains, host project rows, state dir).
- `envvars --output json|yaml` emits an `EnvVars` resource (live names, values, `set` / `extra`).
- `--output table` (or omit) keeps the existing human table or key/value layout on all three.

## [0.0.1] - 2026-10-02

First tagged release.

### Improvements

- Host catalog of projects (`catalog.toml`) with required `domain`, closed tag vocabulary, and `yerk://` identifiers.
- Host placement from `config.toml` (`[domains]`, optional `[[projects]]` path and style overrides) and bound project state.
- Commands: `status`, `path` / `resolve`, `get` / `lookup` (and noun forms), `workspace ensure`, `materialize`, `replica create`, `config` (`path` / `show` / `resolve`), `state update`, `catalog`, `context` / `context dir`, `envvars`, `version`.
- `--output json|yaml|table` on `get`, `lookup`, `status`, `context`, and `config resolve`.
- GitHub Release plus packslip-signed linux/x64 archive.

[Unreleased]: https://github.com/salotz/yerk/compare/v0.0.2...HEAD
[0.0.2]: https://github.com/salotz/yerk/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/salotz/yerk/releases/tag/v0.0.1
