// Package project joins catalog, layout, presence, and git for common ops.
package project

import (
	"context"
	"fmt"
	"io"
	"os"
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

// ProjectStatus builds a ProjectStatus with workspace path and default-replica summary.
func (r Resolver) ProjectStatus(ctx context.Context, p config.Project, opts StatusOptions) (api.ProjectStatus, error) {
	ws, err := r.WorkspacePath(p)
	if err != nil {
		return api.ProjectStatus{}, fmt.Errorf("%s: %w", p.Name, err)
	}
	rep, err := r.replicaStatus(ctx, p, "", opts)
	if err != nil {
		return api.ProjectStatus{}, err
	}
	out := api.NewProjectStatus()
	out.URI = id.ProjectURI(p.Domain, p.Name)
	out.Name = p.Name
	out.Domain = p.Domain
	out.WorkspacePath = ws
	out.WorkspacePresence = classifyWorkspaceDir(ws)
	out.Tags = append([]string(nil), p.Tags...)
	out.Remote = p.Remote
	sum := rep.Summary()
	out.DefaultReplica = &sum
	return out, nil
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
