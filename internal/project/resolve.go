// Package project joins catalog, layout, presence, and git for common ops.
package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/gitcmd"
	"github.com/salotz/yerk/internal/presence"
	"github.com/salotz/yerk/internal/workspace"
)

// FallbackReplica is used when catalog and remote default branch are unavailable.
const FallbackReplica = "main"

// Resolver resolves paths and default replica names.
type Resolver struct {
	Layout workspace.Layout
	Git    gitcmd.Runner
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
	return Resolver{Layout: layout, Git: git}, nil
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
	return r.Layout.ProjectDir(p)
}

// ReplicaPath returns the on-disk path for project + replica distinguisher.
func (r Resolver) ReplicaPath(p config.Project, replica string) (string, error) {
	return r.Layout.ReplicaDir(p, replica)
}

// StatusOptions controls status collection.
type StatusOptions struct {
	Git     bool
	Network bool // resolve default branch via ls-remote
}

// Status builds ReplicaStatus resources for the given projects
// (caller selects by name/tag/all). Each row is the default-replica view
// used by today's status table; P6 will add project-scoped collection.
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
	out.Name = p.Name
	out.Domain = p.Domain
	out.WorkspacePath = ws
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
