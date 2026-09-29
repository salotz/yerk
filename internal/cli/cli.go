// Package cli is the yerk command-line interface.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/envvars"
	"github.com/salotz/yerk/internal/gitcmd"
	"github.com/salotz/yerk/internal/project"
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
	root.AddCommand(newWorkspaceCmd(streams))
	root.AddCommand(newCloneCmd(streams))
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
resolves paths by name, and reports presence and change status.
Vocabulary aligns with PRJX (project, replica, workspace).

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
		tagFilter string
		withGit   bool
		network   bool
	)
	cmd := &cobra.Command{
		Use:   "status",
		Short: "List catalog projects with presence (and optional change) status",
		Long: withEnv(`List projects from catalog.toml with presence status for the default replica.

Without --git: catalog + on-disk presence only (fast).
With --git: also probe change flags (dirty, untracked, ahead/behind, …).
With --network or --git: may use git ls-remote for default branch when
default_replica is unset (otherwise catalog override or fallback "main").

--tag selects only projects that list that declared catalog tag
(must appear in the catalog's top-level tags list). Same selection model
will apply to other bulk ops (workspace ensure, clone, …).`, "status"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			var projects []config.Project
			if tagFilter != "" {
				projects, err = cat.SelectByTag(tagFilter)
				if err != nil {
					return err
				}
			} else {
				projects = append([]config.Project(nil), cat.Projects...)
			}
			if len(projects) == 0 {
				fmt.Fprintf(streams.Out, "No projects matched tag %q.\n", tagFilter)
				return nil
			}

			res, err := project.NewResolver(cfg, gitcmd.New())
			if err != nil {
				return err
			}
			// Default branch via network only when --git or --network.
			useNet := network || withGit
			rows, err := res.Status(cmd.Context(), projects, project.StatusOptions{
				Git:     withGit,
				Network: useNet,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(streams.Out, "workspace.style=%s\n", res.Layout.Style)
			if tagFilter != "" {
				fmt.Fprintf(streams.Out, "filter.tag=%s\n", tagFilter)
			}

			tw := tabwriter.NewWriter(streams.Out, 0, 4, 2, ' ', 0)
			if withGit {
				fmt.Fprintf(tw, "NAME\tDOMAIN\tREPLICA\tPRESENCE\tCHANGE\tBRANCH\tPATH\n")
				for _, row := range rows {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						row.Name, dash(row.Domain), row.Replica, row.Presence,
						row.Change, row.Branch, row.Path)
				}
			} else {
				fmt.Fprintf(tw, "NAME\tDOMAIN\tREPLICA\tPRESENCE\tPATH\n")
				for _, row := range rows {
					fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
						row.Name, dash(row.Domain), row.Replica, row.Presence, row.Path)
				}
			}
			return tw.Flush()
		},
	}
	cmd.Flags().StringVar(&tagFilter, "tag", "", "Only projects with this declared catalog tag")
	cmd.Flags().BoolVar(&withGit, "git", false, "Probe change status with git")
	cmd.Flags().BoolVar(&network, "network", false, "Resolve default branch via git ls-remote")
	return cmd
}

