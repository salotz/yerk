package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/id"
	"github.com/salotz/yerk/internal/placement"
	"github.com/salotz/yerk/internal/presence"
	"github.com/salotz/yerk/internal/workspace"
)

// ConfigResolve builds a placement contribution report for one project (config resolve).
func (r Resolver) ConfigResolve(p config.Project) (api.ConfigResolve, error) {
	in := placement.Input{
		Host:     r.Cfg,
		Project:  p,
		CLIStyle: r.CLIStyle,
	}
	ws, err := r.WorkspacePath(p)
	if err != nil {
		// Still explain with cwd anchor if workspace cannot be resolved.
		in.Anchor = ""
	} else {
		in.Anchor = ws
	}
	rep, err := placement.Explain(in)
	if err != nil {
		return api.ConfigResolve{}, fmt.Errorf("%s: %w", p.ID(), err)
	}
	// Surface placement warnings on stderr like other commands.
	for _, w := range rep.Effective.Warnings {
		r.warnf("warning: %s\n", w)
	}
	out := api.NewConfigResolve()
	out.URI = id.ProjectURI(p.Domain, p.Name)
	out.Target = p.ID()
	out.Anchor = in.Anchor
	out.WorkspacePath = ws
	out.EffectiveStyle = rep.Effective.Style
	out.Bound = rep.Effective.Bound
	out.Warnings = append([]string(nil), rep.Effective.Warnings...)
	out.Files = append([]string(nil), rep.Files...)
	out.Contributions = make([]api.ConfigResolveContribution, 0, len(rep.Contributions))
	for _, c := range rep.Contributions {
		out.Contributions = append(out.Contributions, api.ConfigResolveContribution{
			Order:   c.Order,
			Layer:   c.Layer,
			Path:    c.Path,
			Key:     c.Key,
			Value:   c.Value,
			Applies: c.Applies,
			Note:    c.Note,
		})
	}
	return out, nil
}

// ProjectInfo builds a ProjectInfo resource for one catalog project (ADR 015).
func (r Resolver) ProjectInfo(p config.Project) (api.ProjectInfo, error) {
	ws, err := r.WorkspacePath(p)
	if err != nil {
		return api.ProjectInfo{}, fmt.Errorf("%s: %w", p.ID(), err)
	}
	eff, err := r.EffectivePlacement(p)
	if err != nil {
		return api.ProjectInfo{}, fmt.Errorf("%s: %w", p.ID(), err)
	}
	out := api.NewProjectInfo()
	out.URI = id.ProjectURI(p.Domain, p.Name)
	out.Name = p.Name
	out.Domain = p.Domain
	out.Remote = p.Remote
	out.DefaultReplica = p.DefaultReplica
	out.Tags = append([]string(nil), p.Tags...)
	out.WorkspacePath = ws
	out.WorkspacePresence = classifyWorkspaceDir(ws)
	out.Placement = placementInfo(eff)
	return out, nil
}

// ReplicaInfo builds a ReplicaInfo for project + replica distinguisher (ADR 015).
// Does not run git change probes (cheap read).
func (r Resolver) ReplicaInfo(p config.Project, replica string) (api.ReplicaInfo, error) {
	replica = strings.TrimSpace(replica)
	if replica == "" {
		return api.ReplicaInfo{}, fmt.Errorf("%s: replica distinguisher required", p.ID())
	}
	ws, err := r.WorkspacePath(p)
	if err != nil {
		return api.ReplicaInfo{}, fmt.Errorf("%s: %w", p.ID(), err)
	}
	path, err := r.ReplicaPath(p, replica)
	if err != nil {
		return api.ReplicaInfo{}, fmt.Errorf("%s: %w", p.ID(), err)
	}
	eff, err := r.EffectivePlacement(p)
	if err != nil {
		return api.ReplicaInfo{}, fmt.Errorf("%s: %w", p.ID(), err)
	}
	out := api.NewReplicaInfo()
	out.URI = id.ReplicaURI(p.Domain, p.Name, replica)
	out.Project = p.Name
	out.Replica = replica
	out.Domain = p.Domain
	out.Remote = p.Remote
	out.Tags = append([]string(nil), p.Tags...)
	out.Path = path
	out.Presence = api.PresenceFrom(presence.Classify(path))
	out.WorkspacePath = ws
	out.Placement = placementInfo(eff)
	return out, nil
}

