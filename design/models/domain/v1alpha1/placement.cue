package models

// Effective placement snapshot embedded on get/lookup (ADR 013/015).
#PlacementInfo: {
	// Effective workspace style name.
	style: #WorkspaceStyle

	// True when host project state supplied the winning style.
	bound: bool

	// Contributing layer labels (best-effort debug).
	sources?: [...#PlacementSource]

	// Ambient drift messages when bound state wins.
	warnings?: [...#Warning]
}

// Workspace style name plus optional path parameters (ADR 018).
// Product view; not the TOML decoder.
#StyleSpecification: {
	style?:      #WorkspaceStyle
	mainDir?:    #HostPath
	replicaDir?: #HostPath
}