func newPathCmd(streams IO) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "path <project> [replica]",
		Aliases: []string{"resolve"},
		Short:   "Print the project workspace path, or a replica path when given",
		Long: withEnv(`Resolve on-disk paths for a catalog project (no side effects).

  yerk path <project>           → project workspace directory
  yerk path <project> <replica> → that replica's checkout path

Workspace is catalog path relative to [domains.<domain>] (or absolute).
Style places the replica under the workspace (workspace-dir: <workspace>/<replica>).

Default replica is not implied: pass the distinguisher explicitly (e.g. main).`, "path"),
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
			p, ok := cat.Find(args[0])
			if !ok {
				return fmt.Errorf("project %q not in catalog", args[0])
			}
			res, err := project.NewResolver(cfg, gitcmd.New())
			if err != nil {
				return err
			}
			var path string
			if len(args) == 2 {
				path, err = res.ReplicaPath(p, args[1])
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

func newWorkspaceCmd(streams IO) *cobra.Command {
	root := &cobra.Command{
		Use:   "workspace",
		Short: "Project workspace operations (layout dirs; not git checkouts)",
		Long: withEnv(`Operate on project workspace directories (the folders that own replicas).

Subcommands materialize or inspect layout only. They do not create replica
checkouts — use yerk clone for that.

  yerk workspace ensure <project>...   mkdir named project workspace dirs
  yerk workspace ensure --all          mkdir every catalog project workspace`, "workspace"),
		RunE: requireSubcommand,
	}
	root.AddCommand(newWorkspaceEnsureCmd(streams))
	return root
}

func newWorkspaceEnsureCmd(streams IO) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "ensure [project...]",
		Short: "Create project workspace directories (mkdir; no replica, no clone)",
		Long: withEnv(`Materialize project workspace directories (mkdir -p only).

Creates the catalog project workspace path for each named project.
That is the directory that owns replicas (e.g. …/devel/yerk), not a replica
leaf (…/yerk/main).

With --all, ensure every project in the catalog. Bare ensure with no names
and no --all is an error (bulk mkdir is opt-in).

Never deletes. Does not create replica directories and does not run git clone.
Does not need git or --network.

See also: yerk path <project>, yerk clone <project>.`, "workspace ensure"),
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
			for _, p := range projects {
				path, err := res.WorkspacePath(p)
				if err != nil {
					return fmt.Errorf("%s: %w", p.Name, err)
				}
				if err := workspace.EnsureDir(path); err != nil {
					return err
				}
				fmt.Fprintf(streams.Out, "ensured %s -> %s\n", p.Name, path)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Ensure every project workspace in the catalog")
	return cmd
}

func newCloneCmd(streams IO) *cobra.Command {
	var (
		replicaFlag string
		network     bool
	)
	cmd := &cobra.Command{
		Use:   "clone <project> [replica]",
		Short: "Materialize a replica by cloning the project remote",
		Long: withEnv(`Materialize a replica by cloning the project's remote into the resolved path.

Ensures parent directories, refuses non-empty destinations, then runs git clone
(optionally --branch). When replica is omitted, resolves the remote default
branch via git ls-remote --symref (then catalog default_replica / "main").

Requires git on PATH.`, "clone"),
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
			p, ok := cat.Find(args[0])
			if !ok {
				return fmt.Errorf("project %q not in catalog", args[0])
			}
			if strings.TrimSpace(p.Remote) == "" {
				return fmt.Errorf("project %q has empty remote", p.Name)
			}
			git := gitcmd.New()
			res, err := project.NewResolver(cfg, git)
			if err != nil {
				return err
			}
			replica := replicaFlag
			if len(args) == 2 {
				replica = args[1]
			}
			if replica == "" {
				// Prefer network for clone so we hit the real default branch.
				replica, err = res.DefaultReplicaName(cmd.Context(), p, true)
				if err != nil {
					return err
				}
			}
			path, err := res.ReplicaPath(p, replica)
			if err != nil {
				return err
			}
			if err := workspace.EnsureParents(path); err != nil {
				return err
			}
			fmt.Fprintf(streams.Err, "cloning %s branch %s from %s -> %s\n", p.Name, replica, p.Remote, path)
			if err := git.Clone(cmd.Context(), p.Remote, path, replica); err != nil {
				return err
			}
			fmt.Fprintln(streams.Out, path)
			return nil
		},
	}
	cmd.Flags().StringVar(&replicaFlag, "replica", "", "Replica distinguisher / branch (default: remote HEAD)")
	cmd.Flags().BoolVar(&network, "network", true, "Resolve default branch via git ls-remote when replica omitted")
	_ = network // clone always uses network for default branch when omitted
	return cmd
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
				for name := range cfg.Domains {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					raw := cfg.Domains[name]
					abs, err := config.ExpandUser(strings.TrimSpace(raw))
					if err != nil {
						fmt.Fprintf(streams.Out, "  %s: %s (expand error: %v)\n", name, raw, err)
						continue
					}
					if abs != raw {
						fmt.Fprintf(streams.Out, "  %s: %s → %s\n", name, raw, abs)
					} else {
						fmt.Fprintf(streams.Out, "  %s: %s\n", name, raw)
					}
				}
			}
			fmt.Fprintln(streams.Out, "placement: relative catalog path → <domain-root>/<path>; style places <workspace>/<replica> (or project-dir)")
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
NAME, DOMAIN, PATH, DEFAULT_REPLICA, TAGS, REMOTE.
path is the catalog value (relative or absolute), not the resolved workspace.
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
			fmt.Fprintf(tw, "NAME\tDOMAIN\tPATH\tDEFAULT_REPLICA\tTAGS\tREMOTE\n")
			for _, p := range cat.Projects {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
					p.Name,
					dash(p.Domain),
					dash(p.Path),
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


// requireSubcommand makes parent commands fail on bare invoke or unknown args
// instead of printing help and exiting 0.
func requireSubcommand(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	}
	return fmt.Errorf("%s: subcommand required", cmd.CommandPath())
}

// selectProjects returns catalog rows for the given names, or the full catalog
// when all is true. Caller must already enforce names XOR all.
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
	for _, name := range names {
		p, ok := cat.Find(name)
		if !ok {
			return nil, fmt.Errorf("project %q not in catalog", name)
		}
		out = append(out, p)
	}
	return out, nil
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
