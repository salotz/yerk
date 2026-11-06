package models

// Catalog registry resource: identity and version-control metadata.
// Host workspace paths live in tool config [[projects]] (ADR 014), not here.
#Project: {
	#ResourceMeta
	kind: "Project"

	// Canonical project identifier (yerk://domain/name). ADR 012.
	uri?: #ProjectURI

	// Catalog project name (short; full identifier is domain/name).
	name: #ProjectName

	// Required logical namespace for IDs/URIs (ADR 012).
	domain?: #ProjectDomain

	// Primary git remote URI (or path).
	remote?: #GitRemote

	// Override for remote HEAD when set (main replica).
	defaultReplica?: #ReplicaName

	// Bulk-select labels (members of Catalog.tags).
	tags?: [...#Tag]
}

// Host registry: closed tag vocabulary plus projects.
#Catalog: {
	#ResourceMeta
	kind: "Catalog"

	// Closed vocabulary of allowed project tags.
	tags?: [...#Tag]

	// Registered project list.
	projects: [...#Project]
}