func placementInfo(eff placement.Effective) *api.PlacementInfo {
	return &api.PlacementInfo{
		Style:    eff.Style,
		Bound:    eff.Bound,
		Sources:  append([]string(nil), eff.Sources...),
		Warnings: append([]string(nil), eff.Warnings...),
	}
}

// LookupKind selects which resource lookup must return.
type LookupKind int

const (
	// LookupAny returns replica info when under a replica root, else project.
	LookupAny LookupKind = iota
	// LookupProject returns project info for any path under a workspace.
	LookupProject
	// LookupReplica returns replica info only (errors if only under workspace).
	LookupReplica
)

// LookupResult is the outcome of a path lookup (exactly one info kind set).
type LookupResult struct {
	Project *api.ProjectInfo
	Replica *api.ReplicaInfo
}

// LookupPath maps an on-disk path to project or replica info (ADR 015).
// path may be relative (cwd) or absolute; ~ is expanded.
func (r Resolver) LookupPath(projects []config.Project, path string, kind LookupKind) (LookupResult, error) {
	abs, err := absLookupPath(path)
	if err != nil {
		return LookupResult{}, err
	}

	type hit struct {
		p       config.Project
		ws      string
		root    string // matched root (workspace or replica)
		replica string // non-empty when replica root matched
		depth   int    // len(root) — longer wins
	}
	var best *hit

	for _, p := range projects {
		layout, _, err := r.layoutFor(p)
		if err != nil {
			// Skip projects that cannot resolve placement/path on this host.
			continue
		}
		ws, err := layout.ProjectDir(p)
		if err != nil {
			continue
		}
		ws = filepath.Clean(ws)
		if !pathUnderOrEqual(abs, ws) {
			continue
		}

		// Prefer a replica root under this workspace when one matches.
		repName, repPath, ok := matchReplicaUnder(layout, p, ws, abs)
		if ok {
			h := hit{p: p, ws: ws, root: repPath, replica: repName, depth: len(repPath)}
			if best == nil || h.depth > best.depth {
				cp := h
				best = &cp
			}
			continue
		}

		// Workspace-only match (project scope).
		if kind == LookupReplica {
			continue
		}
		h := hit{p: p, ws: ws, root: ws, depth: len(ws)}
		if best == nil || h.depth > best.depth {
			cp := h
			best = &cp
		}
	}

	if best == nil {
		if kind == LookupReplica {
			return LookupResult{}, fmt.Errorf("path %q is not under a known replica", abs)
		}
		return LookupResult{}, fmt.Errorf("path %q is not under a known project workspace", abs)
	}

	if best.replica != "" && kind != LookupProject {
		info, err := r.ReplicaInfo(best.p, best.replica)
		if err != nil {
			return LookupResult{}, err
		}
		info.MatchedPath = abs
		return LookupResult{Replica: &info}, nil
	}

	// Project info (explicit LookupProject, or LookupAny with workspace-only hit).
	info, err := r.ProjectInfo(best.p)
	if err != nil {
		return LookupResult{}, err
	}
	info.MatchedPath = abs
	return LookupResult{Project: &info}, nil
}

