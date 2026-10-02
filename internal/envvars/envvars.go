// Package envvars is the comprehensive registry of environment variables
// that yerk reads.
//
// This is the single source of truth for CLI environment help text and live
// dumps (ADR 005). Prefer updating this package over scattering env mentions
// in command Long strings.
//
// Static application-info declarations (RFC 030/031) live in
// .appinfo/meta.toml and MUST stay aligned with the tool-owned and platform
// entries here (names, type, policy, default story, enum values).
//
// Naming:
//
//   - YERK__*  — tool / host knobs (ADR 003, RFC 27 field style)
//   - XDG_* / HOME / PATH — platform names yerk actually reads (external)
//   - PRJX_* / PRJX__* — owned by PRJX (RFC 28); not product knobs. Help
//     points at PRJX docs rather than re-listing every name.
//
// Value metadata (Type, Policy, Default, Values) follows RFC 031/032 so
// `yerk help envvars` relays the same facts as .appinfo.
package envvars

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

// Scope classifies who defines the variable.
type Scope string

const (
	// ScopeTool variables are owned by the yerk CLI (YERK__).
	ScopeTool Scope = "tool"
	// ScopePlatform variables come from the OS / XDG and are inherited.
	ScopePlatform Scope = "platform"
)

// Value type names (RFC 032). Empty Type means undocumented in registry terms.
const (
	TypeString = "string"
	TypeEnum   = "enum"
	TypeBool   = "boolean"
)

// Value policies (RFC 032). Empty Policy is treated as PolicyWarn in formatters.
const (
	PolicySilent   = "silent"
	PolicyWarn     = "warn"
	PolicyStrict   = "strict"
	PolicyRequired = "required"
)

// SeeAlsoFullReference points at documentation (not the live-values command).
const SeeAlsoFullReference = "Full environment reference: yerk help envvars"

// SeeAlsoLiveValues points at the runtime name/value dump.
const SeeAlsoLiveValues = "Current process values: yerk envvars"

// SeeAlsoAppinfo points at the static RFC 030/031 declarations.
const SeeAlsoAppinfo = "Static declarations: .appinfo/meta.toml (RFC 030/031)"

// Var is one environment variable entry.
type Var struct {
	// Name is the full process environment key.
	Name string
	// Scope ownership class.
	Scope Scope
	// Summary is one-line help (RFC 031 description).
	Summary string
	// Long is optional extended prose (RFC 031 long_description).
	Long string
	// Type is an RFC 032 type name (string, enum, boolean, …).
	Type string
	// Policy is an RFC 032 value policy (silent|warn|strict|required).
	// Empty means warn.
	Policy string
	// Default describes the default when unset (prose and/or literal).
	// Empty when Policy is required.
	Default string
	// Values lists canonical enum tokens when Type is enum.
	Values []string
	// External is true when yerk uses the name but does not define it.
	External bool
	// Global is true if the var can affect any command that loads host state.
	Global bool
	// Commands lists cobra command paths that read this var (e.g. "status",
	// "config path"). Empty means document at root/full reference only when
	// not matched via Global tooling conventions in formatters.
	Commands []string
	// HelpPrimary: include on root summary and on matching command --help.
	// Full reference always includes every registry entry.
	HelpPrimary bool
}

// EffectivePolicy returns the RFC 032 policy, defaulting to warn.
func (v Var) EffectivePolicy() string {
	if v.Policy == "" {
		return PolicyWarn
	}
	return v.Policy
}

// catalogLoadCommands are commands that load tool config and/or catalog.
func catalogLoadCommands() []string {
	return []string{
		"status", "path", "workspace", "workspace ensure", "materialize",
	}
}

func joinCmds(base []string, extra ...string) []string {
	out := make([]string, 0, len(base)+len(extra))
	out = append(out, base...)
	out = append(out, extra...)
	return out
}

