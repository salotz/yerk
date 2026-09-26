// Package config loads yerk tool configuration.
//
// Layout (XDG):
//
//	$XDG_CONFIG_HOME/yerk/   (default ~/.config/yerk)
//
// Environment:
//
//	YERK__*  tool-specific knobs (double underscore namespace)
//	PRJX__*  PRJX spec concerns (discovered, not owned here)
//
// A missing config file is not an error: Load returns Defaults().
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const (
	// AppName is the binary and config directory name.
	AppName = "yerk"
	// EnvPrefix is the tool env var prefix (YERK__).
	EnvPrefix = "YERK__"
)

// Config is the root tool configuration document.
type Config struct {
	// Workspace describes how project replicas are laid out on the host.
	Workspace Workspace `toml:"workspace"`
	// Projects is the registered project catalog (MVP: name, remote, tags).
	Projects []Project `toml:"projects"`
}

// Workspace controls checkout path conventions.
type Workspace struct {
	// Root is the host tree root for managed projects.
	// Empty means "resolve later from PRJX / host layout".
	Root string `toml:"root"`
	// Style is the layout style for replicas.
	// Supported MVP styles:
	//   - "workspace-dir":   projects/<project>/<replica>
	//   - "project-dir":     projects/<project>__<replica>
	Style string `toml:"style"`
}

// Project is one registered software project.
type Project struct {
	// Name is the short project id used in CLI queries.
	Name string `toml:"name"`
	// Remote is the clone URI (git URL or path).
	Remote string `toml:"remote"`
	// Path is an optional fixed destination; empty uses Workspace rules.
	Path string `toml:"path,omitempty"`
	// Tags group projects for bulk operations (domain, topic, …).
	Tags []string `toml:"tags,omitempty"`
}

// Defaults returns a usable empty configuration.
func Defaults() Config {
	return Config{
		Workspace: Workspace{
			Style: "workspace-dir",
		},
		Projects: nil,
	}
}

// Dir returns the yerk config directory ($XDG_CONFIG_HOME/yerk).
func Dir() (string, error) {
	if v := os.Getenv("YERK__CONFIG_DIR"); v != "" {
		return v, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config home: %w", err)
	}
	return filepath.Join(base, AppName), nil
}

// FilePath returns the primary config file path (config.toml).
func FilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads config.toml if present, else returns Defaults.
// YERK__CONFIG may point at an explicit file path.
func Load() (Config, error) {
	cfg := Defaults()

	path := os.Getenv("YERK__CONFIG")
	if path == "" {
		var err error
		path, err = FilePath()
		if err != nil {
			return cfg, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return applyEnv(cfg), nil
		}
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	return applyEnv(cfg), nil
}

// applyEnv overlays a few well-known YERK__ overrides.
func applyEnv(cfg Config) Config {
	if v := os.Getenv("YERK__WORKSPACE_ROOT"); v != "" {
		cfg.Workspace.Root = v
	}
	if v := os.Getenv("YERK__WORKSPACE_STYLE"); v != "" {
		cfg.Workspace.Style = v
	}
	return cfg
}

// ExampleTOML is a starter config document for docs and `yerk init` later.
const ExampleTOML = `# yerk host project catalog
# Path: ~/.config/yerk/config.toml
# Override file: YERK__CONFIG=/path/to/config.toml
# Override dir:  YERK__CONFIG_DIR=/path/to/dir

[workspace]
# Host root under which projects are checked out (optional; PRJX/host layout may supply this).
# root = "/home/you/tree"
# MVP styles: "workspace-dir" (project/replica) | "project-dir" (project__replica)
style = "workspace-dir"

# [[projects]]
# name = "example"
# remote = "git@github.com:example/example.git"
# tags = ["personal", "devel"]
`
