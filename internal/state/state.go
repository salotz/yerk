// Package state reads and writes host project bindings under XDG state (ADR 013).
//
//	$XDG_STATE_HOME/yerk/projects/<domain>/<project>/state.json
//
// Override root: YERK__STATE_DIR.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/salotz/yerk/internal/api"
)

const (
	// AppName is the state directory leaf under XDG state home.
	AppName = "yerk"
	// FileName is the per-project binding basename.
	FileName = "state.json"
)

// Project holds tool-written binding for one catalog project on this host.
type Project struct {
	// APIVersion matches api resources (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// WorkspaceStyle is the bound style name used for path math (ADR 013).
	WorkspaceStyle string `json:"workspaceStyle,omitempty"`
	// BoundAt is when the binding was first written (RFC3339).
	BoundAt string `json:"boundAt,omitempty"`
	// UpdatedAt is when the binding was last written.
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// Dir returns the yerk state root ($XDG_STATE_HOME/yerk or YERK__STATE_DIR).
func Dir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("YERK__STATE_DIR")); v != "" {
		return v, nil
	}
	base, err := userStateHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, AppName), nil
}

func userStateHome() (string, error) {
	if v := strings.TrimSpace(os.Getenv("XDG_STATE_HOME")); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve state home: %w", err)
	}
	// XDG Base Directory default when XDG_STATE_HOME is unset.
	return filepath.Join(home, ".local", "state"), nil
}

// ProjectFile returns the absolute path to state.json for domain/project.
func ProjectFile(domain, project string) (string, error) {
	domain = strings.TrimSpace(domain)
	project = strings.TrimSpace(project)
	if domain == "" || project == "" {
		return "", fmt.Errorf("state path requires domain and project name")
	}
	if strings.ContainsAny(domain, `/\`) || strings.ContainsAny(project, `/\`) {
		return "", fmt.Errorf("state path segments must not contain separators")
	}
	root, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "projects", domain, project, FileName), nil
}

// LoadProject reads state.json if present. ok is false when missing.
func LoadProject(domain, project string) (Project, bool, error) {
	path, err := ProjectFile(domain, project)
	if err != nil {
		return Project{}, false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Project{}, false, nil
		}
		return Project{}, false, fmt.Errorf("read state %s: %w", path, err)
	}
	var st Project
	if err := json.Unmarshal(data, &st); err != nil {
		return Project{}, false, fmt.Errorf("parse state %s: %w", path, err)
	}
	return st, true, nil
}

// SaveProject writes state.json (mkdir parents). Overwrites existing file.
func SaveProject(domain, project string, st Project) error {
	path, err := ProjectFile(domain, project)
	if err != nil {
		return err
	}
	if strings.TrimSpace(st.APIVersion) == "" {
		st.APIVersion = api.APIVersion
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if strings.TrimSpace(st.BoundAt) == "" {
		st.BoundAt = now
	}
	st.UpdatedAt = now

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("ensure state dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write state %s: %w", path, err)
	}
	return nil
}

// Exists reports whether a binding file is present.
func Exists(domain, project string) (bool, error) {
	_, ok, err := LoadProject(domain, project)
	return ok, err
}

// BindStyle writes workspace style if no binding exists yet.
// If a binding already exists, it is left unchanged (ensure no-op for state).
// Returns whether a new binding was written.
func BindStyle(domain, project, style string) (written bool, err error) {
	style = strings.TrimSpace(style)
	if style == "" {
		return false, fmt.Errorf("bind style: empty workspace style")
	}
	existing, ok, err := LoadProject(domain, project)
	if err != nil {
		return false, err
	}
	if ok {
		return false, nil
	}
	st := Project{
		APIVersion:     api.APIVersion,
		WorkspaceStyle: style,
	}
	// Preserve BoundAt only on create via SaveProject.
	_ = existing
	if err := SaveProject(domain, project, st); err != nil {
		return false, err
	}
	return true, nil
}

// UpdateResult describes one UpdateStyle call.
type UpdateResult struct {
	// Created is true when no binding existed before.
	Created bool
	// Changed is true when workspaceStyle differed from the previous value
	// (always true when Created).
	Changed bool
	// Previous is the prior workspaceStyle (empty if Created).
	Previous string
	// Style is the style written.
	Style string
}

// UpdateStyle writes or overwrites the bound workspace style (explicit rebind).
// Unlike BindStyle, an existing binding is replaced. BoundAt is preserved when
// refreshing; UpdatedAt is always refreshed via SaveProject.
func UpdateStyle(domain, project, style string) (UpdateResult, error) {
	style = strings.TrimSpace(style)
	if style == "" {
		return UpdateResult{}, fmt.Errorf("update style: empty workspace style")
	}
	existing, ok, err := LoadProject(domain, project)
	if err != nil {
		return UpdateResult{}, err
	}
	res := UpdateResult{Style: style}
	st := Project{
		APIVersion:     api.APIVersion,
		WorkspaceStyle: style,
	}
	if ok {
		res.Previous = strings.TrimSpace(existing.WorkspaceStyle)
		st.BoundAt = existing.BoundAt
		if res.Previous != style {
			res.Changed = true
		}
	} else {
		res.Created = true
		res.Changed = true
	}
	if err := SaveProject(domain, project, st); err != nil {
		return UpdateResult{}, err
	}
	return res, nil
}