// All is the full registry in stable display order.
// Keep tool + platform rows aligned with .appinfo/meta.toml [env.vars.*].
func All() []Var {
	return []Var{
		// --- tool (YERK__) ---
		{
			Name:    "YERK__CONFIG_DIR",
			Scope:   ScopeTool,
			Summary: "Directory containing config.toml and catalog.toml by default",
			Long: "When set, both default file paths are under this directory unless " +
				"YERK__CONFIG or YERK__CATALOG overrides a file explicitly.",
			Type:        TypeString,
			Policy:      PolicyWarn,
			Default:     "${XDG_CONFIG_HOME}/yerk (else ~/.config/yerk via platform user config dir)",
			Global:      true,
			HelpPrimary: true,
			Commands: joinCmds(catalogLoadCommands(),
				"config", "config path", "config show",
				"catalog", "catalog path", "catalog show"),
		},
		{
			Name:    "YERK__CONFIG",
			Scope:   ScopeTool,
			Summary: "Absolute path to the tool config file (config.toml)",
			Long: "Tool/host behavior only: [workspace].style, optional [domains] roots, and optional " +
				"[[projects]] host rows (ADR 014). Not the project catalog.",
			Type:        TypeString,
			Policy:      PolicyWarn,
			Default:     "${YERK__CONFIG_DIR}/config.toml (else ${XDG_CONFIG_HOME}/yerk/config.toml)",
			Global:      true,
			HelpPrimary: true,
			Commands: joinCmds(catalogLoadCommands(),
				"config", "config path", "config show"),
		},
		{
			Name:    "YERK__CATALOG",
			Scope:   ScopeTool,
			Summary: "Absolute path to the project catalog file (catalog.toml)",
			Long: "Project registry only (catalog [[projects]]: identity, remote, tags). " +
				"No host workspace path field (ADR 014). Missing file means empty catalog, not an error.",
			Type:        TypeString,
			Policy:      PolicyWarn,
			Default:     "${YERK__CONFIG_DIR}/catalog.toml (else ${XDG_CONFIG_HOME}/yerk/catalog.toml)",
			Global:      true,
			HelpPrimary: true,
			Commands: joinCmds(catalogLoadCommands(),
				"catalog", "catalog path", "catalog show"),
		},
		{
			Name:    "YERK__STATE_DIR",
			Scope:   ScopeTool,
			Summary: "Directory for tool-written project bindings (state.json trees)",
			Long: "Host project state root (ADR 013). Default layout: " +
				"projects/<domain>/<project>/state.json under this directory. " +
				"Does not replace config.toml or catalog.toml.",
			Type:        TypeString,
			Policy:      PolicyWarn,
			Default:     "${XDG_STATE_HOME}/yerk (else ~/.local/state/yerk)",
			Global:      true,
			HelpPrimary: true,
			Commands: joinCmds(catalogLoadCommands(),
				"config", "config show", "workspace ensure", "materialize"),
		},
		{
			Name:    "YERK__WORKSPACE_STYLE",
			Scope:   ScopeTool,
			Summary: "Override [workspace].style for replica path layout",
			Long: "Ambient overlay for workspace style (ADR 013). " +
				"workspace-dir: <workspace>/<replica> (e.g. …/yerk/main). " +
				"project-dir: <dir(workspace)>/<name>__<replica> (e.g. …/yerk__main). " +
				"Conflicts with bound host project state warn; state wins. " +
				"Workspace path: optional [domains] default <root>/<name> or [[projects]] path (ADR 014).",
			Type:        TypeEnum,
			Policy:      PolicyWarn,
			Default:     "unset → config [workspace].style; if empty, built-in default workspace-dir",
			Values:      []string{"workspace-dir", "project-dir"},
			Global:      true,
			HelpPrimary: true,
			Commands:    joinCmds(catalogLoadCommands(), "config show"),
		},

		// --- platform (external; used when resolving defaults or invoking git) ---
		{
			Name:        "XDG_CONFIG_HOME",
			Scope:       ScopePlatform,
			Summary:     "XDG base directory for user config; yerk appends /yerk when YERK__CONFIG_DIR is unset",
			Type:        TypeString,
			Policy:      PolicyWarn,
			Default:     "platform / XDG Base Directory default when unset (commonly ~/.config)",
			External:    true,
			Global:      true,
			HelpPrimary: true,
			Commands:    joinCmds(catalogLoadCommands(), "config", "catalog"),
		},
		{
			Name:        "HOME",
			Scope:       ScopePlatform,
			Summary:     "User home; used by Go user config dir resolution when XDG_CONFIG_HOME is unset",
			Type:        TypeString,
			Policy:      PolicyWarn,
			Default:     "platform home directory for the process user",
			External:    true,
			Global:      true,
			HelpPrimary: false,
			Commands:    joinCmds(catalogLoadCommands(), "config", "catalog"),
		},
		{
			Name:        "PATH",
			Scope:       ScopePlatform,
			Summary:     "Process PATH; must include git for materialize, status change probes, and default-branch ls-remote",
			Type:        TypeString,
			Policy:      PolicyWarn,
			Default:     "inherited process PATH",
			External:    true,
			Global:      false,
			HelpPrimary: true,
			Commands:    []string{"status", "materialize"},
		},
	}
}

