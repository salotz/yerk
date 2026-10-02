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

func TestExplainContributionOrder(t *testing.T) {
	root := t.TempDir()
	t.Setenv("YERK__STATE_DIR", filepath.Join(root, "state"))
	t.Setenv("YERK__CONFIG_DIR", filepath.Join(root, "cfg"))
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")

	cfgDir := filepath.Join(root, "cfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgBody := `[workspace]
style = "workspace-dir"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(cfgBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "catalog.toml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	anchor := filepath.Join(root, "tree", "devel", "proj")
	localDir := filepath.Join(root, "tree", "devel", ".local", "yerk")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(anchor, 0o755); err != nil {
		t.Fatal(err)
	}
	localBody := `[workspace]
style = "project-dir"
`
	if err := os.WriteFile(filepath.Join(localDir, "config.toml"), []byte(localBody), 0o644); err != nil {
		t.Fatal(err)
	}

	host := config.Config{
		Workspace: config.Workspace{Style: workspace.StyleWorkspaceDir},
	}
	// Load host style from file path via Explain using in.Host (already set).
	rep, err := placement.Explain(placement.Input{
		Host:    host,
		Project: config.Project{Domain: "personal", Name: "proj", WorkspaceStyle: ""},
		Anchor:  anchor,
		LookupEnv: func(k string) string {
			if k == "YERK__WORKSPACE_STYLE" {
				return ""
			}
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Effective.Style != workspace.StyleProjectDir {
		t.Fatalf("effective %q want project-dir (dir-local)", rep.Effective.Style)
	}
	var sawLocal, sawHost bool
	for _, c := range rep.Contributions {
		if c.Layer == placement.LayerDirLocal && c.Applies {
			sawLocal = true
			if !strings.Contains(c.Path, ".local/yerk/config.toml") {
				t.Fatalf("dir-local path %q", c.Path)
			}
		}
		if c.Layer == placement.LayerHostConfig {
			sawHost = true
		}
	}
	if !sawLocal || !sawHost {
		t.Fatalf("layers: local=%v host=%v contrib=%+v", sawLocal, sawHost, rep.Contributions)
	}
	if len(rep.Files) == 0 {
		t.Fatal("expected files list")
	}
}

func TestExplainOmitsMissingStatePath(t *testing.T) {
	root := t.TempDir()
	t.Setenv("YERK__STATE_DIR", root)
	t.Setenv("YERK__CONFIG_DIR", filepath.Join(root, "cfg"))
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")
	if err := os.MkdirAll(filepath.Join(root, "cfg"), 0o755); err != nil {
		t.Fatal(err)
	}

	rep, err := placement.Explain(placement.Input{
		Host:      config.Config{Workspace: config.Workspace{Style: workspace.StyleWorkspaceDir}},
		Project:   config.Project{Domain: "personal", Name: "missing-state"},
		LookupEnv: func(string) string { return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range rep.Files {
		if strings.Contains(f, "state.json") {
			t.Fatalf("files must not list missing state.json: %v", rep.Files)
		}
	}
	var stateLayer *placement.Contribution
	for i := range rep.Contributions {
		if rep.Contributions[i].Layer == placement.LayerState {
			stateLayer = &rep.Contributions[i]
			break
		}
	}
	if stateLayer == nil {
		t.Fatal("expected state contribution row")
	}
	if stateLayer.Path != "" {
		t.Fatalf("missing binding must not set Path: %+v", stateLayer)
	}
	if stateLayer.Applies {
		t.Fatal("missing binding must not apply")
	}
	if !strings.Contains(stateLayer.Note, "no binding") {
		t.Fatalf("note: %q", stateLayer.Note)
	}
}

func TestExplainStateBoundInReport(t *testing.T) {
	root := t.TempDir()
	t.Setenv("YERK__STATE_DIR", root)
	if _, err := state.BindStyle("personal", "x", workspace.StyleWorkspaceDir); err != nil {
		t.Fatal(err)
	}
	rep, err := placement.Explain(placement.Input{
		Host:      config.Config{Workspace: config.Workspace{Style: workspace.StyleProjectDir}},
		Project:   config.Project{Domain: "personal", Name: "x"},
		LookupEnv: func(string) string { return "" },
	})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Effective.Bound || rep.Effective.Style != workspace.StyleWorkspaceDir {
		t.Fatalf("eff=%+v", rep.Effective)
	}
	if len(rep.Effective.Warnings) == 0 {
		t.Fatal("expected ambient drift warning")
	}
	found := false
	for _, c := range rep.Contributions {
		if c.Layer == placement.LayerState && c.Applies && c.Value == workspace.StyleWorkspaceDir {
			found = true
		}
	}
	if !found {
		t.Fatalf("state contribution missing: %+v", rep.Contributions)
	}
}
