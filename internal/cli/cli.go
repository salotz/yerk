// Package cli is the yerk command-line interface.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/envvars"
	"github.com/salotz/yerk/internal/gitcmd"
	"github.com/salotz/yerk/internal/id"
	"github.com/salotz/yerk/internal/presence"
	"github.com/salotz/yerk/internal/project"
	"github.com/salotz/yerk/internal/state"
	"github.com/salotz/yerk/internal/version"
	"github.com/salotz/yerk/internal/workspace"
	"github.com/spf13/cobra"
)

// IO bundles process streams so tests can capture output.
type IO struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// DefaultIO uses the process stdio streams.
func DefaultIO() IO {
	return IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}
}

// NewRoot builds the root cobra command.
func NewRoot(streams IO) *cobra.Command {
	root := &cobra.Command{
		Use:           "yerk",
		Short:         "Host multi-project manager (replicas, locals, status); speaks PRJX",
		Long:          longHelp + "\n" + envvars.FormatRootHelp(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.SetIn(streams.In)
	root.SetOut(streams.Out)
	root.SetErr(streams.Err)

	root.AddCommand(newVersionCmd(streams))
	root.AddCommand(newStatusCmd(streams))
	root.AddCommand(newPathCmd(streams))
	root.AddCommand(newGetCmd(streams))
	root.AddCommand(newLookupCmd(streams))
	root.AddCommand(newProjectCmd(streams))
	root.AddCommand(newReplicaCmd(streams))
	root.AddCommand(newWorkspaceCmd(streams))
	root.AddCommand(newMaterializeCmd(streams))
	root.AddCommand(newConfigCmd(streams))
	root.AddCommand(newCatalogCmd(streams))
	root.AddCommand(newEnvvarsCmd(streams))

	return root
}

// Execute runs the root command with process args.
func Execute(ctx context.Context, streams IO, args []string) error {
	root := NewRoot(streams)
	root.SetArgs(args)
	return root.ExecuteContext(ctx)
}

const longHelp = `yerk is a host-level multi-project manager.

It keeps a project catalog, materializes replicas into a workspace layout,
resolves paths by identifier, reports presence and change status, and
exposes get/lookup for agents and tools.
Vocabulary aligns with PRJX (project, replica, workspace).

Project identity (ADR 012): bare domain/name[/replica] or yerk://… URI.
Short unique names expand; ambiguous short names error.

Read one resource: yerk get <id>, yerk lookup <path> (--output json).

Default files:
  Tool config:  $XDG_CONFIG_HOME/yerk/config.toml
  Catalog:      $XDG_CONFIG_HOME/yerk/catalog.toml

Environment help: primary YERK__ knobs are summarized below.
Full documentation: yerk help envvars. Live values: yerk envvars.
`

// withEnv appends the primary env section for a command path to Long.
func withEnv(long, commandPath string) string {
	long = strings.TrimRight(long, "\n")
	return long + "\n\n" + envvars.FormatCommandHelp(commandPath)
}

func newEnvvarsCmd(streams IO) *cobra.Command {
	// Long = documentation (yerk help envvars / yerk envvars --help).
	// Run = live process values (app-scoped env table). ADR 005.
	return &cobra.Command{
		Use:   "envvars",
		Short: "Print yerk-relevant environment variable values for this process",
		Long:  envvars.FormatFullReference(),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			text := envvars.FormatLiveValues(nil, os.Environ())
			if !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			_, err := fmt.Fprint(streams.Out, text)
			return err
		},
	}
}

func newVersionCmd(streams IO) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version identity",
		Long:  withEnv("Print the yerk version identity string. Does not load config or catalog.", "version"),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(streams.Out, "yerk", version.String())
			return err
		},
	}
}

