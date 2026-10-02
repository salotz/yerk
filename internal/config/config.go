// Package config loads yerk tool configuration and the project catalog.
//
// Layout (XDG):
//
//	$XDG_CONFIG_HOME/yerk/   (default ~/.config/yerk)
//	  config.toml   — tool / host behavior (style, optional domain roots, host rows)
//	  catalog.toml  — portable project registry (identity + remote + tags)
//	$XDG_STATE_HOME/yerk/    (default ~/.local/state/yerk) — project bindings (ADR 013)
//
// Environment:
//
//	YERK__CONFIG       explicit tool config file
//	YERK__CONFIG_DIR   alternate config directory
//	YERK__CATALOG      explicit catalog file
//	YERK__STATE_DIR    alternate state directory (project bindings)
//	YERK__WORKSPACE_STYLE  ambient style overlay (workspace-dir | project-dir)
//	PRJX__*            PRJX spec concerns (discovered, not owned here)
//
// Missing files are valid: empty catalog + tool defaults.
// Style merge and dir-local layers live in package placement (ADR 013).
// Path model: ADR 014 — optional [domains] defaults + optional [[projects]] path.
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

// Config is tool / host behavior (not the portable catalog).
type Config struct {
	Workspace Workspace `toml:"workspace"`
	// Domains maps logical domain names to host-absolute roots (ADR 014).
	// Optional. When set, catalog projects in that domain default to
	// <root>/<name> unless a [[projects]] path overrides.
	Domains map[string]string `toml:"domains"`
	// Projects are optional host-local placement rows keyed by domain+name.
	// Use for absolute/~/ overrides, relative path under the domain root,
	// and later host knobs (workspace_style, etc.).
	Projects []HostProject `toml:"projects"`
}

// Workspace holds host-default placement policy knobs (ambient layer).
// Effective style is merged in package placement (ADR 013).
type Workspace struct {
	// Style how replicas sit relative to each project's workspace path:
	//   - "workspace-dir": <workspace>/<replica>
	//   - "project-dir":   <dir(workspace)>/<name>__<replica>
	Style string `toml:"style"`
}

// HostProject is one optional host-local placement row in config.toml.
//
//	[[projects]]
//	name = "yerk"
//	domain = "personal"
//	# path omitted → <domains.personal>/<name>
//	# path = "other/yerk"          # relative → join domain root
//	# path = "~/tree/odd/yerk"     # absolute or ~/ override (domain root optional)
//	# workspace_style = "project-dir"
type HostProject struct {
	Name   string `toml:"name"`
	Domain string `toml:"domain"`
	// Path is optional. Absolute or ~/… overrides placement; relative joins
	// the domain root; empty uses default <domain-root>/<name>.
	Path string `toml:"path,omitempty"`
	// WorkspaceStyle is an optional host-row ambient style override (ADR 013).
	WorkspaceStyle string `toml:"workspace_style,omitempty"`
}

// ID returns the bare project identifier domain/name.
func (p HostProject) ID() string {
	return strings.TrimSpace(p.Domain) + "/" + strings.TrimSpace(p.Name)
}

// Defaults returns usable empty tool configuration.
func Defaults() Config {
	return Config{
		Workspace: Workspace{
			Style: "workspace-dir",
		},
		Domains:  nil,
		Projects: nil,
	}
}

// FindHostProject returns the host row for domain/name, or false if missing.
func (c Config) FindHostProject(domain, name string) (HostProject, bool) {
	domain = strings.TrimSpace(domain)
	name = strings.TrimSpace(name)
	for _, p := range c.Projects {
		if strings.TrimSpace(p.Domain) == domain && strings.TrimSpace(p.Name) == name {
			return p, true
		}
	}
	return HostProject{}, false
}

// DomainRoot returns the expanded absolute root for a domain name.
// ok is false when the domain is not configured (not an error).
func (c Config) DomainRoot(domain string) (root string, ok bool, err error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return "", false, fmt.Errorf("empty domain name")
	}
	if c.Domains == nil {
		return "", false, nil
	}
	raw, found := c.Domains[domain]
	if !found {
		return "", false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, fmt.Errorf("config [domains] %q: empty path", domain)
	}
	expanded, err := ExpandUser(raw)
	if err != nil {
		return "", false, fmt.Errorf("config [domains] %q: %w", domain, err)
	}
	if !filepath.IsAbs(expanded) {
		return "", false, fmt.Errorf("config [domains] %q must be absolute or ~/… (got %q)", domain, raw)
	}
	return filepath.Clean(expanded), true, nil
}

