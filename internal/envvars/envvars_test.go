package envvars_test

import (
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/envvars"
)

func TestToolNamesComplete(t *testing.T) {
	t.Parallel()
	want := []string{
		"YERK__CONFIG_DIR",
		"YERK__CONFIG",
		"YERK__CATALOG",
		"YERK__STATE_DIR",
		"YERK__WORKSPACE_STYLE",
	}
	got := envvars.NamesTool()
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestRegistryHasTypePolicyDefault(t *testing.T) {
	t.Parallel()
	for _, v := range envvars.All() {
		if v.Type == "" {
			t.Fatalf("%s missing type", v.Name)
		}
		if v.EffectivePolicy() == "" {
			t.Fatalf("%s missing policy", v.Name)
		}
		if v.EffectivePolicy() == envvars.PolicyRequired && v.Default != "" {
			t.Fatalf("%s is required but has default %q", v.Name, v.Default)
		}
		if v.EffectivePolicy() != envvars.PolicyRequired && v.Default == "" {
			t.Fatalf("%s non-required should declare default prose", v.Name)
		}
		if v.Type == envvars.TypeEnum && len(v.Values) == 0 {
			t.Fatalf("%s enum missing values", v.Name)
		}
		if v.Scope == envvars.ScopePlatform && !v.External {
			t.Fatalf("%s platform should be external", v.Name)
		}
		if v.Scope == envvars.ScopeTool && v.External {
			t.Fatalf("%s tool should not be external", v.Name)
		}
	}
}

func TestRootHelpIsPrimarySummary(t *testing.T) {
	t.Parallel()
	h := envvars.FormatRootHelp()
	for _, name := range envvars.NamesTool() {
		if !strings.Contains(h, name) {
			t.Fatalf("root help missing primary tool var %s", name)
		}
	}
	if !strings.Contains(h, "type=string") {
		t.Fatal("root help should show type metadata")
	}
	if !strings.Contains(h, "policy=warn") {
		t.Fatal("root help should show policy metadata")
	}
	if !strings.Contains(h, "default:") {
		t.Fatal("root help should show defaults")
	}
	if !strings.Contains(h, "values: workspace-dir, project-dir") {
		t.Fatal("root help should list WORKSPACE_STYLE enum values")
	}
	if !strings.Contains(h, "yerk help envvars") {
		t.Fatal("root help should point at full envvars reference")
	}
	if !strings.Contains(h, ".appinfo/meta.toml") {
		t.Fatal("root help should point at static declarations")
	}
	// Full PRJX dump stays off root summary
	if strings.Contains(h, "PRJX_CACHE_HOME") || strings.Contains(h, "PRJX_ROOT") {
		t.Fatal("root summary should not dump PRJX names")
	}
	if strings.Contains(h, "HOME") && strings.Contains(h, "platform home") {
		t.Fatal("root summary should not list non-primary HOME")
	}
}

func TestFullReferenceListsGlobalAndCommandScoped(t *testing.T) {
	t.Parallel()
	h := envvars.FormatFullReference()
	for _, name := range envvars.NamesTool() {
		if !strings.Contains(h, name) {
			t.Fatalf("full reference missing %s", name)
		}
	}
	if !strings.Contains(h, "XDG_CONFIG_HOME") {
		t.Fatal("expected XDG_CONFIG_HOME")
	}
	if !strings.Contains(h, "type=string") || !strings.Contains(h, "policy=warn") {
		t.Fatal("full reference should include type/policy lines")
	}
	if !strings.Contains(h, "type=enum") {
		t.Fatal("full reference should mark WORKSPACE_STYLE as enum")
	}
	if !strings.Contains(h, "values: workspace-dir, project-dir") {
		t.Fatal("full reference should list enum values")
	}
	if !strings.Contains(h, "external") {
		t.Fatal("full reference should mark platform vars external")
	}
	if !strings.Contains(h, "Global") {
		t.Fatal("expected Global section")
	}
	if !strings.Contains(h, "Command-scoped") {
		t.Fatal("expected Command-scoped section")
	}
	// PATH is command-scoped and should list commands
	if !strings.Contains(h, "PATH") {
		t.Fatal("expected PATH")
	}
	if !strings.Contains(h, "commands:") {
		t.Fatal("command-scoped entries should list commands")
	}
	idxPath := strings.Index(h, "PATH")
	idxScoped := strings.Index(h, "Command-scoped")
	if idxPath < 0 || idxScoped < 0 || idxPath < idxScoped {
		t.Fatal("PATH should appear under Command-scoped")
	}
	// PRJX is a pointer section, not a name dump
	if !strings.Contains(h, "PRJX (RFC 28)") {
		t.Fatal("expected PRJX pointer section")
	}
	if strings.Contains(h, "PRJX_CACHE_HOME") || strings.Contains(h, "PRJX_ROOT") {
		t.Fatal("full reference should not re-list every PRJX_* name")
	}
	if !strings.Contains(h, ".appinfo/meta.toml") {
		t.Fatal("full reference should cite .appinfo")
	}
}

func TestStatusHelpIncludesPrimaryLoadVars(t *testing.T) {
	t.Parallel()
	h := envvars.FormatCommandHelp("status")
	for _, name := range []string{
		"YERK__CONFIG_DIR",
		"YERK__CONFIG",
		"YERK__CATALOG",
		"YERK__WORKSPACE_STYLE",
		"PATH",
	} {
		if !strings.Contains(h, name) {
			t.Fatalf("status help missing %s\n%s", name, h)
		}
	}
	if !strings.Contains(h, "policy=warn") {
		t.Fatalf("status help should show policy\n%s", h)
	}
	if !strings.Contains(h, "default:") {
		t.Fatalf("status help should show default\n%s", h)
	}
	if !strings.Contains(h, "yerk help envvars") {
		t.Fatal("command help should point at full reference")
	}
	if strings.Contains(h, "PRJX_CACHE_HOME") {
		t.Fatal("subcommand should not dump full PRJX list")
	}
	if strings.Contains(h, "platform home") {
		t.Fatal("non-primary HOME should not appear on status --help")
	}
}

func TestVersionHasNoToolEnv(t *testing.T) {
	t.Parallel()
	h := envvars.FormatCommandHelp("version")
	if strings.Contains(h, "YERK__CONFIG") {
		t.Fatalf("version should not claim config env: %s", h)
	}
}

func TestLiveValuesTable(t *testing.T) {
	t.Parallel()
	lookup := func(name string) (string, bool) {
		switch name {
		case "YERK__CONFIG_DIR":
			return "/tmp/yerk-test", true
		case "PATH":
			return "/usr/bin", true
		default:
			return "", false
		}
	}
	environ := []string{
		"YERK__CONFIG_DIR=/tmp/yerk-test",
		"YERK__FUTURE_KNOB=1",
		"PATH=/usr/bin",
		"UNRELATED=x",
	}
	h := envvars.FormatLiveValues(lookup, environ)
	if !strings.Contains(h, "YERK__CONFIG_DIR") || !strings.Contains(h, "/tmp/yerk-test") {
		t.Fatalf("missing set value\n%s", h)
	}
	if !strings.Contains(h, "(unset)") {
		t.Fatalf("expected unset markers\n%s", h)
	}
	if !strings.Contains(h, "YERK__FUTURE_KNOB") {
		t.Fatalf("expected extra prefix key\n%s", h)
	}
	if strings.Contains(h, "UNRELATED") {
		t.Fatalf("should not list unrelated env\n%s", h)
	}
	if strings.Contains(h, "PRJX__*") {
		t.Fatalf("pattern names should not appear as live keys\n%s", h)
	}
}
