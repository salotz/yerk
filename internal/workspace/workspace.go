// Package workspace resolves project and replica paths on the host.
//
// Each catalog project resolves to an absolute **project workspace** directory
// (the folder that owns that project's replicas), e.g. `…/devel/yerk`.
//
// Path resolution (ADR 014):
//
//	1. host [[projects]] absolute or ~/ path
//	2. host [[projects]] relative path → <domains[domain]>/<path>
//	3. else → <domains[domain]>/<name>
//
// Effective style (ADR 013/018 placement merge) decides replica math:
//
//	workspace-dir: <workspace>/<replica>              → …/yerk/main
//	project-dir:   <dir(workspace)>/<name>__<replica> → …/yerk__main
//	name-tags:     main = <workspace>/<name>
//	               other = <workspace>/<name>__<replica>
//	               optional main_dir / replica_dir params
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/salotz/yerk/internal/config"
)

// Style names accepted in config / placement.
const (
	StyleWorkspaceDir = "workspace-dir"
	StyleProjectDir   = "project-dir"
	// StyleNameTags: bare main + __ tagged siblings (ADR 018).
	StyleNameTags = "name-tags"
	// NameTagSep separates project name from replica distinguisher under name-tags.
	NameTagSep = "__"
)

// Layout holds workspace policy and host config for path lookup.
type Layout struct {
	Style string
	Host  config.Config
	// Params are optional style path parameters (ADR 018; name-tags).
	Params config.StyleSpec
	// DefaultReplica is the main distinguisher name (catalog default or "main").
	// Used by name-tags so path math maps the main id to the bare project dir.
	DefaultReplica string
	// Warnf, if set, receives soft warnings (unused for path math today).
	Warnf func(format string, args ...any)
}

// NewLayout builds a Layout from full tool config, applying defaults.
// Prefer NewLayoutStyle when effective style comes from placement.Resolve.
func NewLayout(cfg config.Config) (Layout, error) {
	style := cfg.Workspace.Style
	if style == "" {
		style = StyleWorkspaceDir
	}
	return NewLayoutStyle(style, cfg, nil)
}

// NewLayoutStyle builds a Layout with an explicit effective style (ADR 013).
func NewLayoutStyle(style string, host config.Config, warnf func(string, ...any)) (Layout, error) {
	return NewLayoutFull(style, host, config.StyleSpec{}, "", warnf)
}

// NewLayoutFull builds a Layout with style, optional params, and main distinguisher.
func NewLayoutFull(style string, host config.Config, params config.StyleSpec, defaultReplica string, warnf func(string, ...any)) (Layout, error) {
	style = strings.TrimSpace(style)
	if style == "" {
		style = StyleWorkspaceDir
	}
	if err := ValidateStyle(style); err != nil {
		return Layout{}, err
	}
	if err := ValidateStyleParams(style, params); err != nil {
		return Layout{}, err
	}
	def := strings.TrimSpace(defaultReplica)
	if def == "" {
		def = "main"
	}
	// Params carry dirs only; name comes from style argument.
	params.Style = style
	return Layout{
		Style:          style,
		Host:           host,
		Params:         params,
		DefaultReplica: def,
		Warnf:          warnf,
	}, nil
}

// ValidateStyle reports whether style is known for path math.
func ValidateStyle(style string) error {
	switch strings.TrimSpace(style) {
	case StyleWorkspaceDir, StyleProjectDir, StyleNameTags:
		return nil
	case "":
		return fmt.Errorf("empty workspace style")
	default:
		return fmt.Errorf("unknown workspace style %q (want %s, %s, or %s)",
			style, StyleWorkspaceDir, StyleProjectDir, StyleNameTags)
	}
}

// ValidateStyleParams rejects params on styles that do not accept them.
func ValidateStyleParams(style string, params config.StyleSpec) error {
	if !params.HasParams() {
		return nil
	}
	switch strings.TrimSpace(style) {
	case StyleNameTags:
		return nil
	case StyleWorkspaceDir, StyleProjectDir:
		return fmt.Errorf("workspace style %q does not accept main_dir/replica_dir params (ADR 018)", style)
	default:
		return fmt.Errorf("workspace style %q: params not allowed", style)
	}
}

