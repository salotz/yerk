// Package project joins catalog, layout, presence, and git for common ops.
package project

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/gitcmd"
	"github.com/salotz/yerk/internal/id"
	"github.com/salotz/yerk/internal/placement"
	"github.com/salotz/yerk/internal/presence"
	"github.com/salotz/yerk/internal/state"
	"github.com/salotz/yerk/internal/workspace"
)

// FallbackReplica is used when catalog and remote default branch are unavailable.
const FallbackReplica = "main"

// Resolver resolves paths and default replica names using effective placement.
type Resolver struct {
	Cfg    config.Config
	Layout workspace.Layout // host-default layout; per-project layout via layoutFor
	Git    gitcmd.Runner
	// Warn is where placement / deprecation warnings go (default stderr).
	Warn io.Writer
	// CLIStyle is optional explicit --workspace-style for mutate commands.
	CLIStyle string
}

// NewResolver builds a resolver from tool config.
func NewResolver(cfg config.Config, git gitcmd.Runner) (Resolver, error) {
	layout, err := workspace.NewLayout(cfg)
	if err != nil {
		return Resolver{}, err
	}
	if git == nil {
		git = gitcmd.New()
	}
	return Resolver{Cfg: cfg, Layout: layout, Git: git, Warn: os.Stderr}, nil
}

func (r Resolver) warnf(format string, args ...any) {
	w := r.Warn
	if w == nil {
		w = os.Stderr
	}
	fmt.Fprintf(w, format, args...)
}

// layoutFor returns a Layout using effective placement for project p.
func (r Resolver) layoutFor(p config.Project) (workspace.Layout, placement.Effective, error) {
	in := placement.Input{
		Host:     r.Cfg,
		Project:  p,
		CLIStyle: r.CLIStyle,
	}
	// Anchor dir-local walk at the project workspace when path math allows
	// (host style does not affect workspace root for shipped styles).
	boot, err := workspace.NewLayoutStyle(r.Layout.Style, r.Cfg, r.warnf)
	if err != nil {
		return workspace.Layout{}, placement.Effective{}, err
	}
	if ws, err := boot.ProjectDir(p); err == nil {
		in.Anchor = ws
	}
	eff, err := placement.Resolve(in)
	if err != nil {
		return workspace.Layout{}, placement.Effective{}, err
	}
	for _, w := range eff.Warnings {
		r.warnf("warning: %s\n", w)
	}
	layout, err := workspace.NewLayoutStyle(eff.Style, r.Cfg, r.warnf)
	if err != nil {
		return workspace.Layout{}, placement.Effective{}, err
	}
	return layout, eff, nil
}

// DefaultReplicaName chooses the replica distinguisher for a project.
// network=true allows ls-remote; otherwise uses override or FallbackReplica.
func (r Resolver) DefaultReplicaName(ctx context.Context, p config.Project, network bool) (string, error) {
	if v := strings.TrimSpace(p.DefaultReplica); v != "" {
		return v, nil
	}
	if network && strings.TrimSpace(p.Remote) != "" && r.Git != nil {
		branch, err := r.Git.DefaultBranch(ctx, p.Remote)
		if err == nil && branch != "" {
			return branch, nil
		}
		// fall through on error
	}
	return FallbackReplica, nil
}

// WorkspacePath returns the absolute project workspace directory (owns replicas).
func (r Resolver) WorkspacePath(p config.Project) (string, error) {
	layout, _, err := r.layoutFor(p)
	if err != nil {
		return "", err
	}
	return layout.ProjectDir(p)
}

// ReplicaPath returns the on-disk path for project + replica distinguisher.
func (r Resolver) ReplicaPath(p config.Project, replica string) (string, error) {
	layout, _, err := r.layoutFor(p)
	if err != nil {
		return "", err
	}
	return layout.ReplicaDir(p, replica)
}

// EffectivePlacement returns the merged placement for reporting/tests.
func (r Resolver) EffectivePlacement(p config.Project) (placement.Effective, error) {
	_, eff, err := r.layoutFor(p)
	return eff, err
}

// Replica method names (ADR 016).
const (
	MethodWorktree = "worktree"
	MethodClone    = "clone"
)

// DefaultReplicaMethod is used when catalog and CLI leave method unset.
const DefaultReplicaMethod = MethodWorktree

// CreateReplicaOptions controls yerk replica create (ADR 016).
type CreateReplicaOptions struct {
	// Method is worktree|clone. Empty → catalog replica_method → DefaultReplicaMethod.
	Method string
	// CLIStyle is explicit --workspace-style (also set on Resolver.CLIStyle).
	CLIStyle string
}

// CreateReplicaResult is the outcome of a successful create.
type CreateReplicaResult struct {
	Path   string
	Method string
	Main   string // main replica path when method is worktree; empty for clone
}

