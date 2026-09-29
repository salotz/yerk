// Package project joins catalog, layout, presence, and git for common ops.
package project

import (
	"context"
	"fmt"
	"strings"

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

// Row is one status table row.
type Row struct {
	Name     string
	Domain   string
	Replica  string
	Presence presence.Status
	Change   string
	Branch   string
	Path     string
	Tags     []string
	Remote   string
}

// StatusOptions controls status collection.
type StatusOptions struct {
	Git     bool
	Network bool // resolve default branch via ls-remote
}

// Status builds rows for the given projects (caller selects by name/tag/all).
func (r Resolver) Status(ctx context.Context, projects []config.Project, opts StatusOptions) ([]Row, error) {
	rows := make([]Row, 0, len(projects))
	for _, p := range projects {
		row, err := r.statusOne(ctx, p, opts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (r Resolver) statusOne(ctx context.Context, p config.Project, opts StatusOptions) (Row, error) {
	replica, err := r.DefaultReplicaName(ctx, p, opts.Network)
	if err != nil {
		return Row{}, err
	}
	path, err := r.ReplicaPath(p, replica)
	if err != nil {
		return Row{}, fmt.Errorf("%s: %w", p.Name, err)
	}
	pres := presence.Classify(path)
	row := Row{
		Name:     p.Name,
		Domain:   p.Domain,
		Replica:  replica,
		Presence: pres,
		Change:   "-",
		Branch:   "-",
		Path:     path,
		Tags:     p.Tags,
		Remote:   p.Remote,
	}
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
