package workspace_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/workspace"
)

func TestWorkspaceDirStyleAbsolute(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Config{
		Workspace: config.Workspace{Style: workspace.StyleWorkspaceDir},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Path: "/tree/personal/devel/yerk"}

	proj, err := layout.ProjectDir(p)
	if err != nil {
		t.Fatal(err)
	}
	if proj != "/tree/personal/devel/yerk" {
		t.Fatalf("project dir %q", proj)
	}

	got, err := layout.ReplicaDir(p, "main")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tree/personal/devel/yerk/main" {
		t.Fatalf("replica main: got %q", got)
	}
	got2, err := layout.ReplicaDir(p, "feature")
	if err != nil {
		t.Fatal(err)
	}
	if got2 != "/tree/personal/devel/yerk/feature" {
		t.Fatalf("replica feature: got %q", got2)
	}
}

func TestWorkspaceDirStyleRelativeDomain(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Config{
		Workspace: config.Workspace{Style: workspace.StyleWorkspaceDir},
		Domains:   map[string]string{"personal": "/tree/personal"},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal", Path: "devel/yerk"}

	proj, err := layout.ProjectDir(p)
	if err != nil {
		t.Fatal(err)
	}
	if proj != "/tree/personal/devel/yerk" {
		t.Fatalf("project dir %q", proj)
	}
	got, err := layout.ReplicaDir(p, "main")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tree/personal/devel/yerk/main" {
		t.Fatalf("replica: got %q", got)
	}
}

func TestProjectDirStyle(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Config{
		Workspace: config.Workspace{Style: workspace.StyleProjectDir},
		Domains:   map[string]string{"personal": "/tree/personal"},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal", Path: "devel/yerk"}
	got, err := layout.ReplicaDir(p, "main")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tree/personal/devel/yerk__main" {
		t.Fatalf("got %q", got)
	}
	proj, err := layout.ProjectDir(p)
	if err != nil || proj != "/tree/personal/devel/yerk" {
		t.Fatalf("project dir %q %v", proj, err)
	}
}

func TestReplicaDirMissingPath(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = layout.ReplicaDir(config.Project{Name: "yerk"}, "main")
	if err == nil {
		t.Fatal("expected error without path")
	}
	if !strings.Contains(err.Error(), "no path") {
		t.Fatalf("err: %v", err)
	}
}

func TestReplicaDirRelativeNeedsDomainRoot(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = layout.ReplicaDir(config.Project{Name: "x", Domain: "personal", Path: "devel/x"}, "main")
	if err == nil {
		t.Fatal("expected missing domain root error")
	}
	if !strings.Contains(err.Error(), "no root") {
		t.Fatalf("err: %v", err)
	}

	_, err = layout.ReplicaDir(config.Project{Name: "x", Path: "devel/x"}, "main")
	if err == nil {
		t.Fatal("expected empty domain error")
	}
	if !strings.Contains(err.Error(), "domain is empty") {
		t.Fatalf("err: %v", err)
	}
}

func TestReplicaDirEmptyReplica(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = layout.ReplicaDir(config.Project{Name: "x", Path: "/abs/x"}, "")
	if err == nil {
		t.Fatal("expected empty replica error")
	}
}

func TestUnknownStyle(t *testing.T) {
	t.Parallel()
	_, err := workspace.NewLayout(config.Config{Workspace: config.Workspace{Style: "nope"}})
	if err == nil {
		t.Fatal("expected unknown style error")
	}
}

func TestEnsureDir(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	dir := filepath.Join(base, "devel", "yerk")
	if err := workspace.EnsureDir(dir); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		t.Fatalf("dir: %v %+v", err, st)
	}
}

func TestEnsureParents(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	replica := filepath.Join(base, "a", "b", "repo")
	if err := workspace.EnsureParents(replica); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(base, "a", "b")
	st, err := os.Stat(parent)
	if err != nil || !st.IsDir() {
		t.Fatalf("parent: %v %+v", err, st)
	}
}

func TestEnsureReplicaDir(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	replica := filepath.Join(base, "repo")
	if err := workspace.EnsureReplicaDir(replica); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(replica)
	if err != nil || !st.IsDir() {
		t.Fatalf("replica: %v %+v", err, st)
	}
}

func TestExpandTildeInDomainRoot(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	layout, err := workspace.NewLayout(config.Config{
		Workspace: config.Workspace{Style: workspace.StyleWorkspaceDir},
		Domains:   map[string]string{"personal": "~/tree/personal"},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal", Path: "devel/yerk"}
	got, err := layout.ProjectDir(p)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "tree/personal/devel/yerk")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