func newStatusCmd(streams IO) *cobra.Command {
	var (
		tagFilter    string
		presenceOnly bool
		network      bool
		gitCompat    bool // deprecated alias; ignored when presence-only is set
	)
	cmd := &cobra.Command{
		Use:   "status [project [replica]]",
		Short: "Show project or replica presence and change status",
		Long: withEnv(`Show status for catalog projects (default) or a single project/replica.

  yerk status                         → all projects (or --tag)
  yerk status <project>               → one project
  yerk status <project> <replica>     → one replica checkout
  yerk status --tag <name>            → projects with declared tag

Project rows emphasize the project workspace path and summarize the default
replica (presence/change). There is no REPLICA column on the project view;
use status <project> <replica> for replica-scoped detail.

Change probes run by default when a replica is present (dirty, untracked,
ahead/behind, …). Pass --presence-only to skip git change probes (fast).

With --network (or without --presence-only): may use git ls-remote for the
default branch when default_replica is unset (otherwise catalog override or
fallback "main"). With --presence-only and no --network, default branch is
catalog override or "main" only (no ls-remote).

--tag must appear in the catalog's top-level tags list. Same selection model
will apply to other bulk ops (workspace ensure, materialize, …).`, "status"),
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if tagFilter != "" && len(args) > 0 {
				return errors.New("status: pass --tag or project args, not both")
			}
			// Change on by default; --presence-only opts out. Legacy --git is
			// accepted as a no-op so old scripts keep working.
			withGit := !presenceOnly
			_ = gitCompat

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cat, err := config.LoadCatalog()
			if err != nil {
				return err
			}
			if len(cat.Projects) == 0 {
				fmt.Fprintln(streams.Out, "No projects in catalog.")
				catPath, _ := config.CatalogPath()
				fmt.Fprintf(streams.Out, "Catalog: %s\n", catPath)
				fmt.Fprintln(streams.Out, "Add [[projects]] entries, then re-run: yerk status")
				return nil
			}

			res, err := project.NewResolver(cfg, gitcmd.New())
			if err != nil {
				return err
			}
			// Network for default-branch discovery when probing change, or
			// when the operator asked for --network explicitly.
			useNet := network || withGit
			opts := project.StatusOptions{Git: withGit, Network: useNet}

			fmt.Fprintf(streams.Out, "workspace.style=%s\n", res.Layout.Style)

			if len(args) >= 1 {
				p, ref, err := resolveProjectArgs(cat, args)
				if err != nil {
					return err
				}
				if ref.IsReplica() {
					row, err := res.ReplicaStatus(cmd.Context(), p, ref.Replica, opts)
					if err != nil {
						return err
					}
					return printReplicaStatusTable(streams.Out, []api.ReplicaStatus{row}, withGit)
				}
				rows, err := res.ProjectStatuses(cmd.Context(), []config.Project{p}, opts)
				if err != nil {
					return err
				}
				return printProjectStatusTable(streams.Out, rows, withGit)
			}

			// Project list (all or --tag)
			var projects []config.Project
			switch {
			case tagFilter != "":
				projects, err = cat.SelectByTag(tagFilter)
				if err != nil {
					return err
				}
				fmt.Fprintf(streams.Out, "filter.tag=%s\n", tagFilter)
			default:
				projects = append([]config.Project(nil), cat.Projects...)
			}
			if len(projects) == 0 {
				if tagFilter != "" {
					fmt.Fprintf(streams.Out, "No projects matched tag %q.\n", tagFilter)
					return nil
				}
				fmt.Fprintln(streams.Out, "No projects in catalog.")
				return nil
			}

			rows, err := res.ProjectStatuses(cmd.Context(), projects, opts)
			if err != nil {
				return err
			}
			return printProjectStatusTable(streams.Out, rows, withGit)
		},
	}
	cmd.Flags().StringVar(&tagFilter, "tag", "", "Only projects with this declared catalog tag")
	cmd.Flags().BoolVar(&presenceOnly, "presence-only", false, "Skip git change probes (presence only)")
	cmd.Flags().BoolVar(&network, "network", false, "Resolve default branch via git ls-remote")
	// Deprecated: change is on by default; kept so old invocations do not error.
	cmd.Flags().BoolVar(&gitCompat, "git", false, "Deprecated: change probes are on by default (no-op)")
	_ = cmd.Flags().MarkHidden("git")
	return cmd
}

func newPathCmd(streams IO) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "path <project> [replica]",
		Aliases: []string{"resolve"},
		Short:   "Print the project workspace path, or a replica path when given",
		Long: withEnv(`Resolve on-disk paths for a catalog project (no side effects).

  yerk path <project-id>                  → project workspace directory
  yerk path <project-id> <replica>        → that replica's checkout path
  yerk path <project-id>/<replica>        → replica path (id form)
  yerk path yerk://domain/name[/replica]  → same via canonical URI

Project id: short unique name, domain/name, or yerk://… (ADR 012).
Workspace comes from host config (ADR 014): optional [domains] root + name,
or a config [[projects]] path (absolute/~/ override or relative under the root).
Effective workspace style (host / dir-local / catalog / state / env) places the
replica (workspace-dir: <workspace>/<replica>).

Default replica is not implied for identity: pass the distinguisher explicitly.`, "path"),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cat, err := config.LoadCatalog()
			if err != nil {
				return err
			}
			p, ref, err := resolveProjectArgs(cat, args)
			if err != nil {
				return err
			}
			res, err := project.NewResolver(cfg, gitcmd.New())
			if err != nil {
				return err
			}
			var path string
			if ref.IsReplica() {
				path, err = res.ReplicaPath(p, ref.Replica)
			} else {
				path, err = res.WorkspacePath(p)
			}
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, path)
			return err
		},
	}
	return cmd
}

func newGetCmd(streams IO) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Show project or replica info by identifier",
		Long: withEnv(`Read one project or replica resource by identifier (ADR 015).

  yerk get <project-id>                 → project info (not default replica)
  yerk get <project-id>/<replica>       → replica info
  yerk get yerk://domain/name[/replica]

Id forms: short unique name, domain/name, or yerk://… (ADR 012). A project id
never auto-expands to the main replica.

Default output is human key/value text. Pass --output json for a single JSON
document (apiVersion yerk/v1).

See also: yerk project get, yerk replica get, yerk lookup.`, "get"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlag(output); err != nil {
				return err
			}
			cfg, cat, res, err := loadResolver(streams)
			if err != nil {
				return err
			}
			_ = cfg
			p, ref, err := resolveProjectArgs(cat, args)
			if err != nil {
				return err
			}
			if ref.IsReplica() {
				info, err := res.ReplicaInfo(p, ref.Replica)
				if err != nil {
					return err
				}
				return printReplicaInfo(streams.Out, info, output)
			}
			info, err := res.ProjectInfo(p)
			if err != nil {
				return err
			}
			return printProjectInfo(streams.Out, info, output)
		},
	}
	cmd.Flags().StringVar(&output, "output", "", "Output format: json (default: human text)")
	return cmd
}

