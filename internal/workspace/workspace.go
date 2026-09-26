// Package workspace resolves project and replica paths on the host.
//
// MVP styles (from design notes):
//
//   - workspace-dir: projects/<project>/<replica>
//   - project-dir:   projects/<project>__<replica>
//
// Higher-level PRJX / bimtree domain layout is intentionally out of scope
// for path math here; callers pass an already-chosen root.
package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/salotz/yerk/internal/config"
)

// Style names accepted in config.
const (
	StyleWorkspaceDir = "workspace-dir"
	StyleProjectDir   = "project-dir"
)

// Layout computes on-disk paths for a configured workspace style.
type Layout struct {
	Root  string
	Style string
}

// NewLayout builds a Layout from config, applying defaults.
func NewLayout(ws config.Workspace) (Layout, error) {
	style := ws.Style
	if style == "" {
		style = StyleWorkspaceDir
	}
	switch style {
	case StyleWorkspaceDir, StyleProjectDir:
	default:
		return Layout{}, fmt.Errorf("unknown workspace style %q (want %s or %s)",
			style, StyleWorkspaceDir, StyleProjectDir)
	}
	return Layout{Root: ws.Root, Style: style}, nil
}

// ProjectDir returns the directory that owns a project's replicas.
// For workspace-dir this is <root>/<name>.
// For project-dir this is <root> (replicas are sibling dirs with __).
func (l Layout) ProjectDir(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("empty project name")
	}
	if err := l.requireRoot(); err != nil {
		return "", err
	}
	switch l.Style {
	case StyleWorkspaceDir:
		return filepath.Join(l.Root, name), nil
	case StyleProjectDir:
		return l.Root, nil
	default:
		return "", fmt.Errorf("unknown workspace style %q", l.Style)
	}
}

// ReplicaDir returns the checkout path for one replica of a project.
func (l Layout) ReplicaDir(project, replica string) (string, error) {
	project = strings.TrimSpace(project)
	replica = strings.TrimSpace(replica)
	if project == "" {
		return "", fmt.Errorf("empty project name")
	}
	if replica == "" {
		return "", fmt.Errorf("empty replica name")
	}
	if err := l.requireRoot(); err != nil {
		return "", err
	}
	switch l.Style {
	case StyleWorkspaceDir:
		return filepath.Join(l.Root, project, replica), nil
	case StyleProjectDir:
		return filepath.Join(l.Root, project+"__"+replica), nil
	default:
		return "", fmt.Errorf("unknown workspace style %q", l.Style)
	}
}

func (l Layout) requireRoot() error {
	if strings.TrimSpace(l.Root) == "" {
		return fmt.Errorf("workspace root is not set (config workspace.root or YERK__WORKSPACE_ROOT)")
	}
	return nil
}
