// Package config loads yerk tool configuration and the project catalog.
//
// Layout (XDG):
//
//	$XDG_CONFIG_HOME/yerk/   (default ~/.config/yerk)
//	  config.toml   — tool / host behavior (workspace style, domain roots)
//	  catalog.toml  — project registry (portable relative paths under domains)
//
// Environment:
//
//	YERK__CONFIG       explicit tool config file
//	YERK__CONFIG_DIR   alternate config directory
//	YERK__CATALOG      explicit catalog file
//	YERK__WORKSPACE_STYLE  optional style overlay (workspace-dir | project-dir)
//	PRJX__*            PRJX spec concerns (discovered, not owned here)
//
// Missing files are valid: empty catalog + tool defaults.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	// AppName is the binary and config directory name.
	AppName = "yerk"
	// EnvPrefix is the tool env var prefix (YERK__).
	EnvPrefix = "YERK__"
	// ConfigFileName is the tool config basename.
	ConfigFileName = "config.toml"
	// CatalogFileName is the project catalog basename.
	CatalogFileName = "catalog.toml"
)

// Config is tool / host behavior (not the project list).
type Config struct {
	Workspace Workspace         `toml:"workspace"`
	// Domains maps domain name → absolute host root for that domain.
	// Catalog project paths that are relative join this root (ADR 008).
	Domains map[string]string `toml:"domains"`
}

// Workspace holds placement policy knobs.
//
// Each catalog project resolves to an absolute project-workspace directory
// (domain root + relative path, or an absolute path). Style maps a replica
// distinguisher under that workspace.
type Workspace struct {
	// Style how replicas sit relative to each project's workspace path:
	//   - "workspace-dir": <workspace>/<replica>
	//   - "project-dir":   <dir(workspace)>/<name>__<replica>
	Style string `toml:"style"`
}

// Defaults returns usable empty tool configuration.
func Defaults() Config {
	return Config{
		Workspace: Workspace{
			Style: "workspace-dir",
		},
		Domains: nil,
	}
}

// DomainRoot returns the configured absolute root for domain, or false.
// Leading "~" / "~/" in the configured value is expanded to the user home.
func (c Config) DomainRoot(domain string) (string, bool, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" || c.Domains == nil {
		return "", false, nil
	}
	raw, ok := c.Domains[domain]
	if !ok {
		return "", false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, nil
	}
	abs, err := ExpandUser(raw)
	if err != nil {
		return "", true, fmt.Errorf("domain %q root: %w", domain, err)
	}
	if !filepath.IsAbs(abs) {
		return "", true, fmt.Errorf("domain %q root must be absolute (after ~ expansion), got %q", domain, abs)
	}
	return abs, true, nil
}

// ExpandUser expands a leading "~" or "~/" to the process user's home directory.
// Other paths are cleaned and returned unchanged (relative stays relative).
func ExpandUser(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", nil
	}
	if p == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand ~: %w", err)
		}
		return home, nil
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand ~: %w", err)
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
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

// FilePath returns the primary tool config file path (config.toml).
func FilePath() (string, error) {
	if v := os.Getenv("YERK__CONFIG"); v != "" {
		return v, nil
	}
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// CatalogPath returns the catalog file path (catalog.toml).
func CatalogPath() (string, error) {
	if v := os.Getenv("YERK__CATALOG"); v != "" {
		return v, nil
	}
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, CatalogFileName), nil
}

// Load reads tool config.toml if present, else returns Defaults.
func Load() (Config, error) {
	cfg := Defaults()

	path, err := FilePath()
	if err != nil {
		return cfg, err
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
	if cfg.Domains == nil {
		cfg.Domains = nil
	}
	return applyEnv(cfg), nil
}

func applyEnv(cfg Config) Config {
	if v := os.Getenv("YERK__WORKSPACE_STYLE"); v != "" {
		cfg.Workspace.Style = v
	}
	return cfg
}