func newLookupCmd(streams IO) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "lookup <path>",
		Short: "Show project or replica info for an on-disk path",
		Long: withEnv(`Map a filesystem path to a catalog project or replica (ADR 015).

  yerk lookup <path>          → replica info if under a checkout root, else project
  yerk lookup .               → resolve cwd

Walk-up: any subdirectory under a known workspace or replica root matches.
Longest matching root wins when multiple projects could apply.

Default output is human key/value text. Pass --output json for JSON.

See also: yerk project lookup, yerk replica lookup, yerk get.`, "lookup"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlag(output); err != nil {
				return err
			}
			_, cat, res, err := loadResolver(streams)
			if err != nil {
				return err
			}
			result, err := res.LookupPath(cat.Projects, args[0], project.LookupAny)
			if err != nil {
				return err
			}
			return printLookupResult(streams.Out, result, output)
		},
	}
	cmd.Flags().StringVar(&output, "output", "", "Output format: json (default: human text)")
	return cmd
}

func newProjectCmd(streams IO) *cobra.Command {
	root := &cobra.Command{
		Use:   "project",
		Short: "Project-scoped get and lookup",
		Long: withEnv(`Project-scoped reads (ADR 015).

  yerk project get <project-id>       → project info
  yerk project lookup <path>          → project info for a path under its workspace

Universal forms: yerk get, yerk lookup.`, "project"),
		RunE: requireSubcommand,
	}
	root.AddCommand(newProjectGetCmd(streams))
	root.AddCommand(newProjectLookupCmd(streams))
	return root
}

func newProjectGetCmd(streams IO) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <project-id>",
		Short: "Show project info by identifier",
		Long: withEnv(`Read project info by id (no replica segment).

  yerk project get personal/yerk
  yerk project get yerk://personal/yerk

Rejects replica ids; use yerk replica get or yerk get <id>/<replica>.`, "project get"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlag(output); err != nil {
				return err
			}
			_, cat, res, err := loadResolver(streams)
			if err != nil {
				return err
			}
			p, ref, err := resolveProjectArgs(cat, args)
			if err != nil {
				return err
			}
			if ref.IsReplica() {
				return fmt.Errorf("project get: %q includes a replica; use yerk replica get or yerk get", args[0])
			}
			info, err := res.ProjectInfo(p)
			if err != nil {
				return err
			}
			return printProjectInfo(streams.Out, info, output)
		},
	}
	cmd.Flags().StringVar(&output, "output", "", "Output format: json (default: human text)")
	return cmd
}

func newProjectLookupCmd(streams IO) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "lookup <path>",
		Short: "Show project info for an on-disk path",
		Long: withEnv(`Map a path under a project workspace to project info (ADR 015).

Paths under a replica still resolve to the owning project.`, "project lookup"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlag(output); err != nil {
				return err
			}
			_, cat, res, err := loadResolver(streams)
			if err != nil {
				return err
			}
			result, err := res.LookupPath(cat.Projects, args[0], project.LookupProject)
			if err != nil {
				return err
			}
			return printLookupResult(streams.Out, result, output)
		},
	}
	cmd.Flags().StringVar(&output, "output", "", "Output format: json (default: human text)")
	return cmd
}

func newReplicaCmd(streams IO) *cobra.Command {
	root := &cobra.Command{
		Use:   "replica",
		Short: "Replica-scoped get and lookup",
		Long: withEnv(`Replica-scoped reads (ADR 015).

  yerk replica get <project-id> <replica>
  yerk replica get <project-id>/<replica>
  yerk replica lookup <path>

Universal forms: yerk get, yerk lookup.`, "replica"),
		RunE: requireSubcommand,
	}
	root.AddCommand(newReplicaGetCmd(streams))
	root.AddCommand(newReplicaLookupCmd(streams))
	return root
}

func newReplicaGetCmd(streams IO) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <project-id> [replica]",
		Short: "Show replica info by identifier",
		Long: withEnv(`Read replica info by id.

  yerk replica get personal/yerk main
  yerk replica get personal/yerk/main
  yerk replica get yerk://personal/yerk/main

Replica distinguisher is required (no default-replica expansion).`, "replica get"),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlag(output); err != nil {
				return err
			}
			_, cat, res, err := loadResolver(streams)
			if err != nil {
				return err
			}
			p, ref, err := resolveProjectArgs(cat, args)
			if err != nil {
				return err
			}
			if !ref.IsReplica() {
				return errors.New("replica get: replica distinguisher required (id third segment or second argument)")
			}
			info, err := res.ReplicaInfo(p, ref.Replica)
			if err != nil {
				return err
			}
			return printReplicaInfo(streams.Out, info, output)
		},
	}
	cmd.Flags().StringVar(&output, "output", "", "Output format: json (default: human text)")
	return cmd
}

