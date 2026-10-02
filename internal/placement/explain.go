package placement

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/state"
	"github.com/salotz/yerk/internal/workspace"
)

// Layer names for Contribution.Layer (stable for agents / docs).
const (
	LayerBuiltIn     = "built-in"
	LayerHostConfig  = "host-config"
	LayerDirLocal    = "dir-local"
	LayerCatalog     = "catalog"
	LayerEnv         = "env"
	LayerHostProject = "host-project"
	LayerState       = "state"
	LayerCLI         = "cli"
)

// Contribution is one ordered layer that may set placement keys (ADR 013; config resolve).
type Contribution struct {
	// Order is 1-based position in the merge stack (low → high).
	Order int `json:"order"`
	// Layer is the stable layer id (built-in, host-config, …).
	Layer string `json:"layer"`
	// Path is a filesystem path when the layer is file-backed.
	Path string `json:"path,omitempty"`
	// Key is the placement key considered (v1: workspaceStyle only).
	Key string `json:"key"`
	// Value is the value this layer supplies when Applies is true.
	Value string `json:"value,omitempty"`
	// Applies is true when this layer set Key (non-empty contribution).
	Applies bool `json:"applies"`
	// Note is optional human context (e.g. file missing).
	Note string `json:"note,omitempty"`
}

// Report is the full placement explanation for one target (config resolve v1).
type Report struct {
	// Effective is the merged placement used for path math.
	Effective Effective `json:"effective"`
	// Contributions is the ordered stack (low → high precedence among ambient;
	// state/cli follow ADR 013 rules).
	Contributions []Contribution `json:"contributions"`
	// Files is the ordered list of file paths that were considered (unique, low→high).
	Files []string `json:"files,omitempty"`
}

