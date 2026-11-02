package api

import (
	"sort"
	"strings"

	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/id"
	"github.com/salotz/yerk/internal/presence"
)

// ProjectFromConfig maps a catalog load struct to a Project resource.
func ProjectFromConfig(p config.Project) Project {
	return Project{
		APIVersion:     APIVersion,
		Kind:           KindProject,
		URI:            id.ProjectURI(p.Domain, p.Name),
		Name:           p.Name,
		Domain:         p.Domain,
		Remote:         p.Remote,
		DefaultReplica: p.DefaultReplica,
		Tags:           copyStrings(p.Tags),
	}
}

// CatalogFromConfig maps a loaded catalog to a Catalog resource.
func CatalogFromConfig(c config.Catalog) Catalog {
	projects := make([]Project, 0, len(c.Projects))
	for _, p := range c.Projects {
		projects = append(projects, ProjectFromConfig(p))
	}
	return Catalog{
		APIVersion: APIVersion,
		Kind:       KindCatalog,
		Tags:       copyStrings(c.Tags),
		Projects:   projects,
	}
}

// ToolConfigFromConfig maps loaded tool config plus resolved paths to ToolConfig.
func ToolConfigFromConfig(c config.Config, file, stateDir string) ToolConfig {
	out := NewToolConfig()
	out.File = strings.TrimSpace(file)
	out.Workspace = ToolConfigWorkspace{Style: strings.TrimSpace(c.Workspace.Style)}
	out.StateDir = strings.TrimSpace(stateDir)

	if len(c.Domains) > 0 {
		names := make([]string, 0, len(c.Domains))
		for d := range c.Domains {
			names = append(names, d)
		}
		sort.Strings(names)
		out.Domains = make([]ToolConfigDomain, 0, len(names))
		for _, d := range names {
			raw := strings.TrimSpace(c.Domains[d])
			row := ToolConfigDomain{Name: d, Path: raw}
			if abs, err := config.ExpandUser(raw); err == nil && abs != "" && abs != raw {
				row.PathExpanded = abs
			}
			out.Domains = append(out.Domains, row)
		}
	}

	if len(c.Projects) > 0 {
		rows := make([]ToolConfigProject, 0, len(c.Projects))
		for _, hp := range c.Projects {
			row := ToolConfigProject{
				Name:           strings.TrimSpace(hp.Name),
				Domain:         strings.TrimSpace(hp.Domain),
				Path:           strings.TrimSpace(hp.Path),
				WorkspaceStyle: strings.TrimSpace(hp.WorkspaceStyle),
			}
			if row.Domain != "" && row.Name != "" {
				row.URI = id.ProjectURI(row.Domain, row.Name)
			}
			if row.Path != "" {
				if abs, err := config.ExpandUser(row.Path); err == nil && abs != "" && abs != row.Path {
					row.PathExpanded = abs
				}
			}
			if !hp.Style.IsZero() {
				if row.WorkspaceStyle == "" {
					row.WorkspaceStyle = hp.Style.Name()
				}
				if hp.Style.HasParams() {
					spec := StyleSpec{
						Style:      hp.Style.Name(),
						MainDir:    strings.TrimSpace(hp.Style.MainDir),
						ReplicaDir: strings.TrimSpace(hp.Style.ReplicaDir),
					}
					row.Style = &spec
				}
			}
			rows = append(rows, row)
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].URI < rows[j].URI })
		out.Projects = rows
	}
	return out
}

// PresenceFrom converts a presence classifier value to api.Presence.
func PresenceFrom(s presence.Status) Presence {
	switch s {
	case presence.Missing:
		return PresenceMissing
	case presence.Present:
		return PresencePresent
	case presence.Invalid:
		return PresenceInvalid
	default:
		// Preserve unknown strings rather than inventing a value.
		return Presence(s)
	}
}

// NewReplicaStatus returns a ReplicaStatus with TypeMeta fields set.
func NewReplicaStatus() ReplicaStatus {
	return ReplicaStatus{
		APIVersion: APIVersion,
		Kind:       KindReplicaStatus,
		Change:     "-",
		Branch:     "-",
	}
}

// NewProjectStatus returns a ProjectStatus with TypeMeta fields set.
func NewProjectStatus() ProjectStatus {
	return ProjectStatus{
		APIVersion: APIVersion,
		Kind:       KindProjectStatus,
	}
}

// NewProjectInfo returns a ProjectInfo with TypeMeta fields set.
func NewProjectInfo() ProjectInfo {
	return ProjectInfo{
		APIVersion: APIVersion,
		Kind:       KindProjectInfo,
	}
}

// NewReplicaInfo returns a ReplicaInfo with TypeMeta fields set.
func NewReplicaInfo() ReplicaInfo {
	return ReplicaInfo{
		APIVersion: APIVersion,
		Kind:       KindReplicaInfo,
	}
}

// NewConfigResolve returns a ConfigResolve with TypeMeta fields set.
func NewConfigResolve() ConfigResolve {
	return ConfigResolve{
		APIVersion: APIVersion,
		Kind:       KindConfigResolve,
	}
}

// NewToolConfig returns a ToolConfig with TypeMeta fields set.
func NewToolConfig() ToolConfig {
	return ToolConfig{
		APIVersion: APIVersion,
		Kind:       KindToolConfig,
	}
}

// NewEnvVars returns an EnvVars resource with TypeMeta fields set.
func NewEnvVars() EnvVars {
	return EnvVars{
		APIVersion: APIVersion,
		Kind:       KindEnvVars,
	}
}

// NewToolContext returns a ToolContext with TypeMeta fields set.
func NewToolContext() ToolContext {
	return ToolContext{
		APIVersion: APIVersion,
		Kind:       KindToolContext,
	}
}

// NewDirContext returns a DirContext with TypeMeta fields set.
func NewDirContext() DirContext {
	return DirContext{
		APIVersion: APIVersion,
		Kind:       KindDirContext,
	}
}

// Summary returns a compact ReplicaSummary from a full ReplicaStatus.
func (r ReplicaStatus) Summary() ReplicaSummary {
	return ReplicaSummary{
		Name:     r.Replica,
		Path:     r.Path,
		Presence: r.Presence,
		Change:   r.Change,
		Branch:   r.Branch,
	}
}

func copyStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