func newReplicaLookupCmd(streams IO) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "lookup <path>",
		Short: "Show replica info for an on-disk path",
		Long: withEnv(`Map a path under a replica checkout to replica info (ADR 015).

Errors if the path is only under a project workspace and not under a replica root.`, "replica lookup"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlag(output); err != nil {
				return err
			}
			_, cat, res, err := loadResolver(streams)
			if err != nil {
				return err
			}
			result, err := res.LookupPath(cat.Projects, args[0], project.LookupReplica)
			if err != nil {
				return err
			}
			return printLookupResult(streams.Out, result, output)
		},
	}
	cmd.Flags().StringVar(&output, "output", "", "Output format: json (default: human text)")
	return cmd
}

// loadResolver loads host config + catalog and builds a Resolver (stderr warnings).
func loadResolver(streams IO) (config.Config, config.Catalog, project.Resolver, error) {
	cfg, err := config.Load()
	if err != nil {
		return config.Config{}, config.Catalog{}, project.Resolver{}, err
	}
	cat, err := config.LoadCatalog()
	if err != nil {
		return config.Config{}, config.Catalog{}, project.Resolver{}, err
	}
	res, err := project.NewResolver(cfg, gitcmd.New())
	if err != nil {
		return config.Config{}, config.Catalog{}, project.Resolver{}, err
	}
	res.Warn = streams.Err
	return cfg, cat, res, nil
}

func validateOutputFlag(output string) error {
	switch strings.TrimSpace(output) {
	case "", "json":
		return nil
	default:
		return fmt.Errorf("unsupported --output %q (want json or omit for human text)", output)
	}
}

func printLookupResult(w io.Writer, result project.LookupResult, output string) error {
	if result.Replica != nil {
		return printReplicaInfo(w, *result.Replica, output)
	}
	if result.Project != nil {
		return printProjectInfo(w, *result.Project, output)
	}
	return errors.New("lookup: empty result")
}

func printProjectInfo(w io.Writer, info api.ProjectInfo, output string) error {
	if strings.TrimSpace(output) == "json" {
		return writeJSON(w, info)
	}
	fmt.Fprintf(w, "kind:\t%s\n", info.Kind)
	fmt.Fprintf(w, "uri:\t%s\n", dash(info.URI))
	fmt.Fprintf(w, "name:\t%s\n", info.Name)
	fmt.Fprintf(w, "domain:\t%s\n", dash(info.Domain))
	fmt.Fprintf(w, "remote:\t%s\n", dash(info.Remote))
	fmt.Fprintf(w, "defaultReplica:\t%s\n", dash(info.DefaultReplica))
	fmt.Fprintf(w, "tags:\t%s\n", dash(strings.Join(info.Tags, ",")))
	fmt.Fprintf(w, "workspacePath:\t%s\n", info.WorkspacePath)
	fmt.Fprintf(w, "workspacePresence:\t%s\n", dash(string(info.WorkspacePresence)))
	if info.Placement != nil {
		fmt.Fprintf(w, "placement.style:\t%s\n", dash(info.Placement.Style))
		fmt.Fprintf(w, "placement.bound:\t%t\n", info.Placement.Bound)
		if len(info.Placement.Warnings) > 0 {
			fmt.Fprintf(w, "placement.warnings:\t%s\n", strings.Join(info.Placement.Warnings, "; "))
		}
	}
	if info.MatchedPath != "" {
		fmt.Fprintf(w, "matchedPath:\t%s\n", info.MatchedPath)
	}
	return nil
}

func printReplicaInfo(w io.Writer, info api.ReplicaInfo, output string) error {
	if strings.TrimSpace(output) == "json" {
		return writeJSON(w, info)
	}
	fmt.Fprintf(w, "kind:\t%s\n", info.Kind)
	fmt.Fprintf(w, "uri:\t%s\n", dash(info.URI))
	fmt.Fprintf(w, "project:\t%s\n", info.Project)
	fmt.Fprintf(w, "replica:\t%s\n", info.Replica)
	fmt.Fprintf(w, "domain:\t%s\n", dash(info.Domain))
	fmt.Fprintf(w, "remote:\t%s\n", dash(info.Remote))
	fmt.Fprintf(w, "tags:\t%s\n", dash(strings.Join(info.Tags, ",")))
	fmt.Fprintf(w, "path:\t%s\n", info.Path)
	fmt.Fprintf(w, "presence:\t%s\n", dash(string(info.Presence)))
	fmt.Fprintf(w, "workspacePath:\t%s\n", dash(info.WorkspacePath))
	if info.Placement != nil {
		fmt.Fprintf(w, "placement.style:\t%s\n", dash(info.Placement.Style))
		fmt.Fprintf(w, "placement.bound:\t%t\n", info.Placement.Bound)
		if len(info.Placement.Warnings) > 0 {
			fmt.Fprintf(w, "placement.warnings:\t%s\n", strings.Join(info.Placement.Warnings, "; "))
		}
	}
	if info.MatchedPath != "" {
		fmt.Fprintf(w, "matchedPath:\t%s\n", info.MatchedPath)
	}
	return nil
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	return nil
}

