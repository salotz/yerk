// Package placement merges layered workspace policy (ADR 013).
//
// Ambient stack (low → high): built-in → host config → dir-local (near wins) →
// catalog row → YERK__WORKSPACE_STYLE.
// Bound host project state overrides ambient with warnings.
// Explicit CLI overrides ambient; contradicts bound state → error.
package placement

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/state"
	"github.com/salotz/yerk/internal/workspace"
)

// LocalConfigRel is the relative path under each directory for dir-local yerk config.
const LocalConfigRel = ".local/yerk/config.toml"

// Input is one resolve request for effective placement.
type Input struct {
	// Host is tool config (workspace style).
	Host config.Config
	// Project is the catalog row (optional placement fields).
	Project config.Project
	// Anchor is the start path for dir-local walk. Empty → process cwd.
	// Prefer the resolved project workspace path when known.
	Anchor string
	// CLIStyle is an explicit --workspace-style (empty = unset).
	CLIStyle string
	// SkipState ignores host project state (init snapshot / tests).
	SkipState bool
	// LookupEnv overrides env lookup (tests). nil → os.Getenv.
	LookupEnv func(key string) string
}

// Effective is the merged placement policy for path math and reporting.
type Effective struct {
	// Style is the workspace style name used for layout.
	Style string
	// Bound is true when host project state supplied the winning style.
	Bound bool
	// Warnings are ambient drift messages (state kept).
	Warnings []string
	// Sources names contributing layers for debug (best-effort).
	Sources []string
}

// Resolve merges layers into an Effective style (ADR 013).
func Resolve(in Input) (Effective, error) {
	lookup := in.LookupEnv
	if lookup == nil {
		lookup = os.Getenv
	}

	var eff Effective
	style := workspace.StyleWorkspaceDir
	eff.Sources = append(eff.Sources, "built-in:"+style)

	hostStyle := strings.TrimSpace(in.Host.Workspace.Style)
	if hostStyle != "" {
		style = hostStyle
		eff.Sources = append(eff.Sources, "host-config:"+style)
	}

	anchor := strings.TrimSpace(in.Anchor)
	if anchor == "" {
		if cwd, err := os.Getwd(); err == nil {
			anchor = cwd
		}
	}
	locals, err := collectDirLocalStyles(anchor)
	if err != nil {
		return Effective{}, err
	}
	for _, layer := range locals {
		if layer.Style == "" {
			continue
		}
		style = layer.Style
		eff.Sources = append(eff.Sources, "dir-local:"+layer.Path+":"+style)
	}

	catStyle := strings.TrimSpace(in.Project.WorkspaceStyle)
	if catStyle != "" {
		style = catStyle
		eff.Sources = append(eff.Sources, "catalog:"+style)
	}

	envStyle := strings.TrimSpace(lookup("YERK__WORKSPACE_STYLE"))
	if envStyle != "" {
		style = envStyle
		eff.Sources = append(eff.Sources, "env:"+style)
	}

	// Host [[projects]] row override (config.toml): more specific than process env.
	if hp, ok := in.Host.FindHostProject(in.Project.Domain, in.Project.Name); ok {
		if hs := strings.TrimSpace(hp.WorkspaceStyle); hs != "" {
			style = hs
			eff.Sources = append(eff.Sources, "host-project:"+style)
		}
	}

	ambientStyle := style
	ambientSource := lastSource(eff.Sources)

	if !in.SkipState {
		st, ok, err := state.LoadProject(in.Project.Domain, in.Project.Name)
		if err != nil {
			return Effective{}, err
		}
		if ok {
			bound := strings.TrimSpace(st.WorkspaceStyle)
			if bound != "" {
				eff.Bound = true
				if bound != ambientStyle {
					eff.Warnings = append(eff.Warnings, fmt.Sprintf(
						"placement: bound workspace style %q differs from ambient %q (%s); using bound style",
						bound, ambientStyle, ambientSource))
				}
				style = bound
				eff.Sources = append(eff.Sources, "state:"+style)
			}
		}
	}

	cliStyle := strings.TrimSpace(in.CLIStyle)
	if cliStyle != "" {
		if eff.Bound && cliStyle != style {
			return Effective{}, fmt.Errorf(
				"placement: --workspace-style=%q contradicts bound style %q (rebind not implemented; edit state or omit the flag)",
				cliStyle, style)
		}
		style = cliStyle
		eff.Sources = append(eff.Sources, "cli:"+style)
	}

	if err := workspace.ValidateStyle(style); err != nil {
		return Effective{}, err
	}
	eff.Style = style
	return eff, nil
}

type dirLocalLayer struct {
	Path  string
	Style string
}

// collectDirLocalStyles walks from anchor up to home (inclusive).
// Returns layers farthest→nearest so overwrite yields near-wins.
// When anchor is outside home (e.g. tests under /tmp), walk stops at FS root.
func collectDirLocalStyles(anchor string) ([]dirLocalLayer, error) {
	if strings.TrimSpace(anchor) == "" {
		return nil, nil
	}
	start, err := filepath.Abs(anchor)
	if err != nil {
		return nil, fmt.Errorf("dir-local anchor: %w", err)
	}
	if fi, err := os.Stat(start); err != nil || !fi.IsDir() {
		start = filepath.Dir(start)
	}
	start = filepath.Clean(start)

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("dir-local home: %w", err)
	}
	home = filepath.Clean(home)
	underHome := start == home || strings.HasPrefix(start, home+string(os.PathSeparator))

	var nearFirst []dirLocalLayer
	cur := start
	for {
		cfgPath := filepath.Join(cur, LocalConfigRel)
		if st, err := loadLocalFile(cfgPath); err != nil {
			return nil, err
		} else if st != "" {
			nearFirst = append(nearFirst, dirLocalLayer{Path: cfgPath, Style: st})
		}

		if underHome {
			if cur == home {
				break
			}
			parent := filepath.Dir(cur)
			if parent == cur {
				break
			}
			// Do not walk above home.
			if parent != home && !strings.HasPrefix(parent, home+string(os.PathSeparator)) && parent != home {
				// parent outside home — stop after processing current only
				break
			}
			cur = parent
			continue
		}

		// Outside home: walk to filesystem root (test temp trees).
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}

	out := make([]dirLocalLayer, 0, len(nearFirst))
	for i := len(nearFirst) - 1; i >= 0; i-- {
		out = append(out, nearFirst[i])
	}
	return out, nil
}

func loadLocalFile(path string) (style string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read dir-local %s: %w", path, err)
	}
	var raw struct {
		Workspace struct {
			Style string `toml:"style"`
		} `toml:"workspace"`
	}
	if err := toml.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("parse dir-local %s: %w", path, err)
	}
	return strings.TrimSpace(raw.Workspace.Style), nil
}

func lastSource(sources []string) string {
	if len(sources) == 0 {
		return "built-in"
	}
	return sources[len(sources)-1]
}

// InitStyle is the style to record at initialize: ambient merge without
// existing state (SkipState). CLI still applies if set.
func InitStyle(in Input) (string, error) {
	in.SkipState = true
	eff, err := Resolve(in)
	if err != nil {
		return "", err
	}
	return eff.Style, nil
}
