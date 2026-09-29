package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Catalog is the host-global project registry.
type Catalog struct {
	// Tags is the closed vocabulary of labels projects may use (ADR 010).
	// Declared at the top of catalog.toml; project tags must be members.
	Tags []string `toml:"tags,omitempty"`
	// Projects is the registered project list.
	Projects []Project `toml:"projects"`
}

// Project is one registered software project.
type Project struct {
	// Name is the short project id used in CLI queries (catalog key).
	Name string `toml:"name"`
	// Domain is the namespace used to select a host domain root in config.toml
	// when Path is relative (ADR 008). Also a display label.
	Domain string `toml:"domain,omitempty"`
	// Remote is the primary clone URI (git URL or path). MVP: one remote.
	Remote string `toml:"remote"`
	// Path is the project workspace directory (owns replicas).
	// Prefer a path relative to the domain root (portable catalog), e.g.
	// "devel/yerk" with domain "personal" → <domains.personal>/devel/yerk.
	// Absolute paths are allowed as a host-local escape hatch.
	// Example (workspace-dir): workspace …/devel/yerk → replica main at …/yerk/main.
	Path string `toml:"path,omitempty"`
	// DefaultReplica overrides remote HEAD branch short name when set.
	DefaultReplica string `toml:"default_replica,omitempty"`
	// Tags group projects for bulk operations. Each entry must appear in
	// Catalog.Tags (closed vocabulary).
	Tags []string `toml:"tags,omitempty"`
}

// EmptyCatalog is a catalog with no projects.
func EmptyCatalog() Catalog {
	return Catalog{Projects: nil}
}

// LoadCatalog reads catalog.toml if present, else returns an empty catalog.
// Non-empty catalogs are validated (declared tags; project tags ⊆ declared).
func LoadCatalog() (Catalog, error) {
	path, err := CatalogPath()
	if err != nil {
		return EmptyCatalog(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return EmptyCatalog(), nil
		}
		return EmptyCatalog(), fmt.Errorf("read catalog %s: %w", path, err)
	}

	var cat Catalog
	if err := toml.Unmarshal(data, &cat); err != nil {
		return EmptyCatalog(), fmt.Errorf("parse catalog %s: %w", path, err)
	}
	if err := cat.Validate(); err != nil {
		return EmptyCatalog(), fmt.Errorf("catalog %s: %w", path, err)
	}
	return cat, nil
}

// Validate checks the closed tag vocabulary and project rows.
// Empty catalogs (no projects) are valid even with no declared tags.
func (c Catalog) Validate() error {
	allowed, err := tagSet(c.Tags, "catalog tags")
	if err != nil {
		return err
	}
	for _, p := range c.Projects {
		name := p.Name
		if strings.TrimSpace(name) == "" {
			name = "(unnamed)"
		}
		seen := make(map[string]struct{}, len(p.Tags))
		for _, raw := range p.Tags {
			t := strings.TrimSpace(raw)
			if t == "" {
				return fmt.Errorf("project %q: empty tag name", name)
			}
			if t != raw {
				return fmt.Errorf("project %q: tag %q has leading/trailing space", name, raw)
			}
			if _, ok := allowed[t]; !ok {
				return fmt.Errorf("project %q: tag %q not declared in catalog tags", name, t)
			}
			if _, dup := seen[t]; dup {
				return fmt.Errorf("project %q: duplicate tag %q", name, t)
			}
			seen[t] = struct{}{}
		}
	}
	return nil
}

// tagSet builds a set from a declared tag list; rejects empty and duplicate names.
func tagSet(tags []string, what string) (map[string]struct{}, error) {
	out := make(map[string]struct{}, len(tags))
	for _, raw := range tags {
		t := strings.TrimSpace(raw)
		if t == "" {
			return nil, fmt.Errorf("%s: empty tag name", what)
		}
		if t != raw {
			return nil, fmt.Errorf("%s: tag %q has leading/trailing space", what, raw)
		}
		if _, dup := out[t]; dup {
			return nil, fmt.Errorf("%s: duplicate %q", what, t)
		}
		out[t] = struct{}{}
	}
	return out, nil
}

// HasTag reports whether tag is in the catalog's declared vocabulary.
func (c Catalog) HasTag(tag string) bool {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return false
	}
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// Find returns the project with the given name, or false if missing.
func (c Catalog) Find(name string) (Project, bool) {
	name = strings.TrimSpace(name)
	for _, p := range c.Projects {
		if p.Name == name {
			return p, true
		}
	}
	return Project{}, false
}

// FilterTag returns projects that include tag. Empty tag returns all.
// Does not validate that tag is declared; prefer SelectByTag for CLI filters.
func (c Catalog) FilterTag(tag string) []Project {
	if tag == "" {
		out := make([]Project, len(c.Projects))
		copy(out, c.Projects)
		return out
	}
	var out []Project
	for _, p := range c.Projects {
		for _, t := range p.Tags {
			if t == tag {
				out = append(out, p)
				break
			}
		}
	}
	return out
}

// SelectByTag returns projects that list tag. Tag must be non-empty and
// declared in the catalog vocabulary (ADR 010). An empty match set is not
// an error (declared tag with no projects yet).
func (c Catalog) SelectByTag(tag string) ([]Project, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil, fmt.Errorf("tag required")
	}
	if !c.HasTag(tag) {
		return nil, fmt.Errorf("unknown tag %q (not in catalog tags)", tag)
	}
	return c.FilterTag(tag), nil
}
