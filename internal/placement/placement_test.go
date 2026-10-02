package placement_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/placement"
	"github.com/salotz/yerk/internal/state"
	"github.com/salotz/yerk/internal/workspace"
)

func TestResolvePrecedenceCatalogOverHost(t *testing.T) {
	t.Setenv("YERK__STATE_DIR", t.TempDir())
	os.Unsetenv("YERK__WORKSPACE_STYLE")

	eff, err := placement.Resolve(placement.Input{
		Host:      config.Config{Workspace: config.Workspace{Style: workspace.StyleWorkspaceDir}},
		Project:   config.Project{Domain: "personal", Name: "x", WorkspaceStyle: workspace.StyleProjectDir},
		LookupEnv: func(string) string { return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	if eff.Style != workspace.StyleProjectDir {
		t.Fatalf("style %q", eff.Style)
	}
	if eff.Bound {
		t.Fatal("should not be bound")
	}
}

func TestResolveStateWinsWithWarning(t *testing.T) {
	root := t.TempDir()
	t.Setenv("YERK__STATE_DIR", root)
	if _, err := state.BindStyle("personal", "x", workspace.StyleWorkspaceDir); err != nil {
		t.Fatal(err)
	}
	eff, err := placement.Resolve(placement.Input{
		Host:      config.Config{Workspace: config.Workspace{Style: workspace.StyleProjectDir}},
		Project:   config.Project{Domain: "personal", Name: "x", WorkspaceStyle: workspace.StyleProjectDir},
		LookupEnv: func(string) string { return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	if !eff.Bound || eff.Style != workspace.StyleWorkspaceDir {
		t.Fatalf("eff=%+v", eff)
	}
	if len(eff.Warnings) == 0 {
		t.Fatal("expected ambient drift warning")
	}
	if !strings.Contains(eff.Warnings[0], "bound workspace style") {
		t.Fatalf("warning: %q", eff.Warnings[0])
	}
}

func TestResolveEnvAmbientVsState(t *testing.T) {
	root := t.TempDir()
	t.Setenv("YERK__STATE_DIR", root)
	if _, err := state.BindStyle("personal", "x", workspace.StyleWorkspaceDir); err != nil {
		t.Fatal(err)
	}
	eff, err := placement.Resolve(placement.Input{
		Host:    config.Config{},
		Project: config.Project{Domain: "personal", Name: "x"},
		LookupEnv: func(k string) string {
			if k == "YERK__WORKSPACE_STYLE" {
				return workspace.StyleProjectDir
			}
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if eff.Style != workspace.StyleWorkspaceDir {
		t.Fatalf("state should win, got %q", eff.Style)
	}
	found := false
	for _, w := range eff.Warnings {
		if strings.Contains(w, "YERK__WORKSPACE_STYLE") || strings.Contains(w, "bound workspace style") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected warning, got %v", eff.Warnings)
	}
}

func TestResolveCLIContradictsState(t *testing.T) {
	root := t.TempDir()
	t.Setenv("YERK__STATE_DIR", root)
	if _, err := state.BindStyle("personal", "x", workspace.StyleWorkspaceDir); err != nil {
		t.Fatal(err)
	}
	_, err := placement.Resolve(placement.Input{
		Host:      config.Config{},
		Project:   config.Project{Domain: "personal", Name: "x"},
		CLIStyle:  workspace.StyleProjectDir,
		LookupEnv: func(string) string { return "" },
	})
	if err == nil {
		t.Fatal("expected CLI vs state error")
	}
	if !strings.Contains(err.Error(), "contradicts bound style") {
		t.Fatalf("err: %v", err)
	}
}

func TestResolveDirLocalNearWins(t *testing.T) {
	t.Setenv("YERK__STATE_DIR", t.TempDir())
	base := t.TempDir()
	// far
	far := filepath.Join(base, ".local", "yerk")
	if err := os.MkdirAll(far, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(far, "config.toml"), []byte("[workspace]\nstyle = \"workspace-dir\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// near under base/leaf
	leaf := filepath.Join(base, "leaf")
	near := filepath.Join(leaf, ".local", "yerk")
	if err := os.MkdirAll(near, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(near, "config.toml"), []byte("[workspace]\nstyle = \"project-dir\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	eff, err := placement.Resolve(placement.Input{
		Host:      config.Config{Workspace: config.Workspace{Style: workspace.StyleWorkspaceDir}},
		Project:   config.Project{Domain: "personal", Name: "x"},
		Anchor:    leaf,
		LookupEnv: func(string) string { return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	if eff.Style != workspace.StyleProjectDir {
		t.Fatalf("near should win: got %q sources=%v", eff.Style, eff.Sources)
	}
}

func TestInitStyleSkipsState(t *testing.T) {
	root := t.TempDir()
	t.Setenv("YERK__STATE_DIR", root)
	if _, err := state.BindStyle("personal", "x", workspace.StyleWorkspaceDir); err != nil {
		t.Fatal(err)
	}
	style, err := placement.InitStyle(placement.Input{
		Host:      config.Config{Workspace: config.Workspace{Style: workspace.StyleProjectDir}},
		Project:   config.Project{Domain: "personal", Name: "x"},
		LookupEnv: func(string) string { return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	if style != workspace.StyleProjectDir {
		t.Fatalf("init should snapshot ambient, got %q", style)
	}
}

func TestUnknownStyleError(t *testing.T) {
	t.Setenv("YERK__STATE_DIR", t.TempDir())
	_, err := placement.Resolve(placement.Input{
		Host:      config.Config{Workspace: config.Workspace{Style: "nope"}},
		Project:   config.Project{Domain: "d", Name: "n"},
		LookupEnv: func(string) string { return "" },
	})
	if err == nil {
		t.Fatal("expected unknown style")
	}
}