// ResolveReplicaMethod picks worktree|clone (CLI → catalog → default).
func ResolveReplicaMethod(cliMethod, catalogMethod string) (string, error) {
	m := strings.TrimSpace(cliMethod)
	if m == "" {
		m = strings.TrimSpace(catalogMethod)
	}
	if m == "" {
		m = DefaultReplicaMethod
	}
	switch m {
	case MethodWorktree, MethodClone:
		return m, nil
	default:
		return "", fmt.Errorf("unknown replica method %q (want %s or %s)", m, MethodWorktree, MethodClone)
	}
}

// CreateReplica spins out a new replica under the project layout (ADR 016).
// Destination must not already be present; worktree requires main present.
func (r Resolver) CreateReplica(ctx context.Context, p config.Project, replica string, opts CreateReplicaOptions) (CreateReplicaResult, error) {
	replica = strings.TrimSpace(replica)
	if replica == "" {
		return CreateReplicaResult{}, fmt.Errorf("replica distinguisher required")
	}
	if opts.CLIStyle != "" {
		r.CLIStyle = opts.CLIStyle
	}

	method, err := ResolveReplicaMethod(opts.Method, p.ReplicaMethod)
	if err != nil {
		return CreateReplicaResult{}, err
	}

	dest, err := r.ReplicaPath(p, replica)
	if err != nil {
		return CreateReplicaResult{}, err
	}

	switch presence.Classify(dest) {
	case presence.Present:
		return CreateReplicaResult{}, fmt.Errorf("replica %q already present at %s (refuse; no --if-absent yet)", replica, dest)
	case presence.Invalid:
		return CreateReplicaResult{}, fmt.Errorf("replica path exists but is not a usable git checkout: %s", dest)
	}

	// Missing path may still be an empty or non-empty non-git dir handled by gitcmd.
	if err := workspace.EnsureParents(dest); err != nil {
		return CreateReplicaResult{}, err
	}

	out := CreateReplicaResult{Path: dest, Method: method}

	switch method {
	case MethodWorktree:
		mainName := strings.TrimSpace(p.DefaultReplica)
		if mainName == "" {
			mainName = FallbackReplica
		}
		mainPath, err := r.ReplicaPath(p, mainName)
		if err != nil {
			return CreateReplicaResult{}, err
		}
		out.Main = mainPath
		switch presence.Classify(mainPath) {
		case presence.Present:
			// ok
		case presence.Missing:
			return CreateReplicaResult{}, fmt.Errorf(
				"main replica %q is missing at %s; materialize it first (yerk materialize %s %s) before worktree create",
				mainName, mainPath, p.ID(), mainName)
		default:
			return CreateReplicaResult{}, fmt.Errorf(
				"main replica %q is not a usable git checkout at %s; fix or re-materialize before worktree create",
				mainName, mainPath)
		}
		if replica == mainName {
			return CreateReplicaResult{}, fmt.Errorf(
				"cannot create worktree for main replica name %q; use yerk materialize for the hub checkout",
				replica)
		}
		if r.Git == nil {
			return CreateReplicaResult{}, fmt.Errorf("git runner required")
		}
		if err := r.Git.WorktreeAdd(ctx, mainPath, dest, replica); err != nil {
			return CreateReplicaResult{}, err
		}
	case MethodClone:
		if strings.TrimSpace(p.Remote) == "" {
			return CreateReplicaResult{}, fmt.Errorf("project %q has empty remote", p.ID())
		}
		if r.Git == nil {
			return CreateReplicaResult{}, fmt.Errorf("git runner required")
		}
		// Clone remote default HEAD (session names rarely exist on the remote),
		// then ensure local branch matches the replica distinguisher (ADR 016).
		if err := r.Git.Clone(ctx, p.Remote, dest, ""); err != nil {
			return CreateReplicaResult{}, err
		}
		if err := r.Git.EnsureBranch(ctx, dest, replica); err != nil {
			return CreateReplicaResult{}, err
		}
	default:
		return CreateReplicaResult{}, fmt.Errorf("unknown replica method %q", method)
	}

	if err := r.BindOnInit(p); err != nil {
		return CreateReplicaResult{}, fmt.Errorf("bind state: %w", err)
	}
	return out, nil
}

// BindOnInit writes host project state on first ensure/materialize (ADR 013).
// No-op if already bound. Uses ambient init style (state skipped while computing).
func (r Resolver) BindOnInit(p config.Project) error {
	in := placement.Input{
		Host:     r.Cfg,
		Project:  p,
		CLIStyle: r.CLIStyle,
	}
	if ws, err := r.WorkspacePath(p); err == nil {
		in.Anchor = ws
	}
	style, err := placement.InitStyle(in)
	if err != nil {
		return err
	}
	written, err := state.BindStyle(p.Domain, p.Name, style)
	if err != nil {
		return err
	}
	if written {
		r.warnf("bound %s workspace style %q -> state\n", p.ID(), style)
	}
	return nil
}

