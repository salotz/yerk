package models

// Observed state for one replica checkout on this host.
#ReplicaStatus: {
	#ResourceMeta
	kind: "ReplicaStatus"

	uri?:     #ReplicaURI
	project:  #ProjectName
	replica:  #ReplicaName
	domain?:  #ProjectDomain
	path:     #AbsUnixPath
	presence: #Presence

	// Space-separated change flags, "-" when not probed, or "error".
	change?: #ChangeStatus

	// Checked-out branch short name when known; "-" if N/A.
	branch?: #BranchName

	tags?: [...#Tag]
	remote?: #GitRemote
}

// Compact default-replica snapshot embedded on ProjectStatus.
#ReplicaSummary: {
	name:     #ReplicaName
	path:     #AbsUnixPath
	presence: #Presence
	change?:  #ChangeStatus
	branch?:  #BranchName
}

// Rollup of replica presence/change for one project.
// Clean only when every present replica is clean and none are
// missing/invalid relative to the collected set.
#ProjectOverall: {
	presence:     #OverallPresence
	change?:      #ChangeStatus
	replicaCount: #Count
	presentCount: #Count
}

// Project-scoped status view (workspace + replicas + overall).
// ReplicaStatus remains the one-replica resource.
#ProjectStatus: {
	#ResourceMeta
	kind: "ProjectStatus"

	uri?:               #ProjectURI
	name:               #ProjectName
	domain?:            #ProjectDomain
	workspacePath:      #AbsUnixPath
	workspacePresence?: #Presence

	// Default/main replica (always included when known).
	defaultReplica?: #ReplicaSummary

	// Live on-disk replicas plus default when missing (probed).
	replicas?: [...#ReplicaStatus]

	overall?: #ProjectOverall
	tags?: [...#Tag]
	remote?: #GitRemote
}
