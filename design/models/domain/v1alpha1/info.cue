package models

// Single-project read model for get/lookup (ADR 015).
// Richer than bare Project: host paths, presence, and effective placement.
// workspacePath is the project workspace directory, not a replica checkout.
#ProjectInfo: {
	#ResourceMeta
	kind: "ProjectInfo"

	uri?:            #ProjectURI
	name:            #ProjectName
	domain?:         #ProjectDomain
	remote?:         #GitRemote
	defaultReplica?: #ReplicaName
	tags?: [...#Tag]

	// Absolute project workspace directory.
	workspacePath: #AbsUnixPath

	// Presence of the workspace directory.
	workspacePresence?: #Presence

	placement?: #PlacementInfo

	// Absolute path that triggered lookup (lookup only).
	matchedPath?: #AbsUnixPath
}

// Single-replica read model for get/lookup (ADR 015).
#ReplicaInfo: {
	#ResourceMeta
	kind: "ReplicaInfo"

	uri?: #ReplicaURI

	// Catalog project name.
	project: #ProjectName

	// Replica distinguisher.
	replica: #ReplicaName

	domain?: #ProjectDomain
	remote?: #GitRemote
	tags?: [...#Tag]

	// Absolute on-disk checkout path.
	path: #AbsUnixPath

	presence: #Presence

	// Owning project workspace directory.
	workspacePath?: #AbsUnixPath

	placement?: #PlacementInfo

	matchedPath?: #AbsUnixPath
}