// UpdateStateResult is the outcome of UpdateState for one project.
type UpdateStateResult struct {
	Project  string
	Style    string
	Previous string
	Created  bool
	Changed  bool
}

// UpdateState rebinds host project state from current ambient placement
// (and optional CLIStyle). Unlike BindOnInit, overwrites an existing binding.
// Use yerk state update for explicit refresh after ambient config changes.
func (r Resolver) UpdateState(p config.Project) (UpdateStateResult, error) {
	in := placement.Input{
		Host:     r.Cfg,
		Project:  p,
		CLIStyle: r.CLIStyle,
	}
	if ws, err := r.WorkspacePath(p); err == nil {
		in.Anchor = ws
	}
	// Skip existing state so ambient+CLI is the source of the new binding.
	style, err := placement.InitStyle(in)
	if err != nil {
		return UpdateStateResult{}, err
	}
	ur, err := state.UpdateStyle(p.Domain, p.Name, style)
	if err != nil {
		return UpdateStateResult{}, err
	}
	return UpdateStateResult{
		Project:  p.ID(),
		Style:    ur.Style,
		Previous: ur.Previous,
		Created:  ur.Created,
		Changed:  ur.Changed,
	}, nil
}

// StatusOptions controls status collection.
type StatusOptions struct {
	// Git enables change probes when a replica is present (default for CLI).
	Git bool
	// Network resolves default branch via ls-remote when default_replica is unset.
	Network bool
}

