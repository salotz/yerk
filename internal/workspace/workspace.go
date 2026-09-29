// Package workspace resolves project and replica paths on the host.
//
// Each catalog project resolves to an absolute **project workspace** directory
// (the folder that owns that project's replicas), e.g. `…/devel/yerk`.
//
// Path resolution (ADR 008):
//
//	- absolute catalog path → use as workspace (host escape hatch)
//	- relative catalog path → join config [domains.<domain>] root + path
//
// Style decides how a replica distinguisher maps under that workspace:
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

// Style names accepted in config.
const (
	StyleWorkspaceDir = "workspace-dir"
	StyleProjectDir   = "project-dir"
)

// Layout holds workspace policy and domain roots loaded from tool config.
type Layout struct {
	Style   string
	Domains map[string]string
}

// NewLayout builds a Layout from full tool config, applying defaults.
func NewLayout(cfg config.Config) (Layout, error) {
	style := cfg.Workspace.Style
	if style == "" {
		style = StyleWorkspaceDir
	}
	switch style {
	case StyleWorkspaceDir, StyleProjectDir:
	default:
		return Layout{}, fmt.Errorf("unknown workspace style %q (want %s or %s)",
			style, StyleWorkspaceDir, StyleProjectDir)
	}
	domains := cfg.Domains
	if domains == nil {
		domains = map[string]string{}
	}
	return Layout{Style: style, Domains: domains}, nil
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
	raw := strings.TrimSpace(p.Path)
	name := strings.TrimSpace(p.Name)
	if name == "" {
		name = "?"
	}
	if raw == "" {
		return "", fmt.Errorf("project %q has no path: set catalog path relative to the domain root (e.g. devel/yerk) or an absolute project workspace", name)
	}

	expanded, err := config.ExpandUser(raw)
	if err != nil {
		return "", fmt.Errorf("project %q path: %w", name, err)
	}

	if filepath.IsAbs(expanded) {
		return filepath.Clean(expanded), nil
	}

	// Relative: require domain + configured domain root.
	domain := strings.TrimSpace(p.Domain)
	if domain == "" {
		return "", fmt.Errorf("project %q path %q is relative but domain is empty: set domain to select [domains.<name>] in config.toml, or use an absolute path", name, raw)
	}
	root, ok, err := l.domainRoot(domain)
	if err != nil {
		return "", fmt.Errorf("project %q: %w", name, err)
	}
	if !ok {
		return "", fmt.Errorf("project %q domain %q has no root in config.toml [domains]; add e.g. %s = \"~/tree/%s\"", name, domain, domain, domain)
	}
	return filepath.Join(root, filepath.Clean(expanded)), nil
}

func (l Layout) domainRoot(domain string) (string, bool, error) {
	// Reuse Config.DomainRoot logic via a transient Config so ~ expansion stays one place.
	cfg := config.Config{Domains: l.Domains}
	return cfg.DomainRoot(domain)
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