// Explain returns the contribution stack and effective placement (same merge as Resolve).
func Explain(in Input) (Report, error) {
	lookup := in.LookupEnv
	if lookup == nil {
		lookup = os.Getenv
	}

	var (
		rep   Report
		style = workspace.StyleWorkspaceDir
		order int
	)

	add := func(c Contribution) {
		order++
		c.Order = order
		if c.Key == "" {
			c.Key = "workspaceStyle"
		}
		rep.Contributions = append(rep.Contributions, c)
		if c.Path != "" {
			rep.Files = appendUnique(rep.Files, c.Path)
		}
		if c.Applies && c.Value != "" {
			style = c.Value
			rep.Effective.Sources = append(rep.Effective.Sources, sourceLabel(c))
		}
	}

	// 0 built-in
	add(Contribution{
		Layer:   LayerBuiltIn,
		Value:   workspace.StyleWorkspaceDir,
		Applies: true,
	})

	// 1 host config.toml
	hostPath, _ := config.FilePath()
	hostStyle := strings.TrimSpace(in.Host.Workspace.Style)
	hc := Contribution{Layer: LayerHostConfig, Path: hostPath}
	if hostPath != "" {
		if _, err := os.Stat(hostPath); err != nil {
			if os.IsNotExist(err) {
				hc.Note = "file missing (defaults in memory)"
			} else {
				hc.Note = err.Error()
			}
		}
	}
	if hostStyle != "" {
		hc.Value = hostStyle
		hc.Applies = true
	} else {
		hc.Note = joinNote(hc.Note, "no workspace.style")
	}
	add(hc)

	// 2 dir-local (far → near already ordered by collectDirLocalStyles)
	anchor := strings.TrimSpace(in.Anchor)
	if anchor == "" {
		if cwd, err := os.Getwd(); err == nil {
			anchor = cwd
		}
	}
	locals, err := collectDirLocalStyles(anchor)
	if err != nil {
		return Report{}, err
	}
	for _, layer := range locals {
		c := Contribution{Layer: LayerDirLocal, Path: layer.Path}
		if layer.Style != "" {
			c.Value = layer.Style
			c.Applies = true
		} else {
			c.Note = "no workspace.style"
		}
		add(c)
	}

	// 3 catalog row
	catPath, _ := config.CatalogPath()
	catStyle := strings.TrimSpace(in.Project.WorkspaceStyle)
	cc := Contribution{Layer: LayerCatalog, Path: catPath}
	if in.Project.Name != "" {
		cc.Note = "project " + in.Project.ID()
	}
	if catStyle != "" {
		cc.Value = catStyle
		cc.Applies = true
	} else {
		cc.Note = joinNote(cc.Note, "no workspace_style on row")
	}
	add(cc)

	// 4 env
	envStyle := strings.TrimSpace(lookup("YERK__WORKSPACE_STYLE"))
	ec := Contribution{Layer: LayerEnv, Key: "workspaceStyle"}
	if envStyle != "" {
		ec.Value = envStyle
		ec.Applies = true
		ec.Note = "YERK__WORKSPACE_STYLE"
	} else {
		ec.Note = "YERK__WORKSPACE_STYLE unset"
	}
	add(ec)

	// 5 host [[projects]] row
	hpNote := "no host [[projects]] row"
	var hpStyle string
	if hp, ok := in.Host.FindHostProject(in.Project.Domain, in.Project.Name); ok {
		hpNote = "host [[projects]] " + hp.ID()
		hpStyle = strings.TrimSpace(hp.WorkspaceStyle)
	}
	hpc := Contribution{Layer: LayerHostProject, Path: hostPath, Note: hpNote}
	if hpStyle != "" {
		hpc.Value = hpStyle
		hpc.Applies = true
	} else if strings.HasPrefix(hpNote, "host") {
		hpc.Note = joinNote(hpNote, "no workspace_style")
	}
	add(hpc)

	ambientStyle := style
	ambientSource := lastSource(rep.Effective.Sources)

	// 6 state — only attach Path / files entry when the binding file exists.
	if !in.SkipState {
		sc := Contribution{Layer: LayerState, Key: "workspaceStyle"}
		st, ok, err := state.LoadProject(in.Project.Domain, in.Project.Name)
		if err != nil {
			return Report{}, err
		}
		stPath, pathErr := state.ProjectFile(in.Project.Domain, in.Project.Name)
		if ok {
			if pathErr == nil {
				sc.Path = stPath
			}
			bound := strings.TrimSpace(st.WorkspaceStyle)
			if bound != "" {
				rep.Effective.Bound = true
				if bound != ambientStyle {
					rep.Effective.Warnings = append(rep.Effective.Warnings, fmt.Sprintf(
						"placement: bound workspace style %q differs from ambient %q (%s); using bound style",
						bound, ambientStyle, ambientSource))
				}
				sc.Value = bound
				sc.Applies = true
				style = bound
				rep.Effective.Sources = append(rep.Effective.Sources, sourceLabel(sc))
			} else {
				sc.Note = "state present but workspaceStyle empty"
			}
		} else {
			sc.Note = "no binding"
			if pathErr != nil || stPath == "" {
				sc.Note = "no project identity for state path"
			}
			// Do not set Path or list a missing state.json (operator request).
		}
		order++
		sc.Order = order
		rep.Contributions = append(rep.Contributions, sc)
		if sc.Path != "" {
			rep.Files = appendUnique(rep.Files, sc.Path)
		}
	}

	// 7 CLI
	cliStyle := strings.TrimSpace(in.CLIStyle)
	cli := Contribution{Layer: LayerCLI, Key: "workspaceStyle"}
	if cliStyle != "" {
		if rep.Effective.Bound && cliStyle != style {
			return Report{}, fmt.Errorf(
				"placement: --workspace-style=%q contradicts bound style %q (rebind not implemented; edit state or omit the flag)",
				cliStyle, style)
		}
		cli.Value = cliStyle
		cli.Applies = true
		style = cliStyle
		rep.Effective.Sources = append(rep.Effective.Sources, sourceLabel(cli))
	} else {
		cli.Note = "not set"
	}
	order++
	cli.Order = order
	rep.Contributions = append(rep.Contributions, cli)

	if err := workspace.ValidateStyle(style); err != nil {
		return Report{}, err
	}
	rep.Effective.Style = style
	return rep, nil
}

func sourceLabel(c Contribution) string {
	switch c.Layer {
	case LayerDirLocal:
		return c.Layer + ":" + c.Path + ":" + c.Value
	case LayerEnv:
		return c.Layer + ":" + c.Value
	case LayerHostConfig, LayerCatalog, LayerHostProject:
		return c.Layer + ":" + c.Value
	case LayerState, LayerCLI, LayerBuiltIn:
		return c.Layer + ":" + c.Value
	default:
		return c.Layer + ":" + c.Value
	}
}

func appendUnique(in []string, s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return in
	}
	for _, x := range in {
		if x == s {
			return in
		}
	}
	return append(in, s)
}

func joinNote(a, b string) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return a + "; " + b
	}
}

// AbsPath is a tiny helper for tests/docs (re-export clean abs).
func AbsPath(p string) (string, error) {
	return filepath.Abs(p)
}