// ToolVars returns only YERK__ entries (what the binary owns today).
func ToolVars() []Var {
	var out []Var
	for _, v := range All() {
		if v.Scope == ScopeTool {
			out = append(out, v)
		}
	}
	return out
}

// ForCommand returns vars that apply to a command path (e.g. "status").
// Empty commandPath is not used; callers wanting everything use All().
func ForCommand(commandPath string) []Var {
	commandPath = strings.TrimSpace(commandPath)
	var out []Var
	for _, v := range All() {
		if matchesCommand(v, commandPath) {
			out = append(out, v)
		}
	}
	return out
}

// ForCommandPrimary is ForCommand filtered to HelpPrimary entries.
func ForCommandPrimary(commandPath string) []Var {
	var out []Var
	for _, v := range ForCommand(commandPath) {
		if v.HelpPrimary {
			out = append(out, v)
		}
	}
	return out
}

func matchesCommand(v Var, commandPath string) bool {
	if len(v.Commands) == 0 {
		return false
	}
	for _, c := range v.Commands {
		if c == commandPath {
			return true
		}
		// parent path match: "config" matches "config show"
		if strings.HasPrefix(commandPath, c+" ") || strings.HasPrefix(c, commandPath+" ") {
			return true
		}
		// exact parent: commandPath "config" and entry "config path"
		if commandPath == "config" && strings.HasPrefix(c, "config") {
			return true
		}
		if commandPath == "catalog" && strings.HasPrefix(c, "catalog") {
			return true
		}
	}
	return false
}

func writeVarLines(b *strings.Builder, v Var, detail bool, withCommands bool) {
	fmt.Fprintf(b, "    %-24s %s\n", v.Name, v.Summary)
	if detail {
		meta := formatMeta(v)
		if meta != "" {
			fmt.Fprintf(b, "    %-24s %s\n", "", meta)
		}
		if v.Default != "" && v.EffectivePolicy() != PolicyRequired {
			fmt.Fprintf(b, "    %-24s default: %s\n", "", v.Default)
		}
		if len(v.Values) > 0 {
			fmt.Fprintf(b, "    %-24s values: %s\n", "", strings.Join(v.Values, ", "))
		}
		if detail && v.Long != "" {
			fmt.Fprintf(b, "    %-24s %s\n", "", v.Long)
		}
	} else {
		// Compact primary surfaces still show type/policy + default.
		meta := formatMeta(v)
		if meta != "" {
			fmt.Fprintf(b, "    %-24s %s\n", "", meta)
		}
		if v.Default != "" && v.EffectivePolicy() != PolicyRequired {
			fmt.Fprintf(b, "    %-24s default: %s\n", "", v.Default)
		}
		if len(v.Values) > 0 {
			fmt.Fprintf(b, "    %-24s values: %s\n", "", strings.Join(v.Values, ", "))
		}
	}
	if withCommands {
		cmds := formatCommandList(v)
		if cmds != "" {
			fmt.Fprintf(b, "    %-24s commands: %s\n", "", cmds)
		}
	}
}

func formatMeta(v Var) string {
	var parts []string
	if v.Type != "" {
		parts = append(parts, "type="+v.Type)
	}
	parts = append(parts, "policy="+v.EffectivePolicy())
	if v.External {
		parts = append(parts, "external")
	}
	if v.EffectivePolicy() == PolicyRequired {
		parts = append(parts, "required")
	}
	return strings.Join(parts, ", ")
}

