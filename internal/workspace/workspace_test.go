package workspace_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/workspace"
)

func hostCfg(style string, rows ...config.HostProject) config.Config {
	return config.Config{
		Workspace: config.Workspace{Style: style},
		Projects:  rows,
	}
}

func hostCfgDomains(style string, domains map[string]string, rows ...config.HostProject) config.Config {
	return config.Config{
		Workspace: config.Workspace{Style: style},
		Domains:   domains,
		Projects:  rows,
	}
}

func TestWorkspaceDirStyle(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(hostCfg(workspace.StyleWorkspaceDir,
		config.HostProject{Name: "yerk", Domain: "personal", Path: "/tree/personal/devel/yerk"},
	))
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal"}

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

func TestProjectDirStyle(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(hostCfg(workspace.StyleProjectDir,
		config.HostProject{Name: "yerk", Domain: "personal", Path: "/tree/personal/devel/yerk"},
	))
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal"}
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

func TestReplicaDirMissingPlacement(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = layout.ReplicaDir(config.Project{Name: "yerk", Domain: "personal"}, "main")
	if err == nil {
		t.Fatal("expected error without domains or host path")
	}
	if !strings.Contains(err.Error(), "[domains.personal]") {
		t.Fatalf("err: %v", err)
	}
}

func TestReplicaDirDomainDefault(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(hostCfgDomains(workspace.StyleWorkspaceDir,
		map[string]string{"personal": "/tree/personal/devel"},
	))
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal"}
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

func TestReplicaDirRelativeHostPathNeedsDomain(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(hostCfg(workspace.StyleWorkspaceDir,
		config.HostProject{Name: "x", Domain: "personal", Path: "devel/x"},
	))
	if err != nil {
		t.Fatal(err)
	}
	_, err = layout.ReplicaDir(config.Project{Name: "x", Domain: "personal"}, "main")
	if err == nil {
		t.Fatal("expected relative path error without domain root")
	}
	if !strings.Contains(err.Error(), "relative") {
		t.Fatalf("err: %v", err)
	}

	layout, err = workspace.NewLayout(hostCfgDomains(workspace.StyleWorkspaceDir,
		map[string]string{"personal": "/tree/personal"},
		config.HostProject{Name: "x", Domain: "personal", Path: "devel/x"},
	))
	if err != nil {
		t.Fatal(err)
	}
	got, err := layout.ProjectDir(config.Project{Name: "x", Domain: "personal"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tree/personal/devel/x" {
		t.Fatalf("got %q", got)
	}
}

func TestReplicaDirEmptyReplica(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(hostCfg("",
		config.HostProject{Name: "x", Domain: "personal", Path: "/abs/x"},
	))
	if err != nil {
		t.Fatal(err)
	}
	_, err = layout.ReplicaDir(config.Project{Name: "x", Domain: "personal"}, "")
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

func TestListLiveReplicasWorkspaceDir(t *testing.T) {
	root := t.TempDir()
	ws := filepath.Join(root, "yerk")
	for _, name := range []string{"main", "feat"} {
		if err := os.MkdirAll(filepath.Join(ws, name, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(ws, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	layout, err := workspace.NewLayout(hostCfg(workspace.StyleWorkspaceDir,
		config.HostProject{Name: "yerk", Domain: "personal", Path: ws},
	))
	if err != nil {
		t.Fatal(err)
	}
	got, err := layout.ListLiveReplicas(config.Project{Name: "yerk", Domain: "personal"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "feat" || got[1] != "main" {
		t.Fatalf("got %v", got)
	}
}

func TestListLiveReplicasProjectDir(t *testing.T) {
	root := t.TempDir()
	ws := filepath.Join(root, "yerk") // project workspace path (not a replica)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, leaf := range []string{"yerk__main", "yerk__sess"} {
		if err := os.MkdirAll(filepath.Join(root, leaf, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "other__x", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	layout, err := workspace.NewLayout(hostCfg(workspace.StyleProjectDir,
		config.HostProject{Name: "yerk", Domain: "personal", Path: ws},
	))
	if err != nil {
		t.Fatal(err)
	}
	got, err := layout.ListLiveReplicas(config.Project{Name: "yerk", Domain: "personal"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "main" || got[1] != "sess" {
		t.Fatalf("got %v", got)
	}
}

func TestExpandTildeInHostPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	layout, err := workspace.NewLayout(hostCfg(workspace.StyleWorkspaceDir,
		config.HostProject{Name: "yerk", Domain: "personal", Path: "~/tree/personal/devel/yerk"},
	))
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal"}
	got, err := layout.ProjectDir(p)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "tree/personal/devel/yerk")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