func newWorkspaceCmd(streams IO) *cobra.Command {
	root := &cobra.Command{
		Use:   "workspace",
		Short: "Project workspace operations (layout dirs; not git checkouts)",
		Long: withEnv(`Operate on project workspace directories (the folders that own replicas).

Subcommands create or inspect layout only. They do not create replica
checkouts — use yerk materialize for that.

  yerk workspace ensure <project-id>...   mkdir named project workspace dirs
  yerk workspace ensure --all             mkdir every catalog project workspace`, "workspace"),
		RunE: requireSubcommand,
	}
	root.AddCommand(newWorkspaceEnsureCmd(streams))
	return root
}

func newWorkspaceEnsureCmd(streams IO) *cobra.Command {
	var (
		all       bool
		styleFlag string
	)
	cmd := &cobra.Command{
		Use:   "ensure [project...]",
		Short: "Create project workspace directories (mkdir; no replica checkout)",
		Long: withEnv(`Materialize project workspace directories (mkdir -p only).

Creates the catalog project workspace path for each named project.
That is the directory that owns replicas (e.g. …/devel/yerk), not a replica
leaf (…/yerk/main).

On first successful ensure for a project, writes host project state (bound
workspace style) under $XDG_STATE_HOME/yerk (or YERK__STATE_DIR). Already
initialized projects are a state no-op.

With --all, ensure every project in the catalog. Bare ensure with no names
and no --all is an error (bulk mkdir is opt-in).

--workspace-style sets an explicit style for this invocation; if it contradicts
bound state, ensure errors (no silent rebind).

Never deletes. Does not create replica directories and does not run git.
Does not need git or --network.

See also: yerk path <project>, yerk materialize <project>.`, "workspace ensure"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if all && len(args) > 0 {
				return errors.New("workspace ensure: pass project names or --all, not both")
			}
			if !all && len(args) == 0 {
				return errors.New("workspace ensure: name one or more projects, or pass --all")
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cat, err := config.LoadCatalog()
			if err != nil {
				return err
			}
			projects, err := selectProjects(cat, args, all)
			if err != nil {
				return err
			}
			res, err := project.NewResolver(cfg, gitcmd.New())
			if err != nil {
				return err
			}
			res.CLIStyle = styleFlag
			res.Warn = streams.Err
			for _, p := range projects {
				path, err := res.WorkspacePath(p)
				if err != nil {
					return fmt.Errorf("%s: %w", p.ID(), err)
				}
				if err := workspace.EnsureDir(path); err != nil {
					return err
				}
				if err := res.BindOnInit(p); err != nil {
					return fmt.Errorf("%s: bind state: %w", p.ID(), err)
				}
				fmt.Fprintf(streams.Out, "ensured %s -> %s\n", p.ID(), path)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Ensure every project workspace in the catalog")
	cmd.Flags().StringVar(&styleFlag, "workspace-style", "", "Explicit workspace style (errors if contradicts bound state)")
	return cmd
}

func newMaterializeCmd(streams IO) *cobra.Command {
	var (
		replicaFlag string
		tagFilter   string
		all         bool
		network     bool
		styleFlag   string
	)
	cmd := &cobra.Command{
		Use:   "materialize [project [replica]]",
		Short: "Materialize replica(s) from project remote(s) via git clone",
		Long: withEnv(`Materialize replica checkouts from catalog project remotes (git clone under the hood).

Product command is materialize (no top-level clone alias). Session spin-out
(worktree vs clone method) is a later replica create command.

Single project (id forms: short unique name, domain/name, yerk://…):
  yerk materialize <project-id> [replica]
  yerk materialize <project-id> --replica <name>
  yerk materialize <project-id>/<replica>

Bulk (opt-in; mutually exclusive selectors):
  yerk materialize --all
  yerk materialize --tag <name>

Bulk materializes each selected project's default replica unless --replica is set
(same distinguisher applied to every selected project). Bare materialize with no
names and no --all/--tag is an error.

If the destination is already a usable git checkout (presence=present), materialize
skips git and reports that the replica is already present (still prints the
path). A path that exists but is not a usable checkout (invalid) remains an
error. Missing destinations are cloned as usual.

Ensures parent directories, refuses non-empty non-checkout destinations, then
runs git clone (optionally --branch). When replica is omitted per project,
resolves the remote default branch via git ls-remote --symref (then catalog
default_replica / "main").

--tag must be a declared catalog tag. Unknown tags error. A declared tag
with zero matching projects is an error for materialize (empty mutate selection).
--all with an empty catalog is an error.

Requires git on PATH.`, "materialize"),
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cat, err := config.LoadCatalog()
			if err != nil {
				return err
			}
			projects, sharedReplica, err := resolveMaterializeSelection(cat, args, all, tagFilter, replicaFlag)
			if err != nil {
				return err
			}
			git := gitcmd.New()
			res, err := project.NewResolver(cfg, git)
			if err != nil {
				return err
			}
			res.CLIStyle = styleFlag
			res.Warn = streams.Err
			ctx := cmd.Context()
			var firstErr error
			okCount := 0
			for _, p := range projects {
				if err := materializeOne(ctx, streams, res, git, p, sharedReplica, network); err != nil {
					fmt.Fprintf(streams.Err, "error: %v\n", err)
					if firstErr == nil {
						firstErr = err
					}
					continue
				}
				okCount++
			}
			if firstErr != nil {
				if okCount > 0 {
					return fmt.Errorf("materialize: %d ok, with errors: %w", okCount, firstErr)
				}
				return firstErr
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&replicaFlag, "replica", "", "Replica distinguisher / branch (default: per-project remote HEAD)")
	cmd.Flags().StringVar(&tagFilter, "tag", "", "Materialize every project with this declared catalog tag")
	cmd.Flags().BoolVar(&all, "all", false, "Materialize every project in the catalog")
	cmd.Flags().BoolVar(&network, "network", true, "Resolve default branch via git ls-remote when replica omitted")
	cmd.Flags().StringVar(&styleFlag, "workspace-style", "", "Explicit workspace style (errors if contradicts bound state)")
	return cmd
}

// cloneOne materializes one project's replica. If the path is already a
// usable checkout, it reports already-present and succeeds without git clone.
func materializeOne(ctx context.Context, streams IO, res project.Resolver, git gitcmd.Runner, p config.Project, sharedReplica string, network bool) error {
	label := p.ID()
	if strings.TrimSpace(p.Remote) == "" {
		return fmt.Errorf("project %q has empty remote", label)
	}
	replica := sharedReplica
	var err error
	if replica == "" {
		replica, err = res.DefaultReplicaName(ctx, p, network)
		if err != nil {
			return fmt.Errorf("%s: %w", label, err)
		}
	}
	path, err := res.ReplicaPath(p, replica)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}

	switch presence.Classify(path) {
	case presence.Present:
		if err := res.BindOnInit(p); err != nil {
			return fmt.Errorf("%s: bind state: %w", label, err)
		}
		fmt.Fprintf(streams.Err, "already present %s replica %s -> %s\n", label, replica, path)
		fmt.Fprintln(streams.Out, path)
		return nil
	case presence.Invalid:
		return fmt.Errorf("%s: replica path exists but is not a usable git checkout: %s", label, path)
	}

	if err := workspace.EnsureParents(path); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	fmt.Fprintf(streams.Err, "materializing %s branch %s from %s -> %s\n", label, replica, p.Remote, path)
	if err := git.Clone(ctx, p.Remote, path, replica); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if err := res.BindOnInit(p); err != nil {
		return fmt.Errorf("%s: bind state: %w", label, err)
	}
	fmt.Fprintln(streams.Out, path)
	return nil
}

// resolveCloneSelection picks projects and an optional shared replica name.
// Modes (XOR): single/multi handled as single project+optional replica args,
// --all, or --tag. Mutate empty selection is an error.
func resolveMaterializeSelection(cat config.Catalog, args []string, all bool, tag, replicaFlag string) ([]config.Project, string, error) {
	replicaFlag = strings.TrimSpace(replicaFlag)
	tag = strings.TrimSpace(tag)
	nSel := 0
	if all {
		nSel++
	}
	if tag != "" {
		nSel++
	}
	if len(args) > 0 {
		nSel++
	}
	if nSel > 1 {
		return nil, "", errors.New("materialize: pass project args, --all, or --tag, not a combination")
	}
	if nSel == 0 {
		return nil, "", errors.New("materialize: name a project, or pass --all or --tag")
	}

	if all {
		projects, err := selectProjects(cat, nil, true)
		if err != nil {
			return nil, "", err
		}
		return projects, replicaFlag, nil
	}
	if tag != "" {
		projects, err := cat.SelectByTag(tag)
		if err != nil {
			return nil, "", err
		}
		if len(projects) == 0 {
			return nil, "", fmt.Errorf("materialize: no projects matched tag %q", tag)
		}
		return projects, replicaFlag, nil
	}

	if len(args) == 0 {
		return nil, "", errors.New("materialize: name a project, or pass --all or --tag")
	}
	if len(args) > 2 {
		return nil, "", errors.New("materialize: too many arguments (use --all or --tag for bulk)")
	}
	p, ref, err := resolveProjectArgs(cat, args)
	if err != nil {
		return nil, "", err
	}
	replica := replicaFlag
	if ref.IsReplica() {
		if replicaFlag != "" && replicaFlag != ref.Replica {
			return nil, "", errors.New("materialize: pass replica as argument or --replica, not both with different values")
		}
		replica = ref.Replica
	}
	return []config.Project{p}, replica, nil
}

// resolveProjectArgs expands CLI project args into a catalog row + id.Ref.
// Accepts one arg (id may include replica) or two args (project id + replica).
func resolveProjectArgs(cat config.Catalog, args []string) (config.Project, id.Ref, error) {
	if len(args) == 0 {
		return config.Project{}, id.Ref{}, errors.New("project identifier required")
	}
	if len(args) > 2 {
		return config.Project{}, id.Ref{}, errors.New("too many arguments")
	}
	p, ref, err := cat.Resolve(args[0])
	if err != nil {
		return config.Project{}, id.Ref{}, err
	}
	if len(args) == 2 {
		rep := strings.TrimSpace(args[1])
		if rep == "" {
			return config.Project{}, id.Ref{}, errors.New("empty replica distinguisher")
		}
		if ref.IsReplica() && ref.Replica != rep {
			return config.Project{}, id.Ref{}, fmt.Errorf("replica specified twice with different values (%q vs %q)", ref.Replica, rep)
		}
		if err := validateReplicaSegment(rep); err != nil {
			return config.Project{}, id.Ref{}, err
		}
		ref.Replica = rep
	}
	return p, ref, nil
}

func validateReplicaSegment(s string) error {
	r, err := id.Parse(s)
	if err != nil || r.Domain != "" || r.Replica != "" {
		return fmt.Errorf("invalid replica distinguisher %q", s)
	}
	return nil
}

func newConfigCmd(streams IO) *cobra.Command {
	root := &cobra.Command{
		Use:   "config",
		Short: "Show tool config path and summary",
		Long: withEnv(`Inspect tool/host configuration (config.toml), not the project catalog.

Subcommands: path, show.

Portable sample documents live in the repository under examples/ (not a CLI
subcommand).`, "config"),
		RunE: requireSubcommand,
	}
	root.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the resolved tool config file path",
		Long:  withEnv("Print the absolute path yerk will use for config.toml.", "config path"),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.FilePath()
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, path)
			return err
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Load and summarize the active tool config",
		Long:  withEnv("Load config.toml (and env overlays) and print a short summary.", "config show"),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			path, _ := config.FilePath()
			fmt.Fprintf(streams.Out, "file: %s\n", path)
			fmt.Fprintf(streams.Out, "workspace.style: %s\n", cfg.Workspace.Style)
			if len(cfg.Domains) == 0 {
				fmt.Fprintln(streams.Out, "domains: (none)")
			} else {
				fmt.Fprintln(streams.Out, "domains:")
				names := make([]string, 0, len(cfg.Domains))
				for d := range cfg.Domains {
					names = append(names, d)
				}
				sort.Strings(names)
				for _, d := range names {
					raw := cfg.Domains[d]
					abs, err := config.ExpandUser(strings.TrimSpace(raw))
					if err != nil {
						fmt.Fprintf(streams.Out, "  %s: %s (expand error: %v)\n", d, raw, err)
						continue
					}
					if abs != raw {
						fmt.Fprintf(streams.Out, "  %s: %s → %s\n", d, raw, abs)
					} else {
						fmt.Fprintf(streams.Out, "  %s: %s\n", d, raw)
					}
				}
			}
			if len(cfg.Projects) == 0 {
				fmt.Fprintln(streams.Out, "projects: (none)")
			} else {
				fmt.Fprintln(streams.Out, "projects:")
				// stable order by bare id
				type row struct{ id, path, style string }
				rows := make([]row, 0, len(cfg.Projects))
				for _, hp := range cfg.Projects {
					rows = append(rows, row{id: hp.ID(), path: hp.Path, style: hp.WorkspaceStyle})
				}
				sort.Slice(rows, func(i, j int) bool { return rows[i].id < rows[j].id })
				for _, r := range rows {
					extra := ""
					if s := strings.TrimSpace(r.style); s != "" {
						extra = " style=" + s
					}
					raw := strings.TrimSpace(r.path)
					if raw == "" {
						fmt.Fprintf(streams.Out, "  %s: (default <domain-root>/<name>)%s\n", r.id, extra)
						continue
					}
					abs, err := config.ExpandUser(raw)
					if err != nil {
						fmt.Fprintf(streams.Out, "  %s: %s (expand error: %v)%s\n", r.id, r.path, err, extra)
						continue
					}
					if abs != raw {
						fmt.Fprintf(streams.Out, "  %s: %s → %s%s\n", r.id, r.path, abs, extra)
					} else {
						fmt.Fprintf(streams.Out, "  %s: %s%s\n", r.id, r.path, extra)
					}
				}
			}
			fmt.Fprintln(streams.Out, "placement: [domains] default <root>/<name> or [[projects]] path (ADR 014); style from placement layers (ADR 013)")
			stateDir, err := stateDirOrNote()
			if err != nil {
				fmt.Fprintf(streams.Out, "state.dir: (%v)\n", err)
			} else {
				fmt.Fprintf(streams.Out, "state.dir: %s\n", stateDir)
			}
			return nil
		},
	})
	return root
}

