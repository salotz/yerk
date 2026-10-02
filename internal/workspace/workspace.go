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
// Effective style (ADR 013 placement merge) decides replica math:
//
//	workspace-dir: <workspace>/<replica>           → …/yerk/main
//	project-dir:   <dir(workspace)>/<name>__<replica> → …/yerk__main
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/salotz/yerk/internal/config"
)

// Style names accepted in config / placement.
const (
	StyleWorkspaceDir = "workspace-dir"
	StyleProjectDir   = "project-dir"
	// StyleNameTags is reserved (path math in a later phase).
	StyleNameTags = "name-tags"
)

// Layout holds workspace policy and host config for path lookup.
type Layout struct {
	Style string
	Host  config.Config
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
	style = strings.TrimSpace(style)
	if style == "" {
		style = StyleWorkspaceDir
	}
	if err := ValidateStyle(style); err != nil {
		return Layout{}, err
	}
	return Layout{Style: style, Host: host, Warnf: warnf}, nil
}

// ValidateStyle reports whether style is known for path math.
// name-tags is recognized as locked but not implemented yet.
func ValidateStyle(style string) error {
	switch strings.TrimSpace(style) {
	case StyleWorkspaceDir, StyleProjectDir:
		return nil
	case StyleNameTags:
		return fmt.Errorf("workspace style %q is not implemented yet (path math ships in a later phase)", StyleNameTags)
	case "":
		return fmt.Errorf("empty workspace style")
	default:
		return fmt.Errorf("unknown workspace style %q (want %s or %s)",
			style, StyleWorkspaceDir, StyleProjectDir)
	}
}

// ProjectDir returns the absolute project workspace directory.
// That directory owns the project's replicas under workspace-dir style.
func (l Layout) ProjectDir(p config.Project) (string, error) {
	return l.projectWorkspace(p)
}

// ReplicaDir returns the checkout path for one replica of a project.
// replica is the distinguisher (often the default branch short name).
func (l Layout) ReplicaDir(p config.Project, replica string) (string, error) {
	ws, err := l.projectWorkspace(p)
	if err != nil {
		return "", err
	}
	replica = strings.TrimSpace(replica)
	if replica == "" {
		return "", fmt.Errorf("project %q: empty replica distinguisher", p.Name)
	}
	// Reject path separators in distinguisher so style math stays a single segment.
	if strings.ContainsAny(replica, `/\`) {
		return "", fmt.Errorf("project %q: replica %q must be a single path segment", p.Name, replica)
	}

	switch l.Style {
	case StyleWorkspaceDir:
		return filepath.Join(ws, replica), nil
	case StyleProjectDir:
		name := strings.TrimSpace(p.Name)
		if name == "" {
			return "", fmt.Errorf("project has empty name (needed for %s layout)", StyleProjectDir)
		}
		if strings.ContainsAny(name, `/\`) {
			return "", fmt.Errorf("project name %q must be a single path segment", name)
		}
		parent := filepath.Dir(ws)
		return filepath.Join(parent, name+"__"+replica), nil
	default:
		return "", fmt.Errorf("unknown workspace style %q", l.Style)
	}
}

func (l Layout) projectWorkspace(p config.Project) (string, error) {
	return l.Host.ProjectWorkspacePath(p.Domain, p.Name)
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
