package api

import (
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