func newCatalogCmd(streams IO) *cobra.Command {
	root := &cobra.Command{
		Use:   "catalog",
		Short: "Show catalog path and table of projects",
		Long: withEnv(`Inspect the project catalog (catalog.toml), not tool config.

Subcommands: path, show.

Portable sample documents live in the repository under examples/ (not a CLI
subcommand).`, "catalog"),
		RunE: requireSubcommand,
	}
	root.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the resolved catalog file path",
		Long:  withEnv("Print the absolute path yerk will use for catalog.toml.", "catalog path"),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.CatalogPath()
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Out, path)
			return err
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print catalog projects as a table",
		Long: withEnv(`Load catalog.toml and print registered projects as a table.

Prints the declared tag vocabulary, then columns:
NAME, DOMAIN, DEFAULT_REPLICA, TAGS, REMOTE.
Workspace paths are host-local (config.toml [domains] and/or [[projects]]; ADR 014).
Project tags must be members of the top-level catalog tags list (ADR 010).`, "catalog show"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cat, err := config.LoadCatalog()
			if err != nil {
				return err
			}
			path, _ := config.CatalogPath()
			fmt.Fprintf(streams.Out, "file: %s\n", path)
			if len(cat.Tags) == 0 {
				fmt.Fprintln(streams.Out, "tags: (none declared)")
			} else {
				fmt.Fprintf(streams.Out, "tags: %s\n", strings.Join(cat.Tags, ", "))
			}
			if len(cat.Projects) == 0 {
				fmt.Fprintln(streams.Out, "projects: 0")
				return nil
			}
			fmt.Fprintf(streams.Out, "projects: %d\n", len(cat.Projects))
			tw := tabwriter.NewWriter(streams.Out, 0, 4, 2, ' ', 0)
			fmt.Fprintf(tw, "NAME\tDOMAIN\tDEFAULT_REPLICA\tTAGS\tREMOTE\n")
			for _, p := range cat.Projects {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					p.Name,
					dash(p.Domain),
					dash(p.DefaultReplica),
					dash(strings.Join(p.Tags, ",")),
					dash(p.Remote),
				)
			}
			return tw.Flush()
		},
	})
	return root
}