func formatCommandList(v Var) string {
	if v.Global && len(v.Commands) == 0 {
		return "(global; all host-state commands)"
	}
	if len(v.Commands) == 0 {
		return ""
	}
	// Stable unique display: prefer unique sorted short paths.
	seen := make(map[string]struct{}, len(v.Commands))
	var list []string
	for _, c := range v.Commands {
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		list = append(list, c)
	}
	sort.Strings(list)
	return strings.Join(list, ", ")
}

// FormatRootHelp returns the Environment section for `yerk --help` (primary summary).
func FormatRootHelp() string {
	var b strings.Builder
	b.WriteString("Environment variables (summary):\n")
	b.WriteString("  Tool-owned names use YERK__ (double underscore).\n")
	b.WriteString("  Each entry shows type, policy (RFC 032), and default.\n")
	b.WriteString("\n")
	b.WriteString("  Primary tool knobs:\n")
	b.WriteString("\n")
	for _, v := range All() {
		if !v.HelpPrimary || v.Scope != ScopeTool {
			continue
		}
		writeVarLines(&b, v, false, false)
	}
	b.WriteString("\n")
	b.WriteString("  Also commonly relevant:\n")
	for _, v := range All() {
		if !v.HelpPrimary || v.Scope != ScopePlatform {
			continue
		}
		writeVarLines(&b, v, false, false)
	}
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(SeeAlsoFullReference)
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(SeeAlsoLiveValues)
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(SeeAlsoAppinfo)
	b.WriteString("\n")
	return b.String()
}

// FormatFullReference returns the complete environment documentation for
// `yerk help envvars` (and `yerk envvars --help`).
func FormatFullReference() string {
	var b strings.Builder
	b.WriteString("Environment variable reference\n")
	b.WriteString("\n")
	b.WriteString("yerk reads tool knobs under YERK__ (double underscore; ADR 003).\n")
	b.WriteString("Metadata matches .appinfo/meta.toml and RFC 031/032:\n")
	b.WriteString("  type, policy (silent|warn|strict|required), default, enum values.\n")
	b.WriteString("policy=required means the variable must be set (no default).\n")
	b.WriteString("This list is generated from the internal registry (single source of truth for CLI help).\n")
	b.WriteString("Command --help shows a primary subset; this page is complete for yerk-owned/read names.\n")
	b.WriteString("For current process values (app-scoped env), run: yerk envvars\n")
	b.WriteString("\n")

	b.WriteString("Global (affect host state / config load for many commands)\n")
	b.WriteString("\n")
	for _, v := range All() {
		if !v.Global {
			continue
		}
		writeVarLines(&b, v, true, false)
		b.WriteString("\n")
	}

	b.WriteString("Command-scoped (not global; listed commands only)\n")
	b.WriteString("\n")
	anyScoped := false
	for _, v := range All() {
		if v.Global {
			continue
		}
		anyScoped = true
		writeVarLines(&b, v, true, true)
		b.WriteString("\n")
	}
	if !anyScoped {
		b.WriteString("    (none)\n\n")
	}

	b.WriteString("PRJX (RFC 28) — not yerk product env\n")
	b.WriteString("\n")
	b.WriteString("    PRJX_* / PRJX__* names are defined by PRJX, not by yerk.\n")
	b.WriteString("    yerk does not require them for near-term catalog/materialize/status.\n")
	b.WriteString("    See .prjx-root, .config/_project-meta.toml, and salotz RFC 28.\n")
	b.WriteString("    They are intentionally omitted from .appinfo/meta.toml [env.vars].\n")
	b.WriteString("\n")

	b.WriteString("See also: .appinfo/meta.toml, design/decisions/003-config-xdg-and-env.md,\n")
	b.WriteString("design/decisions/005-cli-help-and-envvars.md, examples/ in the repository.\n")
	b.WriteString(SeeAlsoLiveValues)
	b.WriteString("\n")
	return b.String()
}

