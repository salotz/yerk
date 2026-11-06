package models

// Host/tool config dump for `config show` (ADR 019).
// Distinct from ConfigResolve (per-project placement stack).
#ToolConfig: {
	#ResourceMeta
	kind: "ToolConfig"

	// Resolved config.toml path.
	file?: #AbsUnixPath

	workspace: #ToolConfigWorkspace

	// Optional domain-name → root map (stable name order).
	domains?: [...#ToolConfigDomain]

	// Optional host-local placement rows (stable ID order).
	projects?: [...#ToolConfigProject]

	stateDir?: #AbsUnixPath
}

#ToolConfigWorkspace: {
	style?: #WorkspaceStyle
}

// One [domains] entry.
#ToolConfigDomain: {
	name:          #ProjectDomain
	path:          #HostPath
	pathExpanded?: #AbsUnixPath
}

// One host [[projects]] row (paths, not catalog identity).
#ToolConfigProject: {
	name:            #ProjectName
	domain?:         #ProjectDomain
	uri?:            #ProjectURI
	path?:           #HostPath
	pathExpanded?:   #AbsUnixPath
	workspaceStyle?: #WorkspaceStyle
	style?:          #StyleSpecification
}

// One layer in a config resolve dump.
#ConfigResolveContribution: {
	order:   int
	layer:   #PlacementLayer
	path?:   #AbsUnixPath
	key:     #PlacementKey
	value?:  #PlacementValue
	applies: bool
	note?:   string
}

// Target-scoped placement contribution report (`config resolve`).
#ConfigResolve: {
	#ResourceMeta
	kind: "ConfigResolve"

	uri?:           #ProjectURI
	target?:        #ConfigTarget
	anchor?:        #AbsUnixPath
	workspacePath?: #AbsUnixPath
	effectiveStyle: #WorkspaceStyle
	bound:          bool
	warnings?: [...#Warning]
	files?: [...#AbsUnixPath]
	contributions: [...#ConfigResolveContribution]
}
