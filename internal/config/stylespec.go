package config

import (
	"fmt"
	"strings"
)

// StyleSpec is a workspace style name plus optional path parameters (ADR 018).
//
// TOML accepts either a plain string:
//
//	workspace_style = "name-tags"
//
// or an inline table:
//
//	workspace_style = { style = "name-tags", main_dir = "~/.bimker", replica_dir = "…" }
//
// Decoded via ParseStyleValue from interface{} (catalog/host flexible fields).
type StyleSpec struct {
	// Style is the base style name (workspace-dir, project-dir, name-tags).
	Style string `json:"style,omitempty"`
	// MainDir overrides the main replica checkout path (absolute or ~/…).
	MainDir string `json:"mainDir,omitempty"`
	// ReplicaDir is the container for tagged non-main replicas (absolute or ~/…).
	ReplicaDir string `json:"replicaDir,omitempty"`
}

// IsZero reports whether no style name or params are set.
func (s StyleSpec) IsZero() bool {
	return strings.TrimSpace(s.Style) == "" &&
		strings.TrimSpace(s.MainDir) == "" &&
		strings.TrimSpace(s.ReplicaDir) == ""
}

// Name returns the trimmed style name (empty if unset).
func (s StyleSpec) Name() string {
	return strings.TrimSpace(s.Style)
}

// HasParams reports whether path parameters are set.
func (s StyleSpec) HasParams() bool {
	return strings.TrimSpace(s.MainDir) != "" || strings.TrimSpace(s.ReplicaDir) != ""
}

// String returns a short debug form.
func (s StyleSpec) String() string {
	n := s.Name()
	if n == "" {
		return ""
	}
	if !s.HasParams() {
		return n
	}
	return fmt.Sprintf("%s(main_dir=%q,replica_dir=%q)", n,
		strings.TrimSpace(s.MainDir), strings.TrimSpace(s.ReplicaDir))
}

// ParseStyleValue converts a TOML-decoded workspace_style value (string or table map).
func ParseStyleValue(v any) (StyleSpec, error) {
	if v == nil {
		return StyleSpec{}, nil
	}
	switch t := v.(type) {
	case string:
		return StyleSpec{Style: strings.TrimSpace(t)}, nil
	case map[string]any:
		return styleFromMap(t)
	default:
		return StyleSpec{}, fmt.Errorf("workspace_style: unsupported type %T (want string or table)", v)
	}
}

func styleFromMap(m map[string]any) (StyleSpec, error) {
	get := func(keys ...string) string {
		for _, k := range keys {
			if raw, ok := m[k]; ok {
				switch x := raw.(type) {
				case string:
					return strings.TrimSpace(x)
				}
			}
		}
		return ""
	}
	style := get("style", "name")
	if style == "" {
		return StyleSpec{}, fmt.Errorf("workspace_style table requires style = \"…\"")
	}
	return StyleSpec{
		Style:      style,
		MainDir:    get("main_dir", "mainDir"),
		ReplicaDir: get("replica_dir", "replicaDir"),
	}, nil
}

// ParseStyleSpec builds a name-only spec (CLI / env).
func ParseStyleSpec(name string) StyleSpec {
	return StyleSpec{Style: strings.TrimSpace(name)}
}