func stateDirOrNote() (string, error) {
	return state.Dir()
}

// requireSubcommand prints the command's help when invoked with no args
// (e.g. `yerk config`), and errors on unknown trailing args.
func requireSubcommand(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	}
	return cmd.Help()
}

// selectProjects returns catalog rows for the given identifiers, or the full
// catalog when all is true. Caller must already enforce names XOR all.
// Each name is expanded via Catalog.Resolve (short / bare / URI).
func selectProjects(cat config.Catalog, names []string, all bool) ([]config.Project, error) {
	if all {
		if len(cat.Projects) == 0 {
			return nil, errors.New("catalog is empty")
		}
		out := make([]config.Project, len(cat.Projects))
		copy(out, cat.Projects)
		return out, nil
	}
	if len(names) == 0 {
		return nil, errors.New("no projects selected")
	}
	var out []config.Project
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		p, ref, err := cat.Resolve(name)
		if err != nil {
			return nil, err
		}
		if ref.IsReplica() {
			return nil, fmt.Errorf("%q: workspace selection is project-scoped (omit replica segment)", name)
		}
		pid := p.ID()
		if _, dup := seen[pid]; dup {
			continue
		}
		seen[pid] = struct{}{}
		out = append(out, p)
	}
	return out, nil
}

// printProjectStatusTable writes the human project-scoped status table.
// PATH is the project workspace; PRESENCE/CHANGE summarize the default replica.
func printProjectStatusTable(w io.Writer, rows []api.ProjectStatus, withGit bool) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if withGit {
		fmt.Fprintf(tw, "NAME\tDOMAIN\tPRESENCE\tCHANGE\tPATH\n")
		for _, row := range rows {
			pres, change := projectReplicaCols(row)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
				row.Name, dash(row.Domain), pres, change, row.WorkspacePath)
		}
	} else {
		fmt.Fprintf(tw, "NAME\tDOMAIN\tPRESENCE\tPATH\n")
		for _, row := range rows {
			pres, _ := projectReplicaCols(row)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
				row.Name, dash(row.Domain), pres, row.WorkspacePath)
		}
	}
	return tw.Flush()
}

func projectReplicaCols(row api.ProjectStatus) (presence, change string) {
	if row.DefaultReplica == nil {
		return "-", "-"
	}
	pres := string(row.DefaultReplica.Presence)
	if pres == "" {
		pres = "-"
	}
	change = row.DefaultReplica.Change
	if change == "" {
		change = "-"
	}
	return pres, change
}

// printReplicaStatusTable writes a replica-scoped status table.
func printReplicaStatusTable(w io.Writer, rows []api.ReplicaStatus, withGit bool) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if withGit {
		fmt.Fprintf(tw, "PROJECT\tREPLICA\tDOMAIN\tPRESENCE\tCHANGE\tBRANCH\tPATH\n")
		for _, row := range rows {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				row.Project, row.Replica, dash(row.Domain), row.Presence,
				dash(row.Change), dash(row.Branch), row.Path)
		}
	} else {
		fmt.Fprintf(tw, "PROJECT\tREPLICA\tDOMAIN\tPRESENCE\tPATH\n")
		for _, row := range rows {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
				row.Project, row.Replica, dash(row.Domain), row.Presence, row.Path)
		}
	}
	return tw.Flush()
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