// FormatCommandHelp returns an Environment section for a subcommand --help
// (primary vars only + pointers).
func FormatCommandHelp(commandPath string) string {
	vars := ForCommandPrimary(commandPath)
	var b strings.Builder
	b.WriteString("Environment variables (most relevant):\n")
	if len(vars) == 0 {
		b.WriteString("  (none beyond process defaults for this command)\n")
		b.WriteString("  ")
		b.WriteString(SeeAlsoFullReference)
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(SeeAlsoLiveValues)
		b.WriteString("\n")
		return b.String()
	}
	for _, v := range vars {
		writeVarLines(&b, v, false, false)
	}
	b.WriteString("  ")
	b.WriteString(SeeAlsoFullReference)
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(SeeAlsoLiveValues)
	b.WriteString("\n")
	return b.String()
}

// isPatternName reports registry rows that are not real process keys.
func isPatternName(name string) bool {
	return strings.Contains(name, "*")
}

// LookupEnvFunc is like os.LookupEnv (value, wasSet).
type LookupEnvFunc func(name string) (string, bool)

// FormatLiveValues returns a NAME / VALUE table for the current process,
// like a shell `env` limited to yerk-relevant keys.
//
// lookup defaults to os.LookupEnv when nil. environ is scanned for extra
// YERK__ keys not in the registry (pass os.Environ() in production).
func FormatLiveValues(lookup LookupEnvFunc, environ []string) string {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	type row struct {
		name  string
		value string
		extra bool // not in registry (prefix scan)
	}
	var rows []row
	seen := make(map[string]struct{})

	for _, v := range All() {
		if isPatternName(v.Name) {
			continue
		}
		seen[v.Name] = struct{}{}
		val, ok := lookup(v.Name)
		if !ok {
			rows = append(rows, row{name: v.Name, value: "(unset)"})
			continue
		}
		rows = append(rows, row{name: v.Name, value: val})
	}

	// Forward-compatible / stray knobs: take name+value from environ entries
	// so test injects and process scans both work without a second lookup.
	for _, name := range extraPrefixNames(environ, seen, "YERK__") {
		rows = append(rows, row{
			name:  name,
			value: valueFromEnviron(environ, name),
			extra: true,
		})
	}

	var b strings.Builder
	b.WriteString("yerk-relevant environment (this process)\n")
	b.WriteString("Raw values from the process environment.\n")
	b.WriteString("Missing keys shown as (unset); a key set to empty string shows blank.\n")
	b.WriteString("Does not resolve effective config paths (see yerk config path).\n")
	b.WriteString("Documentation: yerk help envvars\n")
	b.WriteString("\n")

	tw := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "NAME\tVALUE\n")
	for _, r := range rows {
		name := r.name
		if r.extra {
			name = r.name + " *"
		}
		fmt.Fprintf(tw, "%s\t%s\n", name, displayValue(r.value))
	}
	_ = tw.Flush()
	b.WriteString("\n")
	b.WriteString("* = present in process env but not in the built-in registry\n")
	return b.String()
}

func extraPrefixNames(environ []string, seen map[string]struct{}, prefixes ...string) []string {
	if len(environ) == 0 {
		return nil
	}
	var out []string
	for _, entry := range environ {
		name, _, ok := strings.Cut(entry, "=")
		if !ok || name == "" {
			continue
		}
		if _, already := seen[name]; already {
			continue
		}
		for _, p := range prefixes {
			if strings.HasPrefix(name, p) {
				out = append(out, name)
				break
			}
		}
	}
	sort.Strings(out)
	return out
}

func valueFromEnviron(environ []string, want string) string {
	prefix := want + "="
	for _, entry := range environ {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}
	return ""
}

func displayValue(v string) string {
	// Keep single-line table cells; collapse newlines in pathological values.
	if strings.ContainsAny(v, "\n\r") {
		v = strings.ReplaceAll(v, "\r\n", "\\n")
		v = strings.ReplaceAll(v, "\n", "\\n")
		v = strings.ReplaceAll(v, "\r", "\\n")
	}
	return v
}

// NamesTool returns stable list of YERK__ names (for tests).
func NamesTool() []string {
	var names []string
	for _, v := range ToolVars() {
		names = append(names, v.Name)
	}
	return names
}
