package models

import dirshape "salotz.info/rfcs/033_dirshape:schemas"

// On-disk layouts (RFC 033). Not the TOML/JSON file schemas.
// Product path math for project-dir and name-tags is parameterized
// (ADR 018); those styles are not a single layout spec.

#ConfigFilename:  "config.toml"
#CatalogFilename: "catalog.toml"
#StateFilename:   "state.json"

// Injected root is $XDG_CONFIG_HOME/yerk (or YERK__CONFIG_DIR).
#HostConfigLayoutSpec: dirshape.#LayoutSpec & {
	entries: {
		config: {name: #ConfigFilename, presence: "optional"}
		catalog: {name: #CatalogFilename, presence: "optional"}
	}
}

// Injected root is $XDG_STATE_HOME/yerk (or YERK__STATE_DIR).
#HostStateLayoutSpec: dirshape.#LayoutSpec & {
	entries: {
		projects: {
			name:     "projects/"
			presence: "optional"
			children: {
				domain: {
					name:     "*/"
					match:    "glob"
					presence: "optional"
					children: {
						project: {
							name:     "*/"
							match:    "glob"
							presence: "optional"
							children: {
								state: {name: #StateFilename}
							}
						}
					}
				}
			}
		}
	}
}

// Injected root is a directory on the walk from the workspace toward $HOME.
// ADR 013 dir-local: <dir>/.local/yerk/config.toml
#DirLocalLayoutSpec: dirshape.#LayoutSpec & {
	entries: {
		local: {
			name:     ".local/"
			presence: "optional"
			children: {
				yerk: {
					name:     "yerk/"
					presence: "optional"
					children: {
						config: {name: #ConfigFilename, presence: "optional"}
					}
				}
			}
		}
	}
}

// Injected root is the project workspace. workspace-dir style (ADR 018):
// replicas are child directories named by distinguisher.
#WorkspaceDirLayoutSpec: dirshape.#LayoutSpec & {
	entries: {
		replica: {
			name:     "*/"
			match:    "glob"
			presence: "optional"
		}
	}
}

_examples: {
	configFile: #ConfigFilename & "config.toml"
	stateFile:  #StateFilename & "state.json"
	config:     #HostConfigLayoutSpec
	state:      #HostStateLayoutSpec
	dirLocal:   #DirLocalLayoutSpec
	workspace:  #WorkspaceDirLayoutSpec
}