// ProjectWorkspacePath returns the absolute project workspace for domain/name.
//
// Resolution (ADR 014):
//  1. Host [[projects]] path absolute or ~/… → that path
//  2. Host [[projects]] path relative → <domains[domain]>/<path>
//  3. No path (missing row or empty path) → <domains[domain]>/<name>
//  4. Otherwise error: configure [domains.<domain>] or a full [[projects]] path
func (c Config) ProjectWorkspacePath(domain, name string) (string, error) {
	domain = strings.TrimSpace(domain)
	name = strings.TrimSpace(name)
	if domain == "" || name == "" {
		return "", fmt.Errorf("project workspace path requires domain and name")
	}
	key := domain + "/" + name

	var rawPath string
	if hp, ok := c.FindHostProject(domain, name); ok {
		rawPath = strings.TrimSpace(hp.Path)
	}

	// Explicit absolute / ~/ host path wins (no domain root required).
	if rawPath != "" {
		expanded, err := ExpandUser(rawPath)
		if err != nil {
			return "", fmt.Errorf("config [[projects]] %q path: %w", key, err)
		}
		if filepath.IsAbs(expanded) {
			return filepath.Clean(expanded), nil
		}
		// Relative host path: join domain root.
		root, ok, err := c.DomainRoot(domain)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("config [[projects]] %q path %q is relative but [domains.%s] is not set; use an absolute or ~/ path, or set the domain root", key, rawPath, domain)
		}
		return filepath.Clean(filepath.Join(root, expanded)), nil
	}

	// Default: domain root + project name (no host path required).
	root, ok, err := c.DomainRoot(domain)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("no workspace path for %q: set [domains.%s] (default <root>/%s) or a config [[projects]] row with an absolute or ~/ path", key, domain, name)
	}
	return filepath.Clean(filepath.Join(root, name)), nil
}

// ValidateHostProjects checks host project rows and domain map keys.
// Path expansion/abs checks run at resolve time so missing home is still loadable.
func (c Config) ValidateHostProjects() error {
	for domain, raw := range c.Domains {
		d := strings.TrimSpace(domain)
		if d == "" {
			return fmt.Errorf("config [domains]: empty domain key")
		}
		if strings.ContainsAny(d, `/\`) {
			return fmt.Errorf("config [domains] %q: domain must be a single segment", d)
		}
		if strings.TrimSpace(raw) == "" {
			return fmt.Errorf("config [domains] %q: empty path", d)
		}
	}

	seen := make(map[string]struct{}, len(c.Projects))
	for i, p := range c.Projects {
		name := strings.TrimSpace(p.Name)
		domain := strings.TrimSpace(p.Domain)
		label := name
		if domain != "" && name != "" {
			label = domain + "/" + name
		}
		if label == "" {
			label = fmt.Sprintf("projects[%d]", i)
		}
		if name == "" {
			return fmt.Errorf("config [[projects]] %s: empty name", label)
		}
		if domain == "" {
			return fmt.Errorf("config [[projects]] %q: empty domain", name)
		}
		if strings.ContainsAny(name, `/\`) || strings.ContainsAny(domain, `/\`) {
			return fmt.Errorf("config [[projects]] %q: name and domain must be single segments", label)
		}
		if _, dup := seen[label]; dup {
			return fmt.Errorf("config [[projects]]: duplicate id %q", label)
		}
		seen[label] = struct{}{}
		// path may be empty (default to domain root + name)
	}
	return nil
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
	if err := cfg.ValidateHostProjects(); err != nil {
		return cfg, fmt.Errorf("config %s: %w", path, err)
	}
	return applyEnv(cfg), nil
}

func applyEnv(cfg Config) Config {
	if v := os.Getenv("YERK__WORKSPACE_STYLE"); v != "" {
		cfg.Workspace.Style = v
	}
	return cfg
}
