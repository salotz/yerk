package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/salotz/yerk/internal/id"
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
	// Name is the short project name (within Domain). Full bare id is domain/name.
	Name string `toml:"name"`
	// Domain is the required logical namespace for ids/URIs (ADR 012).
	// Not a filesystem root (ADR 014).
	Domain string `toml:"domain"`
	// Remote is the primary git remote URI (or path). MVP: one remote.
	Remote string `toml:"remote"`
	// DefaultReplica overrides remote HEAD branch short name when set (main replica).
	DefaultReplica string `toml:"default_replica,omitempty"`
	// WorkspaceStyle is an optional per-project ambient style override (ADR 013).
	WorkspaceStyle string `toml:"workspace_style,omitempty"`
	// ReplicaMethod is optional (worktree|clone); consumed by replica create later.
	ReplicaMethod string `toml:"replica_method,omitempty"`
	// Tags group projects for bulk operations. Each entry must appear in
	// Catalog.Tags (closed vocabulary).
	Tags []string `toml:"tags,omitempty"`
}

// ID returns the bare project identifier domain/name.
func (p Project) ID() string {
	return id.Ref{Domain: strings.TrimSpace(p.Domain), Project: strings.TrimSpace(p.Name)}.BareProject()
}

// URI returns the canonical project URI yerk://domain/name.
func (p Project) URI() string {
	return id.ProjectURI(strings.TrimSpace(p.Domain), strings.TrimSpace(p.Name))
}

// EmptyCatalog is a catalog with no projects.
func EmptyCatalog() Catalog {
	return Catalog{Projects: nil}
}

// LoadCatalog reads catalog.toml if present, else returns an empty catalog.
// Non-empty catalogs are validated (declared tags; project tags ⊆ declared).
// Legacy project `path` fields are rejected (host paths belong in config.toml
// [[projects]] — ADR 014).
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

	if err := rejectLegacyProjectPaths(data); err != nil {
		return EmptyCatalog(), fmt.Errorf("catalog %s: %w", path, err)
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

// rejectLegacyProjectPaths fails if any catalog [[projects]] row still sets path=
// (host workspace location moved to config.toml [[projects]], ADR 014).
func rejectLegacyProjectPaths(data []byte) error {
	var raw struct {
		Projects []map[string]any `toml:"projects"`
	}
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil // parse errors handled by main unmarshal
	}
	for i, row := range raw.Projects {
		if _, ok := row["path"]; !ok {
			continue
		}
		name, _ := row["name"].(string)
		domain, _ := row["domain"].(string)
		label := strings.TrimSpace(name)
		if d := strings.TrimSpace(domain); d != "" && label != "" {
			label = d + "/" + label
		}
		if label == "" {
			label = fmt.Sprintf("projects[%d]", i)
		}
		return fmt.Errorf("project %q: catalog path is not supported (ADR 014); set config.toml [domains] and/or [[projects]] path, then remove path from the catalog row", label)
	}
	return nil
}

// Validate checks identity fields, closed tag vocabulary, and project rows.
// Empty catalogs (no projects) are valid even with no declared tags.
func (c Catalog) Validate() error {
	allowed, err := tagSet(c.Tags, "catalog tags")
	if err != nil {
		return err
	}
	seenID := make(map[string]struct{}, len(c.Projects))
	for _, p := range c.Projects {
		name := strings.TrimSpace(p.Name)
		domain := strings.TrimSpace(p.Domain)
		label := name
		if label == "" {
			label = "(unnamed)"
		}
		if name == "" {
			return fmt.Errorf("project %q: empty name", label)
		}
		if domain == "" {
			return fmt.Errorf("project %q: domain is required (ADR 012)", name)
		}
		if err := validateIdentitySegment("domain", domain); err != nil {
			return fmt.Errorf("project %q: %w", name, err)
		}
		if err := validateIdentitySegment("name", name); err != nil {
			return fmt.Errorf("project %q: %w", name, err)
		}
		pid := domain + "/" + name
		if _, dup := seenID[pid]; dup {
			return fmt.Errorf("duplicate project id %q", pid)
		}
		seenID[pid] = struct{}{}

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

func validateIdentitySegment(what, s string) error {
	ref, err := id.Parse(s)
	if err != nil || ref.Domain != "" || ref.Replica != "" || ref.Project != s {
		return fmt.Errorf("invalid %s %q (use letters, digits, '-', '_', '.')", what, s)
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

// Find returns the project with the given short name, or false if missing.
// If multiple domains share the name, the first row wins — prefer Resolve
// for CLI/user input (ADR 012 ambiguity errors).
func (c Catalog) Find(name string) (Project, bool) {
	name = strings.TrimSpace(name)
	for _, p := range c.Projects {
		if p.Name == name {
			return p, true
		}
	}
	return Project{}, false
}

// FindID returns the project with bare id domain/name (or URI-equivalent pair).
func (c Catalog) FindID(domain, name string) (Project, bool) {
	domain = strings.TrimSpace(domain)
	name = strings.TrimSpace(name)
	for _, p := range c.Projects {
		if p.Domain == domain && p.Name == name {
			return p, true
		}
	}
	return Project{}, false
}

// ProjectKeys returns identity keys for id.Expand.
func (c Catalog) ProjectKeys() []id.ProjectKey {
	out := make([]id.ProjectKey, 0, len(c.Projects))
	for _, p := range c.Projects {
		out = append(out, id.ProjectKey{
			Domain: strings.TrimSpace(p.Domain),
			Name:   strings.TrimSpace(p.Name),
		})
	}
	return out
}

// Resolve expands a bare id, short name, or yerk:// URI to a catalog project
// and parsed ref. The ref may include a replica segment from the input.
func (c Catalog) Resolve(input string) (Project, id.Ref, error) {
	ref, err := id.Expand(input, c.ProjectKeys())
	if err != nil {
		return Project{}, id.Ref{}, err
	}
	p, ok := c.FindID(ref.Domain, ref.Project)
	if !ok {
		return Project{}, id.Ref{}, fmt.Errorf("project %q not in catalog", ref.BareProject())
	}
	return p, ref, nil
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
