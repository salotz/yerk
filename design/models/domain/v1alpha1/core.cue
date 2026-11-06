package models

import (
	"salotz.info/git/refs"
	"salotz.info/git/url"
	"salotz.info/unix/envvar"
	"salotz.info/unix/path"
	"salotz.info/unix/varsets/posix"
	rfc027 "salotz.info/rfcs/027_env-nexps:schemas"
	xdg "specifications.freedesktop.org/basedir:schemas"
)

// Product API version stamped on yerk resources.
// v1alpha1: first draft of the v1 generation; no compatibility promise.
#APIVersion: "yerk/v1alpha1"

// Closed set of wire Kind values (ADR 011).
#Kind:
	"Catalog" |
	"Project" |
	"ProjectStatus" |
	"ReplicaStatus" |
	"ProjectInfo" |
	"ReplicaInfo" |
	"ConfigResolve" |
	"ToolConfig" |
	"ToolContext" |
	"DirContext" |
	"EnvVars"

// On-disk presence of an expected path (replica checkout or workspace dir).
#Presence: "missing" | "present" | "invalid"

// Rollup of replica presence for one project. Distinct from #Presence.
#OverallPresence:
	"all-present" |
	"partial" |
	"missing" |
	"invalid" |
	"none"

// DirContext: path matched a project workspace or a replica root.
#DirMatch: "project" | "replica"

// Shared resource metadata on every resource kind.
#ResourceMeta: {
	apiVersion: #APIVersion
	kind:       #Kind
}

// Workspace layout style (ADR 013/018).
#WorkspaceStyle: "workspace-dir" | "project-dir" | "name-tags"

// ADR 012 identifier segment (regex fragment, no anchors).
_identityPattern: "[A-Za-z0-9._-]+"

#IdentitySegment: string & =~"^\(_identityPattern)$"

#ProjectDomain: #IdentitySegment
#ProjectName:   #IdentitySegment
#ReplicaName:   #IdentitySegment

// Closed-vocabulary member (ADR 010). Same charset as identifier segments.
#Tag: #IdentitySegment

// Canonical URIs (ADR 012). Not url.#AbsoluteURL: scheme and segments are product grammar.
_projectURIPattern: "yerk://\(_identityPattern)/\(_identityPattern)"
_replicaURIPattern: "\(_projectURIPattern)/\(_identityPattern)"

#ProjectURI: string & =~"^\(_projectURIPattern)$"
#ReplicaURI: string & =~"^\(_replicaURIPattern)$"
#YerkURI:    #ProjectURI | #ReplicaURI

// Bare IDs (no scheme).
_projectIDPattern: "\(_identityPattern)/\(_identityPattern)"
_replicaIDPattern: "\(_projectIDPattern)/\(_identityPattern)"

#ProjectID: string & =~"^\(_projectIDPattern)$"
#ReplicaID: string & =~"^\(_replicaIDPattern)$"

// Absolute Unix path (lexical). Not a URL.
#AbsUnixPath: path.#AbsUnixPath

// Path as written in host config: absolute, ~/…, or relative.
#HostPath: path.#HostPath

// Git remote: https/ssh/file URL, scp (git@host:path), or local path.
// Not url.#AbsoluteURL: that rejects scp-style and bare paths.
#GitRemote: url.#Remote

// Space-separated change flags, "-" when not probed, or "error".
#ChangeStatus: string

// Git branch short name when known; "-" if N/A. May contain '/' (feat/foo).
#BranchName: refs.#BranchName | "-"

// Placement merge layer id (ADR 013).
#PlacementLayer:
	"built-in" |
	"host-config" |
	"dir-local" |
	"catalog" |
	"env" |
	"host-project" |
	"state" |
	"cli"

// Placement key considered in a config-resolve contribution (v1).
#PlacementKey: "workspaceStyle"

// Best-effort debug label for a contributing layer (may include path/value).
#PlacementSource: string

#PlacementValue: string

#Warning: string

// config resolve target as given (bare ID, URI, or path).
#ConfigTarget: string

// Process env name (mixed case). Live dumps may include extra names.
#EnvVarName: envvar.#Name

// Tool-owned knobs (RFC 027 name form, YERK__ field style).
#ToolEnvVar: rfc027.#EnvName & =~"^YERK__[A-Z][A-Z0-9_]*$"

// Platform names yerk reads. Not product knobs.
#PlatformEnvVar: (xdg.#Var | posix.#Name) & (
			"XDG_CONFIG_HOME" | "XDG_STATE_HOME" | "HOME" | "PATH")

// Binary stamp. Not RFC 002 #Version: untagged builds use "0.0.0-dev".
#ProductVersion: string

#CommandName: string

#Count: int & >=0

_examples: {
	domain:   #ProjectDomain & "personal"
	name:     #ProjectName & "yerk"
	replica:  #ReplicaName & "main"
	tag:      #Tag & "devel"
	uri:      #ProjectURI & "yerk://personal/yerk"
	ruri:     #ReplicaURI & "yerk://personal/yerk/main"
	pid:      #ProjectID & "personal/yerk"
	rid:      #ReplicaID & "personal/yerk/main"
	abs:      #AbsUnixPath & "/home/salotz/tree/personal/devel/yerk"
	host:     #HostPath & "~/tree/personal/devel/yerk"
	remote:   #GitRemote & "git@github.com:salotz/yerk.git"
	branch:   #BranchName & "feat/foo"
	detached: #BranchName & "-"
	env:      #EnvVarName & "YERK__CONFIG_DIR"
	mixed:    #EnvVarName & "http_proxy"
	tool:     #ToolEnvVar & "YERK__CONFIG_DIR"
	xdgCfg:   #PlatformEnvVar & "XDG_CONFIG_HOME"
	home:     #PlatformEnvVar & "HOME"
	style:    #WorkspaceStyle & "name-tags"
	presence: #Presence & "present"
}
