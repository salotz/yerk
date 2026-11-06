package models

// Live process environment dump for `yerk envvars` (ADR 005/019).
#EnvVars: {
	#ResourceMeta
	kind: "EnvVars"
	vars: [...#EnvVarValue]
}

#EnvVarValue: {
	name:   #EnvVarName
	value?: string
	set:    bool
	extra?: bool
}

// Tool-wide agent dump (ADR 017): vocabulary, commands, host paths.
#ToolContext: {
	#ResourceMeta
	kind: "ToolContext"

	version?: #ProductVersion
	vocabulary?: [...#ContextTerm]
	commands?: [...#ContextCommand]
	paths: #ToolContextPaths
	envPrimary?: [...#EnvVarName]
	howTo?: [...string]
	notes?: [...string]
}

#ContextTerm: {
	term: string
	def:  string
}

#ContextCommand: {
	name: #CommandName
	role: string
}

#ToolContextPaths: {
	configDir?:   #AbsUnixPath
	configFile?:  #AbsUnixPath
	catalogFile?: #AbsUnixPath
	stateDir?:    #AbsUnixPath
}

// Directory-scoped agent dump (ADR 017): lookup + placement + short status.
#DirContext: {
	#ResourceMeta
	kind: "DirContext"

	// Absolute path that was resolved (current working directory or argument).
	path: #AbsUnixPath

	matched?: #DirMatch

	project?: #ProjectInfo
	replica?: #ReplicaInfo

	placement?:      #PlacementInfo
	effectiveStyle?: #WorkspaceStyle
	bound?:          bool
	warnings?: [...#Warning]
	statusOverall?: #ProjectOverall
	liveReplicas?: [...#ReplicaName]
	notes?: [...string]
}
