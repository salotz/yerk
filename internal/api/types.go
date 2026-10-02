package api

// APIVersion is the version string stamped on resources in this package.
const APIVersion = "yerk/v1"

// Kind names for TypeMeta-style Kind fields.
const (
	KindCatalog       = "Catalog"
	KindProject       = "Project"
	KindProjectStatus = "ProjectStatus"
	KindReplicaStatus = "ReplicaStatus"
	KindProjectInfo   = "ProjectInfo"
	KindReplicaInfo   = "ReplicaInfo"
	KindConfigResolve = "ConfigResolve"
	KindToolContext   = "ToolContext"
	KindDirContext    = "DirContext"
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

// Project is a catalog registry resource: identity and VCS metadata.
// Host workspace paths live in tool config [[projects]] (ADR 014), not here.
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

// PlacementInfo is effective placement snapshot embedded on get/lookup (ADR 013/015).
type PlacementInfo struct {
	// Style is the effective workspace style name.
	Style string `json:"style"`
	// Bound is true when host project state supplied the winning style.
	Bound bool `json:"bound"`
	// Sources lists contributing layer labels (best-effort debug).
	Sources []string `json:"sources,omitempty"`
	// Warnings are ambient drift messages when bound state wins.
	Warnings []string `json:"warnings,omitempty"`
}

// ProjectInfo is a single-project read model for get/lookup (ADR 015).
// Richer than bare Project: host paths, presence, and effective placement.
type ProjectInfo struct {
	// APIVersion is the resource API version (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// Kind is always KindProjectInfo.
	Kind string `json:"kind"`
	// URI is the canonical project identifier (yerk://domain/name).
	URI string `json:"uri,omitempty"`
	// Name is the catalog project name.
	Name string `json:"name"`
	// Domain is the project domain namespace (ADR 012).
	Domain string `json:"domain,omitempty"`
	// Remote is the primary git remote URI (or path).
	Remote string `json:"remote,omitempty"`
	// DefaultReplica is the catalog main-replica override when set.
	DefaultReplica string `json:"defaultReplica,omitempty"`
	// Tags echoes catalog tags.
	Tags []string `json:"tags,omitempty"`
	// WorkspacePath is the absolute project workspace directory.
	WorkspacePath string `json:"workspacePath"`
	// WorkspacePresence is presence of the workspace directory.
	WorkspacePresence Presence `json:"workspacePresence,omitempty"`
	// Placement is effective style / binding for this project on this host.
	Placement *PlacementInfo `json:"placement,omitempty"`
	// MatchedPath is the absolute path that triggered lookup (lookup only).
	MatchedPath string `json:"matchedPath,omitempty"`
}

// ReplicaInfo is a single-replica read model for get/lookup (ADR 015).
type ReplicaInfo struct {
	// APIVersion is the resource API version (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// Kind is always KindReplicaInfo.
	Kind string `json:"kind"`
	// URI is the canonical replica identifier (yerk://domain/project/replica).
	URI string `json:"uri,omitempty"`
	// Project is the catalog project name.
	Project string `json:"project"`
	// Replica is the replica distinguisher.
	Replica string `json:"replica"`
	// Domain is the project domain namespace (ADR 012).
	Domain string `json:"domain,omitempty"`
	// Remote echoes the catalog remote URI.
	Remote string `json:"remote,omitempty"`
	// Tags echoes catalog tags.
	Tags []string `json:"tags,omitempty"`
	// Path is the absolute on-disk checkout path.
	Path string `json:"path"`
	// Presence is missing | present | invalid for Path.
	Presence Presence `json:"presence"`
	// WorkspacePath is the owning project workspace directory.
	WorkspacePath string `json:"workspacePath,omitempty"`
	// Placement is effective style / binding for the project on this host.
	Placement *PlacementInfo `json:"placement,omitempty"`
	// MatchedPath is the absolute path that triggered lookup (lookup only).
	MatchedPath string `json:"matchedPath,omitempty"`
}

// ConfigResolveContribution is one layer in a config resolve dump (ADR 015 follow-on / Phase 4).
type ConfigResolveContribution struct {
	Order   int    `json:"order"`
	Layer   string `json:"layer"`
	Path    string `json:"path,omitempty"`
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
	Applies bool   `json:"applies"`
	Note    string `json:"note,omitempty"`
}

// ConfigResolve is the target-scoped placement contribution report (config resolve).
type ConfigResolve struct {
	// APIVersion is the resource API version (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// Kind is always KindConfigResolve.
	Kind string `json:"kind"`
	// URI is the canonical project URI when the target is a project.
	URI string `json:"uri,omitempty"`
	// Target is the bare project id or path argument as resolved.
	Target string `json:"target,omitempty"`
	// Anchor is the dir-local walk start path.
	Anchor string `json:"anchor,omitempty"`
	// WorkspacePath is the resolved project workspace when known.
	WorkspacePath string `json:"workspacePath,omitempty"`
	// EffectiveStyle is the winning workspace style.
	EffectiveStyle string `json:"effectiveStyle"`
	// Bound is true when host project state supplied the winning style.
	Bound bool `json:"bound"`
	// Warnings are ambient drift messages.
	Warnings []string `json:"warnings,omitempty"`
	// Files are file paths considered (ordered, unique).
	Files []string `json:"files,omitempty"`
	// Contributions is the ordered layer stack.
	Contributions []ConfigResolveContribution `json:"contributions"`
}

// ProjectOverall is a rollup of replica presence/change for one project.
// Emphasizes divergence: clean only when every present replica is clean and
// none are missing/invalid relative to the collected set.
type ProjectOverall struct {
	// Presence summarizes replica roots: all-present | partial | missing | invalid | none.
	Presence string `json:"presence"`
	// Change is a bag of flags across present replicas (or "-" when not probed).
	// Prefers surfacing dirty/untracked/ahead/behind/error over a lone clean.
	Change string `json:"change,omitempty"`
	// ReplicaCount is how many replica rows were collected.
	ReplicaCount int `json:"replicaCount"`
	// PresentCount is how many of those are presence=present.
	PresentCount int `json:"presentCount"`
}

// ToolContext is a tool-wide agent dump (ADR 017): vocabulary, commands, host paths.
type ToolContext struct {
	// APIVersion is the resource API version (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// Kind is always KindToolContext.
	Kind string `json:"kind"`
	// Version is the yerk version identity string.
	Version string `json:"version,omitempty"`
	// Vocabulary is short PRJX/yerk noun definitions.
	Vocabulary []ContextTerm `json:"vocabulary,omitempty"`
	// Commands maps primary verbs to one-line roles.
	Commands []ContextCommand `json:"commands,omitempty"`
	// Paths are resolved host directories/files for this process.
	Paths ToolContextPaths `json:"paths"`
	// EnvPrimary lists primary YERK__ knobs (names only; live values via envvars).
	EnvPrimary []string `json:"envPrimary,omitempty"`
	// HowTo are short task pointers (not full docs).
	HowTo []string `json:"howTo,omitempty"`
	// Notes are stability / usage caveats for agents.
	Notes []string `json:"notes,omitempty"`
}

// ContextTerm is one vocabulary entry.
type ContextTerm struct {
	Term string `json:"term"`
	Def  string `json:"def"`
}

// ContextCommand is one command map row.
type ContextCommand struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// ToolContextPaths holds resolved XDG / config locations for this process.
type ToolContextPaths struct {
	ConfigDir   string `json:"configDir,omitempty"`
	ConfigFile  string `json:"configFile,omitempty"`
	CatalogFile string `json:"catalogFile,omitempty"`
	StateDir    string `json:"stateDir,omitempty"`
}

// DirContext is a directory-scoped agent dump (ADR 017): lookup + placement + short status.
type DirContext struct {
	// APIVersion is the resource API version (yerk/v1).
	APIVersion string `json:"apiVersion"`
	// Kind is always KindDirContext.
	Kind string `json:"kind"`
	// Path is the absolute path that was resolved (cwd or argument).
	Path string `json:"path"`
	// Matched is "project" or "replica" when under a catalog workspace.
	Matched string `json:"matched,omitempty"`
	// Project is set when the path matched a project (and always when replica matched).
	Project *ProjectInfo `json:"project,omitempty"`
	// Replica is set when the path matched under a replica root.
	Replica *ReplicaInfo `json:"replica,omitempty"`
	// Placement is a compact effective placement snapshot.
	Placement *PlacementInfo `json:"placement,omitempty"`
	// EffectiveStyle is the winning workspace style (duplicate of placement.style for scanners).
	EffectiveStyle string `json:"effectiveStyle,omitempty"`
	// Bound is true when host project state supplied the winning style.
	Bound bool `json:"bound,omitempty"`
	// Warnings are placement ambient-drift messages.
	Warnings []string `json:"warnings,omitempty"`
	// StatusOverall is the project-level presence/change rollup when collected.
	StatusOverall *ProjectOverall `json:"statusOverall,omitempty"`
	// LiveReplicas lists distinguisher names discovered on disk (when collected).
	LiveReplicas []string `json:"liveReplicas,omitempty"`
	// Notes are short operator/agent hints for this dump.
	Notes []string `json:"notes,omitempty"`
}

// ProjectStatus is a project-scoped status view (workspace + replicas + overall).
// Multi-project list UX emphasizes WorkspacePath and Overall; single-project
// detail also lists Replicas. ReplicaStatus remains the one-replica resource.
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
	// DefaultReplica summarizes the default/main replica (always included when known).
	DefaultReplica *ReplicaSummary `json:"defaultReplica,omitempty"`
	// Replicas lists live on-disk replicas plus default when missing (probed).
	Replicas []ReplicaStatus `json:"replicas,omitempty"`
	// Overall rolls up presence/change across Replicas.
	Overall *ProjectOverall `json:"overall,omitempty"`
	// Tags echoes catalog tags.
	Tags []string `json:"tags,omitempty"`
	// Remote echoes the catalog remote URI.
	Remote string `json:"remote,omitempty"`
}
