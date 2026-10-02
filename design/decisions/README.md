# Decisions

Architecture decision records for `yerk`.

| ID | Title |
| --- | --- |
| [001](./001-name-and-identity.md) | Name and identity (`yerk`) |
| [002](./002-go-standalone-binary.md) | Go standalone binary |
| [003](./003-config-xdg-and-env.md) | XDG config and `YERK__` env |
| [004](./004-config-and-catalog-split.md) | Split tool config and project catalog |
| [005](./005-cli-help-and-envvars.md) | CLI help layout and env var docs (cobra) |
| [006](./006-ad-hoc-docs-diataxis.md) | Ad hoc plain Markdown docs (Diátaxis); publication later |
| [007](./007-examples-and-host-local-data.md) | `examples/` tree; no host-local fixtures in-repo |
| [008](./008-domain-roots-and-relative-catalog-paths.md) | Domain roots + relative catalog paths (superseded for resolve by 014) |
| [009](./009-workspace-subcommand-and-ensure-scope.md) | `yerk workspace ensure`; workspace dir only |
| [010](./010-catalog-tag-vocabulary.md) | Closed catalog tag vocabulary |
| [011](./011-api-resources.md) | Model-driven API resources (`internal/api`) |
| [012](./012-identifiers-and-yerk-uri.md) | Identifiers and `yerk://` URI currency |
| [013](./013-placement-policy-and-host-state.md) | Placement layers and host project state |
| [014](./014-host-local-path-model.md) | Host-local path model (successor to 008) |
| [015](./015-get-and-lookup.md) | Project/replica get and path lookup |
| [016](./016-replica-create.md) | Replica create (worktree \| clone method) |
| [017](./017-agent-context-dumps.md) | Agent context dumps (`yerk context`) |
