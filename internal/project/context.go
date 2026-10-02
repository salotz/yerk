package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/state"
	"github.com/salotz/yerk/internal/version"
)

// ToolContext builds a tool-wide agent dump (ADR 017). No catalog match required.
func ToolContext() (api.ToolContext, error) {
	out := api.NewToolContext()
	out.Version = version.String()
	out.Vocabulary = []api.ContextTerm{
		{Term: "catalog", Def: "Portable registry of projects (identity, remote, tags) in catalog.toml"},
		{Term: "project", Def: "Catalog row: domain/name; workspace owns replicas"},
		{Term: "domain", Def: "Id namespace (ADR 012); optional host [domains] root uses the same string"},
		{Term: "replica", Def: "One checkout under a project workspace (distinguisher often = branch)"},
		{Term: "workspace", Def: "Host directory that owns a project's replicas (not itself a git checkout)"},
		{Term: "materialize", Def: "Bootstrap replica from catalog remote (git clone); already-present ok"},
		{Term: "replica create", Def: "Session spin-out: worktree from main or clone method (ADR 016)"},
		{Term: "presence", Def: "missing | present | invalid for an expected path"},
		{Term: "change", Def: "Git probe flags (dirty, untracked, ahead/behind, …) when present"},
		{Term: "yerk://", Def: "Canonical URI currency: yerk://domain/name[/replica]"},
	}
	out.Commands = []api.ContextCommand{
		{Name: "status", Role: "presence + change; one project lists live replicas + overall"},
		{Name: "path / resolve", Role: "print workspace or replica path"},
		{Name: "get / lookup", Role: "one-resource info by id or path (--output json)"},
		{Name: "config resolve", Role: "placement contribution stack for a project"},
		{Name: "state update", Role: "rebind host project state from ambient"},
		{Name: "workspace ensure", Role: "mkdir project workspace only"},
		{Name: "materialize", Role: "clone replica(s) from remote"},
		{Name: "replica create", Role: "worktree|clone session replica"},
		{Name: "context / context dir", Role: "agent dumps (this command)"},
		{Name: "catalog / envvars / version", Role: "show catalog, live env, version"},
	}
	cfgDir, err := config.Dir()
	if err != nil {
		return out, err
	}
	cfgFile, err := config.FilePath()
	if err != nil {
		return out, err
	}
	catFile, err := config.CatalogPath()
	if err != nil {
		return out, err
	}
	stDir, err := state.Dir()
	if err != nil {
		return out, err
	}
	out.Paths = api.ToolContextPaths{
		ConfigDir:   cfgDir,
		ConfigFile:  cfgFile,
		CatalogFile: catFile,
		StateDir:    stDir,
	}
	out.EnvPrimary = []string{
		"YERK__CONFIG_DIR",
		"YERK__CONFIG",
		"YERK__CATALOG",
		"YERK__STATE_DIR",
		"YERK__WORKSPACE_STYLE",
	}
	out.HowTo = []string{
		"Register projects by editing catalog.toml (domain + name + remote)",
		"Host paths: config.toml [domains] and optional [[projects]] (ADR 014)",
		"Bootstrap: yerk workspace ensure <id>; yerk materialize <id>",
		"Session: yerk replica create <id> <name> (main must exist for worktree)",
		"Where am I: yerk context dir; detail: yerk get / yerk lookup --output json",
		"Why style X: yerk config resolve <project-id>",
		"Full env docs: yerk help envvars; live: yerk envvars",
	}
	out.Notes = []string{
		"JSON keys are best-effort stable under apiVersion yerk/v1 (ADR 017); avoid casual renames",
		"Identity is not a filesystem path; use yerk:// or bare domain/name",
		"Do not scrape status tables when get/lookup/context JSON is available",
	}
	return out, nil
}

// DirContextOptions controls directory context collection.
type DirContextOptions struct {
	// Git enables change probes when collecting project status overall (default false for cheap dumps).
	Git bool
	// Network allows ls-remote for default replica name when Git is on.
	Network bool
}

// DirContext builds a directory-scoped agent dump for path (empty → cwd).
// Composes lookup + info + compact placement + optional presence status.
func (r Resolver) DirContext(ctx context.Context, projects []config.Project, path string, opts DirContextOptions) (api.DirContext, error) {
	out := api.NewDirContext()

	path = strings.TrimSpace(path)
	if path == "" {
		wd, err := os.Getwd()
		if err != nil {
			return out, fmt.Errorf("context dir: cwd: %w", err)
		}
		path = wd
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return out, fmt.Errorf("context dir: abs path: %w", err)
	}
	out.Path = abs

	result, err := r.LookupPath(projects, abs, LookupAny)
	if err != nil {
		return out, err
	}

	var p config.Project
	switch {
	case result.Replica != nil:
		out.Matched = "replica"
		rep := *result.Replica
		out.Replica = &rep
		// Project info for the same row (catalog).
		cp, ok := findProject(projects, rep.Domain, rep.Project)
		if !ok {
			return out, fmt.Errorf("context dir: catalog missing %s/%s", rep.Domain, rep.Project)
		}
		p = cp
		pi, err := r.ProjectInfo(p)
		if err != nil {
			return out, err
		}
		out.Project = &pi
	case result.Project != nil:
		out.Matched = "project"
		pi := *result.Project
		out.Project = &pi
		cp, ok := findProject(projects, pi.Domain, pi.Name)
		if !ok {
			return out, fmt.Errorf("context dir: catalog missing %s/%s", pi.Domain, pi.Name)
		}
		p = cp
	default:
		return out, fmt.Errorf("context dir: empty lookup for %s", abs)
	}

	// Compact placement from ConfigResolve (same merge; no second policy).
	cr, err := r.ConfigResolve(p)
	if err != nil {
		return out, err
	}
	out.EffectiveStyle = cr.EffectiveStyle
	out.Bound = cr.Bound
	out.Warnings = append([]string(nil), cr.Warnings...)
	if out.Project != nil && out.Project.Placement != nil {
		out.Placement = out.Project.Placement
	} else if out.Replica != nil && out.Replica.Placement != nil {
		out.Placement = out.Replica.Placement
	}

	// Short status: presence-oriented by default (cheap).
	stOpts := StatusOptions{Git: opts.Git, Network: opts.Network || opts.Git}
	st, err := r.ProjectStatus(ctx, p, stOpts)
	if err != nil {
		return out, err
	}
	if st.Overall != nil {
		ov := *st.Overall
		out.StatusOverall = &ov
	}
	for _, row := range st.Replicas {
		if row.Presence == api.PresencePresent {
			out.LiveReplicas = append(out.LiveReplicas, row.Replica)
		}
	}

	out.Notes = []string{
		"Use yerk get / yerk lookup --output json for single-resource detail",
		"Use yerk config resolve " + p.ID() + " for full contribution stack",
		"Use yerk status " + p.ID() + " for full per-replica change table",
	}
	return out, nil
}

func findProject(projects []config.Project, domain, name string) (config.Project, bool) {
	domain = strings.TrimSpace(domain)
	name = strings.TrimSpace(name)
	for _, p := range projects {
		if strings.TrimSpace(p.Domain) == domain && strings.TrimSpace(p.Name) == name {
			return p, true
		}
	}
	return config.Project{}, false
}
