package project_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/project"
)

func TestToolContext(t *testing.T) {
	t.Setenv("YERK__CONFIG_DIR", filepath.Join(t.TempDir(), "cfg"))
	t.Setenv("YERK__STATE_DIR", filepath.Join(t.TempDir(), "state"))
	tc, err := project.ToolContext()
	if err != nil {
		t.Fatal(err)
	}
	if tc.Kind != api.KindToolContext || tc.APIVersion != api.APIVersion {
		t.Fatalf("meta: %+v", tc)
	}
	if tc.Paths.ConfigDir == "" || tc.Paths.StateDir == "" {
		t.Fatalf("paths: %+v", tc.Paths)
	}
	if len(tc.Vocabulary) == 0 || len(tc.Commands) == 0 {
		t.Fatalf("empty vocabulary/commands")
	}
}

func TestDirContextReplicaPath(t *testing.T) {
	isolateState(t)
	root := t.TempDir()
	ws := filepath.Join(root, "devel", "yerk")
	rep := filepath.Join(ws, "main")
	deep := filepath.Join(rep, "internal", "cli")
	if err := os.MkdirAll(filepath.Join(rep, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Projects:  hostProjects(root, "personal/yerk", "devel/yerk"),
	}
	r, err := project.NewResolver(cfg, fakeGit{})
	if err != nil {
		t.Fatal(err)
	}
	projects := []config.Project{
		{Name: "yerk", Domain: "personal", Remote: "x", DefaultReplica: "main"},
	}
	dc, err := r.DirContext(context.Background(), projects, deep, project.DirContextOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if dc.Kind != api.KindDirContext || dc.Matched != "replica" {
		t.Fatalf("%+v", dc)
	}
	if dc.Replica == nil || dc.Replica.Replica != "main" {
		t.Fatalf("replica: %+v", dc.Replica)
	}
	if dc.Project == nil || dc.Project.Name != "yerk" {
		t.Fatalf("project: %+v", dc.Project)
	}
	if dc.EffectiveStyle != "workspace-dir" {
		t.Fatalf("style %q", dc.EffectiveStyle)
	}
	if dc.StatusOverall == nil {
		t.Fatal("expected status overall")
	}
	found := false
	for _, n := range dc.LiveReplicas {
		if n == "main" {
			found = true
		}
	}
	if !found {
		t.Fatalf("liveReplicas %v", dc.LiveReplicas)
	}
}

func TestDirContextUnknownPath(t *testing.T) {
	isolateState(t)
	root := t.TempDir()
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Projects:  hostProjects(root, "personal/yerk", "devel/yerk"),
	}
	r, err := project.NewResolver(cfg, fakeGit{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.DirContext(context.Background(), []config.Project{
		{Name: "yerk", Domain: "personal", Remote: "x"},
	}, filepath.Join(root, "nowhere"), project.DirContextOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not under") && !strings.Contains(err.Error(), "known") {
		t.Fatalf("got %v", err)
	}
}
