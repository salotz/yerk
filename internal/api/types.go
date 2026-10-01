package api

// APIVersion is the version string stamped on resources in this package.
const APIVersion = "yerk/v1"

// Kind names for TypeMeta-style Kind fields.
const (
	KindCatalog       = "Catalog"
	KindProject       = "Project"
	KindProjectStatus = "ProjectStatus"
	KindReplicaStatus = "ReplicaStatus"
)

// Presence is on-disk presence of an expected path (replica checkout or,
// for project workspace dirs, existence/usability as laid out later).
type Presence string

const (
	// PresenceMissing: expected path does not exist.
	PresenceMissing Presence = "missing"
	// PresencePresent: path exists and is usable for its role (git checkout
	// for replicas).
	PresencePresent Presence = "present"
	// PresenceInvalid: path exists but is not usable for its role.
	PresenceInvalid Presence = "invalid"
)

// Project is a catalog registry resource: identity and desired placement.
// Path is the catalog value (relative or absolute), not necessarily resolved.
type Project struct {
	// APIVersion is the resource API version (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// Kind is always KindProject.
	Kind string `json:"kind"`
	// URI is the canonical project identifier (yerk://domain/name). ADR 012.
	URI string `json:"uri,omitempty"`
	// Name is the catalog project name (short; full id is domain/name).
	Name string `json:"name"`
	// Domain is the required logical namespace for ids/URIs (ADR 012).
	Domain string `json:"domain,omitempty"`
	// Remote is the primary git remote URI (or path).
	Remote string `json:"remote,omitempty"`
	// Path is the catalog project-workspace path (relative or absolute).
	Path string `json:"path,omitempty"`
	// DefaultReplica overrides remote HEAD when set (main replica).
	DefaultReplica string `json:"defaultReplica,omitempty"`
	// Tags are bulk-select labels (members of Catalog.Tags).
	Tags []string `json:"tags,omitempty"`
}

// Catalog is the host registry: closed tag vocabulary plus projects.
type Catalog struct {
	// APIVersion is the resource API version (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// Kind is always KindCatalog.
	Kind string `json:"kind"`
	// Tags is the closed vocabulary of allowed project tags.
	Tags []string `json:"tags,omitempty"`
	// Projects is the registered project list.
	Projects []Project `json:"projects"`
}

// ReplicaStatus is observed state for one replica checkout on this host.
type ReplicaStatus struct {
	// APIVersion is the resource API version (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// Kind is always KindReplicaStatus.
	Kind string `json:"kind"`
	// URI is the canonical replica identifier (yerk://domain/project/replica).
	URI string `json:"uri,omitempty"`
	// Project is the catalog project name.
	Project string `json:"project"`
	// Replica is the replica distinguisher (often default branch short name).
	Replica string `json:"replica"`
	// Domain is the project domain namespace (ADR 012).
	Domain string `json:"domain,omitempty"`
	// Path is the absolute on-disk checkout path.
	Path string `json:"path"`
	// Presence is missing | present | invalid for Path.
	Presence Presence `json:"presence"`
	// Change is space-separated change flags, "-" when not probed, or "error".
	Change string `json:"change,omitempty"`
	// Branch is the checked-out branch short name when known; "-" if N/A.
	Branch string `json:"branch,omitempty"`
	// Tags echoes catalog tags for filtering/display.
	Tags []string `json:"tags,omitempty"`
	// Remote echoes the catalog remote URI.
	Remote string `json:"remote,omitempty"`
}

// ReplicaSummary is a compact default-replica snapshot embedded on ProjectStatus.
type ReplicaSummary struct {
	// Name is the replica distinguisher.
	Name string `json:"name"`
	// Path is the absolute checkout path.
	Path string `json:"path"`
	// Presence is missing | present | invalid.
	Presence Presence `json:"presence"`
	// Change is space-separated flags or "-" / empty when not probed.
	Change string `json:"change,omitempty"`
	// Branch is the checked-out branch when known.
	Branch string `json:"branch,omitempty"`
}

// ProjectStatus is a project-scoped status view (workspace + optional replica summary).
// Default list UX (P6) should emphasize WorkspacePath; ReplicaStatus owns
// full replica-scoped detail.
type ProjectStatus struct {
	// APIVersion is the resource API version (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// Kind is always KindProjectStatus.
	Kind string `json:"kind"`
	// URI is the canonical project identifier (yerk://domain/name).
	URI string `json:"uri,omitempty"`
	// Name is the catalog project name.
	Name string `json:"name"`
	// Domain is the project domain namespace (ADR 012).
	Domain string `json:"domain,omitempty"`
	// WorkspacePath is the absolute project workspace directory (owns replicas).
	WorkspacePath string `json:"workspacePath"`
	// WorkspacePresence is optional presence of the workspace directory itself
	// (not a git checkout classifier). Empty when not evaluated.
	WorkspacePresence Presence `json:"workspacePresence,omitempty"`
	// DefaultReplica summarizes the default replica when collected.
	DefaultReplica *ReplicaSummary `json:"defaultReplica,omitempty"`
	// Tags echoes catalog tags.
	Tags []string `json:"tags,omitempty"`
	// Remote echoes the catalog remote URI.
	Remote string `json:"remote,omitempty"`
}