// ProjectDir returns the absolute project workspace directory.
// That directory owns the project's replicas under workspace-dir / name-tags.
func (l Layout) ProjectDir(p config.Project) (string, error) {
	return l.projectWorkspace(p)
}

// ReplicaDir returns the checkout path for one replica of a project.
// replica is the distinguisher (often the default branch short name).
func (l Layout) ReplicaDir(p config.Project, replica string) (string, error) {
	replica = strings.TrimSpace(replica)
	if replica == "" {
		return "", fmt.Errorf("project %q: empty replica distinguisher", p.Name)
	}
	// Reject path separators in distinguisher so style math stays a single segment.
	if strings.ContainsAny(replica, `/\`) {
		return "", fmt.Errorf("project %q: replica %q must be a single path segment", p.Name, replica)
	}

	ws, err := l.projectWorkspace(p)
	if err != nil {
		return "", err
	}

	switch l.Style {
	case StyleWorkspaceDir:
		return filepath.Join(ws, replica), nil
	case StyleProjectDir:
		name, err := projectNameSegment(p)
		if err != nil {
			return "", err
		}
		parent := filepath.Dir(ws)
		return filepath.Join(parent, name+NameTagSep+replica), nil
	case StyleNameTags:
		return l.nameTagsReplicaDir(p, ws, replica)
	default:
		return "", fmt.Errorf("unknown workspace style %q", l.Style)
	}
}

func (l Layout) nameTagsReplicaDir(p config.Project, ws, replica string) (string, error) {
	name, err := projectNameSegment(p)
	if err != nil {
		return "", err
	}
	isMain := replica == l.DefaultReplica || replica == name

	if isMain {
		if md := strings.TrimSpace(l.Params.MainDir); md != "" {
			return expandAbsParam("main_dir", md)
		}
		return filepath.Join(ws, name), nil
	}

	container := ws
	if rd := strings.TrimSpace(l.Params.ReplicaDir); rd != "" {
		var err error
		container, err = expandAbsParam("replica_dir", rd)
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(container, name+NameTagSep+replica), nil
}

func expandAbsParam(key, raw string) (string, error) {
	expanded, err := config.ExpandUser(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("style param %s: %w", key, err)
	}
	if !filepath.IsAbs(expanded) {
		return "", fmt.Errorf("style param %s must be absolute or ~/… (got %q)", key, raw)
	}
	return filepath.Clean(expanded), nil
}

func projectNameSegment(p config.Project) (string, error) {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return "", fmt.Errorf("project has empty name (needed for layout path math)")
	}
	if strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("project name %q must be a single path segment", name)
	}
	return name, nil
}

func (l Layout) projectWorkspace(p config.Project) (string, error) {
	return l.Host.ProjectWorkspacePath(p.Domain, p.Name)
}

// ListLiveReplicas returns distinguisher names for on-disk checkouts that look
// like usable git repos under this project's layout (presence=present).
// Missing default/main is not invented here — callers may add it.
// Names are sorted lexicographically.
func (l Layout) ListLiveReplicas(p config.Project) ([]string, error) {
	switch l.Style {
	case StyleWorkspaceDir:
		return l.listLiveWorkspaceDir(p)
	case StyleProjectDir:
		return l.listLiveProjectDir(p)
	case StyleNameTags:
		return l.listLiveNameTags(p)
	default:
		return nil, fmt.Errorf("unknown workspace style %q", l.Style)
	}
}

func (l Layout) listLiveWorkspaceDir(p config.Project) ([]string, error) {
	ws, err := l.projectWorkspace(p)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(ws)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list workspace %s: %w", ws, err)
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if name == "" || name == "." || name == ".." || strings.HasPrefix(name, ".") {
			continue
		}
		if strings.ContainsAny(name, `/\`) {
			continue
		}
		path := filepath.Join(ws, name)
		if !isLiveCheckout(path) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func (l Layout) listLiveProjectDir(p config.Project) ([]string, error) {
	ws, err := l.projectWorkspace(p)
	if err != nil {
		return nil, err
	}
	name, err := projectNameSegment(p)
	if err != nil {
		return nil, err
	}
	prefix := name + NameTagSep
	parent := filepath.Dir(ws)
	entries, err := os.ReadDir(parent)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list project-dir parent %s: %w", parent, err)
	}
	var names []string
	for _, e := range entries {
		ent := e.Name()
		if !strings.HasPrefix(ent, prefix) {
			continue
		}
		rep := strings.TrimPrefix(ent, prefix)
		if rep == "" || strings.ContainsAny(rep, `/\`) {
			continue
		}
		path := filepath.Join(parent, ent)
		if !isLiveCheckout(path) {
			continue
		}
		names = append(names, rep)
	}
	sort.Strings(names)
	return names, nil
}

func (l Layout) listLiveNameTags(p config.Project) ([]string, error) {
	name, err := projectNameSegment(p)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var names []string
	add := func(rep string) {
		rep = strings.TrimSpace(rep)
		if rep == "" {
			return
		}
		if _, ok := seen[rep]; ok {
			return
		}
		seen[rep] = struct{}{}
		names = append(names, rep)
	}

	// Main checkout (bare name or main_dir).
	mainPath, err := l.ReplicaDir(p, l.DefaultReplica)
	if err == nil && isLiveCheckout(mainPath) {
		add(l.DefaultReplica)
	}

	// Tagged siblings under replica container.
	container, err := l.nameTagsTagContainer(p)
	if err != nil {
		return nil, err
	}
	prefix := name + NameTagSep
	entries, err := os.ReadDir(container)
	if err != nil {
		if os.IsNotExist(err) {
			sort.Strings(names)
			return names, nil
		}
		return nil, fmt.Errorf("list name-tags container %s: %w", container, err)
	}
	for _, e := range entries {
		ent := e.Name()
		if !strings.HasPrefix(ent, prefix) {
			continue
		}
		rep := strings.TrimPrefix(ent, prefix)
		if rep == "" || strings.ContainsAny(rep, `/\`) {
			continue
		}
		path := filepath.Join(container, ent)
		if !isLiveCheckout(path) {
			continue
		}
		add(rep)
	}
	sort.Strings(names)
	return names, nil
}

func (l Layout) nameTagsTagContainer(p config.Project) (string, error) {
	if rd := strings.TrimSpace(l.Params.ReplicaDir); rd != "" {
		return expandAbsParam("replica_dir", rd)
	}
	return l.projectWorkspace(p)
}

func isLiveCheckout(path string) bool {
	fi, err := os.Stat(path)
	if err != nil || !fi.IsDir() {
		return false
	}
	gitPath := filepath.Join(path, ".git")
	gfi, err := os.Lstat(gitPath)
	if err != nil {
		return false
	}
	return gfi.IsDir() || gfi.Mode().IsRegular()
}

// EnsureDir creates dir (mkdir -p). Never deletes.
// Used for project workspace materialization (yerk workspace ensure).
func EnsureDir(dir string) error {
	dir = strings.TrimSpace(dir)
	if dir == "" || dir == "." {
		return fmt.Errorf("cannot ensure empty directory path")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure dir %s: %w", dir, err)
	}
	return nil
}

// EnsureParents creates the parent directory of path (mkdir -p).
// Used before clone so the replica leaf can be created by git. Never deletes.
func EnsureParents(path string) error {
	parent := filepath.Dir(path)
	if parent == "" || parent == "." {
		return fmt.Errorf("cannot ensure parents for %q", path)
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("ensure parents of %s: %w", path, err)
	}
	return nil
}

// EnsureReplicaDir creates the replica directory itself (mkdir -p).
// Prefer EnsureDir for project workspaces; this remains for callers that
// intentionally materialize a replica leaf without cloning.
func EnsureReplicaDir(replicaPath string) error {
	return EnsureDir(replicaPath)
}