// matchReplicaUnder finds the best replica root that contains abs.
// For workspace-dir: first segment under ws is the distinguisher candidate.
// For project-dir / name-tags: bare main and/or name__replica siblings.
func matchReplicaUnder(layout workspace.Layout, p config.Project, ws, abs string) (replica, repPath string, ok bool) {
	style := layout.Style
	switch style {
	case workspace.StyleWorkspaceDir:
		rel, err := filepath.Rel(ws, abs)
		if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
			return "", "", false
		}
		// abs == ws → not under a replica leaf yet.
		parts := splitPath(rel)
		if len(parts) == 0 {
			return "", "", false
		}
		cand := parts[0]
		path, err := layout.ReplicaDir(p, cand)
		if err != nil {
			return "", "", false
		}
		path = filepath.Clean(path)
		if pathUnderOrEqual(abs, path) {
			return cand, path, true
		}
		return "", "", false

	case workspace.StyleNameTags:
		return matchNameTagsReplica(layout, p, ws, abs)

	case workspace.StyleProjectDir:
		// Replicas are siblings: <parent>/<name>__<replica>. Walk abs upward
		// and detect a leaf matching that pattern next to the workspace.
		parent := filepath.Dir(ws)
		name := strings.TrimSpace(p.Name)
		prefix := name + workspace.NameTagSep
		for cur := abs; ; {
			if filepath.Dir(cur) == parent {
				base := filepath.Base(cur)
				if strings.HasPrefix(base, prefix) {
					cand := strings.TrimPrefix(base, prefix)
					if cand != "" {
						path, err := layout.ReplicaDir(p, cand)
						if err == nil {
							path = filepath.Clean(path)
							if pathUnderOrEqual(abs, path) {
								return cand, path, true
							}
						}
					}
				}
			}
			next := filepath.Dir(cur)
			if next == cur {
				break
			}
			// Stop once we leave the workspace parent tree.
			if cur == parent {
				break
			}
			cur = next
		}
		return "", "", false

	default:
		return "", "", false
	}
}

// matchNameTagsReplica matches bare main (container/<name> or main_dir) and
// tagged siblings (container/<name>__R or replica_dir).
func matchNameTagsReplica(layout workspace.Layout, p config.Project, ws, abs string) (replica, repPath string, ok bool) {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return "", "", false
	}
	def := layout.DefaultReplica
	if def == "" {
		def = "main"
	}

	// Prefer longer root: try known live-ish candidates via ReplicaDir for
	// main + any tagged leaf under abs walk.
	type cand struct {
		rep  string
		path string
	}
	var cands []cand

	if mainPath, err := layout.ReplicaDir(p, def); err == nil {
		cands = append(cands, cand{rep: def, path: filepath.Clean(mainPath)})
	}

	// Walk upward from abs; if base is name__R under a plausible parent, add it.
	prefix := name + workspace.NameTagSep
	for cur := abs; ; {
		base := filepath.Base(cur)
		if base == name {
			// Bare project dir → main distinguisher.
			cands = append(cands, cand{rep: def, path: filepath.Clean(cur)})
		}
		if strings.HasPrefix(base, prefix) {
			tag := strings.TrimPrefix(base, prefix)
			if tag != "" && !strings.ContainsAny(tag, `/\`) {
				if path, err := layout.ReplicaDir(p, tag); err == nil {
					cands = append(cands, cand{rep: tag, path: filepath.Clean(path)})
				} else {
					cands = append(cands, cand{rep: tag, path: filepath.Clean(cur)})
				}
			}
		}
		next := filepath.Dir(cur)
		if next == cur {
			break
		}
		// Stop after leaving workspace and common parents (main_dir may be outside ws).
		if cur == ws || cur == filepath.Dir(ws) {
			// one more step allowed for main_dir outside; continue a few levels
		}
		cur = next
		// Bound walk: don't climb forever.
		if len(filepath.Clean(cur)) < 2 {
			break
		}
	}

	var best *cand
	for i := range cands {
		c := cands[i]
		if !pathUnderOrEqual(abs, c.path) {
			continue
		}
		if best == nil || len(c.path) > len(best.path) {
			cp := c
			best = &cp
		}
	}
	if best == nil {
		return "", "", false
	}
	return best.rep, best.path, true
}

func absLookupPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	expanded, err := config.ExpandUser(path)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(expanded) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve path: %w", err)
		}
		expanded = filepath.Join(cwd, expanded)
	}
	// Prefer EvalSymlinks when the path exists; else Clean only.
	if resolved, err := filepath.EvalSymlinks(expanded); err == nil {
		return resolved, nil
	}
	return filepath.Clean(expanded), nil
}

func pathUnderOrEqual(path, root string) bool {
	path = filepath.Clean(path)
	root = filepath.Clean(root)
	if path == root {
		return true
	}
	sep := string(os.PathSeparator)
	return strings.HasPrefix(path, root+sep)
}

func splitPath(rel string) []string {
	rel = filepath.ToSlash(rel)
	var parts []string
	for _, p := range strings.Split(rel, "/") {
		if p == "" || p == "." {
			continue
		}
		parts = append(parts, p)
	}
	return parts
}
