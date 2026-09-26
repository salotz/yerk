// Package cli is the yerk command-line interface.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/salotz/yerk/internal/config"
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
		Long:          longHelp,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.SetIn(streams.In)
	root.SetOut(streams.Out)
	root.SetErr(streams.Err)

	root.AddCommand(newVersionCmd(streams))
	root.AddCommand(newStatusCmd(streams))
	root.AddCommand(newConfigCmd(streams))

	// MVP placeholders: implement next.
	root.AddCommand(stubCmd(streams, "register", "Register a project in the host catalog"))
	root.AddCommand(stubCmd(streams, "clone", "Clone a registered project replica into the workspace"))
	root.AddCommand(stubCmd(streams, "pull", "Pull one or many registered projects"))
	root.AddCommand(stubCmd(streams, "push", "Push one or many registered projects"))

	return root
}

// Execute runs the root command with process args.
func Execute(ctx context.Context, streams IO, args []string) error {
	root := NewRoot(streams)
	root.SetArgs(args)
	return root.ExecuteContext(ctx)
}

const longHelp = `yerk is a host-level multi-project manager.

It registers software projects, checks out replicas into a workspace layout,
reports status across many repos, and (later) stages host-local configuration
into those checkouts. Vocabulary aligns with PRJX (project, replica, workspace).

Config:  $XDG_CONFIG_HOME/yerk/config.toml  (override with YERK__CONFIG)
Env:     YERK__… for tool knobs; discover .prjx-root / PRJX__… for PRJX concerns
`

func newVersionCmd(streams IO) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version identity",
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
		asJSON    bool
	)
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report registered projects (MVP: catalog only, no git probe yet)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			projects := cfg.Projects
			if tagFilter != "" {
				filtered := projects[:0]
				for _, p := range projects {
					if hasTag(p.Tags, tagFilter) {
						filtered = append(filtered, p)
					}
				}
				projects = filtered
			}

			if len(projects) == 0 {
				fmt.Fprintln(streams.Out, "No projects registered.")
				cfgPath, _ := config.FilePath()
				fmt.Fprintf(streams.Out, "Config: %s\n", cfgPath)
				fmt.Fprintln(streams.Out, "Add [[projects]] entries, then re-run: yerk status")
				return nil
			}

			layout, err := workspace.NewLayout(cfg.Workspace)
			if err != nil {
				return err
			}

			if asJSON {
				// Keep MVP free of a JSON dependency in the status path:
				// emit a simple line-oriented form until a real reporter exists.
				fmt.Fprintln(streams.Err, "note: --json is reserved; printing text table for now")
			}

			fmt.Fprintf(streams.Out, "workspace.style=%s", layout.Style)
			if layout.Root != "" {
				fmt.Fprintf(streams.Out, " root=%s", layout.Root)
			}
			fmt.Fprintln(streams.Out)
			fmt.Fprintln(streams.Out)

			fmt.Fprintf(streams.Out, "%-20s %-40s %s\n", "NAME", "REMOTE", "TAGS")
			for _, p := range projects {
				fmt.Fprintf(streams.Out, "%-20s %-40s %s\n",
					p.Name, truncate(p.Remote, 40), joinTags(p.Tags))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&tagFilter, "tag", "", "Only projects with this tag")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Reserved: machine-readable output")
	return cmd
}

func newConfigCmd(streams IO) *cobra.Command {
	root := &cobra.Command{
		Use:   "config",
		Short: "Show config paths and example document",
	}
	root.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the resolved config file path",
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
		Use:   "example",
		Short: "Print an example config.toml",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprint(streams.Out, config.ExampleTOML)
			return err
		},
	})
	root.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Load and summarize the active config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			path, _ := config.FilePath()
			fmt.Fprintf(streams.Out, "file: %s\n", path)
			fmt.Fprintf(streams.Out, "workspace.style: %s\n", cfg.Workspace.Style)
			if cfg.Workspace.Root != "" {
				fmt.Fprintf(streams.Out, "workspace.root: %s\n", cfg.Workspace.Root)
			} else {
				fmt.Fprintln(streams.Out, "workspace.root: (unset)")
			}
			fmt.Fprintf(streams.Out, "projects: %d\n", len(cfg.Projects))
			return nil
		},
	})
	return root
}

func stubCmd(streams IO, use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short + " (not implemented yet)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New(use + ": not implemented yet (MVP scaffold)")
		},
	}
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func joinTags(tags []string) string {
	if len(tags) == 0 {
		return "-"
	}
	out := tags[0]
	for i := 1; i < len(tags); i++ {
		out += "," + tags[i]
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