// ProjectStatuses builds ProjectStatus resources for the given projects
// (caller selects by name/tag/all). Serial collection (P7 may parallelize).
func (r Resolver) ProjectStatuses(ctx context.Context, projects []config.Project, opts StatusOptions) ([]api.ProjectStatus, error) {
	out := make([]api.ProjectStatus, 0, len(projects))
	for _, p := range projects {
		st, err := r.ProjectStatus(ctx, p, opts)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, nil
}

// Status builds default-replica ReplicaStatus resources for the given projects.
// Prefer ProjectStatuses for the default human project list (P6).
func (r Resolver) Status(ctx context.Context, projects []config.Project, opts StatusOptions) ([]api.ReplicaStatus, error) {
	rows := make([]api.ReplicaStatus, 0, len(projects))
	for _, p := range projects {
		row, err := r.replicaStatus(ctx, p, "", opts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// ReplicaStatus builds one ReplicaStatus for project + replica distinguisher.
// Empty replica uses DefaultReplicaName (honoring opts.Network).
func (r Resolver) ReplicaStatus(ctx context.Context, p config.Project, replica string, opts StatusOptions) (api.ReplicaStatus, error) {
	return r.replicaStatus(ctx, p, replica, opts)
}

// ProjectStatus builds a ProjectStatus with workspace path, live replicas,
// default-replica summary, and an overall rollup (presence + change).
func (r Resolver) ProjectStatus(ctx context.Context, p config.Project, opts StatusOptions) (api.ProjectStatus, error) {
	ws, err := r.WorkspacePath(p)
	if err != nil {
		return api.ProjectStatus{}, fmt.Errorf("%s: %w", p.Name, err)
	}
	layout, _, err := r.layoutFor(p)
	if err != nil {
		return api.ProjectStatus{}, fmt.Errorf("%s: %w", p.Name, err)
	}

	defaultName, err := r.DefaultReplicaName(ctx, p, opts.Network)
	if err != nil {
		return api.ProjectStatus{}, fmt.Errorf("%s: %w", p.Name, err)
	}

	live, err := layout.ListLiveReplicas(p)
	if err != nil {
		return api.ProjectStatus{}, fmt.Errorf("%s: %w", p.Name, err)
	}

	// Collect unique replica names: all live + default (even if missing).
	seen := make(map[string]struct{}, len(live)+1)
	var names []string
	add := func(n string) {
		n = strings.TrimSpace(n)
		if n == "" {
			return
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		names = append(names, n)
	}
	for _, n := range live {
		add(n)
	}
	add(defaultName)
	sort.Strings(names)

	replicas := make([]api.ReplicaStatus, 0, len(names))
	for _, name := range names {
		row, err := r.replicaStatus(ctx, p, name, opts)
		if err != nil {
			return api.ProjectStatus{}, err
		}
		replicas = append(replicas, row)
	}

	out := api.NewProjectStatus()
	out.URI = id.ProjectURI(p.Domain, p.Name)
	out.Name = p.Name
	out.Domain = p.Domain
	out.WorkspacePath = ws
	out.WorkspacePresence = classifyWorkspaceDir(ws)
	out.Tags = append([]string(nil), p.Tags...)
	out.Remote = p.Remote
	out.Replicas = replicas
	overall := rollupProjectOverall(replicas, opts.Git)
	out.Overall = &overall

	// DefaultReplica summary: prefer the matching row; else synthesize empty.
	for i := range replicas {
		if replicas[i].Replica == defaultName {
			sum := replicas[i].Summary()
			out.DefaultReplica = &sum
			break
		}
	}
	if out.DefaultReplica == nil {
		// Should not happen (default always added), but keep shape stable.
		rep, err := r.replicaStatus(ctx, p, defaultName, opts)
		if err == nil {
			sum := rep.Summary()
			out.DefaultReplica = &sum
		}
	}
	return out, nil
}

// rollupProjectOverall merges replica rows into one presence + change bag.
// Presence prefers invalid > partial/missing mix > all present > none.
// Change prefers problem flags (dirty, untracked, ahead, behind, error, …)
// and only reports clean when every present replica is clean and git was on.
func rollupProjectOverall(replicas []api.ReplicaStatus, withGit bool) api.ProjectOverall {
	out := api.ProjectOverall{ReplicaCount: len(replicas)}
	if len(replicas) == 0 {
		out.Presence = "none"
		out.Change = "-"
		return out
	}

	var present, missing, invalid int
	for _, r := range replicas {
		switch r.Presence {
		case api.PresencePresent:
			present++
		case api.PresenceInvalid:
			invalid++
		default:
			missing++
		}
	}
	out.PresentCount = present

	switch {
	case invalid > 0 && present == 0 && missing == 0:
		out.Presence = "invalid"
	case invalid > 0:
		out.Presence = "invalid"
	case present == len(replicas):
		out.Presence = "all-present"
	case present == 0:
		out.Presence = "missing"
	default:
		out.Presence = "partial"
	}

	if !withGit {
		out.Change = "-"
		return out
	}
	if present == 0 {
		out.Change = "-"
		return out
	}

	// Union change tokens from present replicas; drop bare clean until end.
	tokenSet := make(map[string]struct{})
	var order []string
	addTok := func(t string) {
		t = strings.TrimSpace(t)
		if t == "" || t == "-" {
			return
		}
		if _, ok := tokenSet[t]; ok {
			return
		}
		tokenSet[t] = struct{}{}
		order = append(order, t)
	}
	allClean := true
	for _, r := range replicas {
		if r.Presence != api.PresencePresent {
			continue
		}
		ch := strings.TrimSpace(r.Change)
		if ch == "" || ch == "-" {
			allClean = false
			continue
		}
		parts := strings.Fields(ch)
		hasClean := false
		hasProblem := false
		for _, p := range parts {
			if p == "clean" {
				hasClean = true
				continue
			}
			hasProblem = true
			addTok(p)
		}
		if hasProblem || !hasClean {
			allClean = false
		}
	}
	if len(order) == 0 {
		if allClean {
			out.Change = "clean"
		} else {
			out.Change = "-"
		}
		return out
	}
	// Prefer problem tokens only (no trailing clean when mixed).
	out.Change = strings.Join(order, " ")
	return out
}

func (r Resolver) replicaStatus(ctx context.Context, p config.Project, replica string, opts StatusOptions) (api.ReplicaStatus, error) {
	var err error
	if strings.TrimSpace(replica) == "" {
		replica, err = r.DefaultReplicaName(ctx, p, opts.Network)
		if err != nil {
			return api.ReplicaStatus{}, err
		}
	}
	path, err := r.ReplicaPath(p, replica)
	if err != nil {
		return api.ReplicaStatus{}, fmt.Errorf("%s: %w", p.Name, err)
	}
	pres := presence.Classify(path)
	row := api.NewReplicaStatus()
	row.URI = id.ReplicaURI(p.Domain, p.Name, replica)
	row.Project = p.Name
	row.Replica = replica
	row.Domain = p.Domain
	row.Path = path
	row.Presence = api.PresenceFrom(pres)
	row.Tags = append([]string(nil), p.Tags...)
	row.Remote = p.Remote

	if opts.Git && pres == presence.Present && r.Git != nil {
		probe, err := r.Git.Probe(ctx, path)
		if err != nil {
			row.Change = "error"
			row.Branch = "-"
			return row, nil
		}
		row.Change = probe.Flags()
		if probe.Branch != "" {
			row.Branch = probe.Branch
		}
	}
	return row, nil
}

// classifyWorkspaceDir reports whether the project workspace directory exists.
// Unlike replica presence, this is not a git-checkout classifier.
func classifyWorkspaceDir(path string) api.Presence {
	if path == "" {
		return api.PresenceMissing
	}
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return api.PresenceMissing
		}
		return api.PresenceInvalid
	}
	if !fi.IsDir() {
		return api.PresenceInvalid
	}
	return api.PresencePresent
}
