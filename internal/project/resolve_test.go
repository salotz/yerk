package project_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/gitcmd"
	"github.com/salotz/yerk/internal/project"
)

// fakeGit implements gitcmd.Runner for unit tests.
type fakeGit struct {
	branch string
	probe  gitcmd.ProbeResult
	err    error
}

func (f fakeGit) DefaultBranch(ctx context.Context, remote string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if f.branch != "" {
		return f.branch, nil
	}
	return "main", nil
}

func (f fakeGit) Clone(ctx context.Context, remote, dest, branch string) error {
	return nil
}

func (f fakeGit) WorktreeAdd(ctx context.Context, mainRepo, dest, branch string) error {
	return nil
}

func (f fakeGit) EnsureBranch(ctx context.Context, repoPath, branch string) error {
	return nil
}

func (f fakeGit) Probe(ctx context.Context, repoPath string) (gitcmd.ProbeResult, error) {
	if f.err != nil {
		return gitcmd.ProbeResult{}, f.err
	}
	return f.probe, nil
}

func hostProjects(root string, pairs ...string) []config.HostProject {
	var out []config.HostProject
	for i := 0; i+1 < len(pairs); i += 2 {
		id := pairs[i] // domain/name
		domain, name, ok := strings.Cut(id, "/")
		if !ok {
			panic("bad id " + id)
		}
		out = append(out, config.HostProject{
			Domain: domain, Name: name, Path: filepath.Join(root, pairs[i+1]),
		})
	}
	return out
}

// isolateState points YERK__STATE_DIR at a temp tree so tests never touch
// the operator's real XDG state (permission / pollution).
func isolateState(t *testing.T) {
	t.Helper()
	t.Setenv("YERK__STATE_DIR", filepath.Join(t.TempDir(), "state"))
}

func TestResolveReplicaMethod(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cli, cat, want string
		err            bool
	}{
		{"", "", project.MethodWorktree, false},
		{"", "clone", project.MethodClone, false},
		{"worktree", "clone", project.MethodWorktree, false},
		{"clone", "", project.MethodClone, false},
		{"nope", "", "", true},
	}
	for _, tc := range cases {
		got, err := project.ResolveReplicaMethod(tc.cli, tc.cat)
		if tc.err {
			if err == nil {
				t.Fatalf("cli=%q cat=%q: want err", tc.cli, tc.cat)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Fatalf("cli=%q cat=%q: got %q %v want %q", tc.cli, tc.cat, got, err, tc.want)
		}
	}
}

func TestCreateReplicaWorktreeRequiresMain(t *testing.T) {
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
	p := config.Project{Name: "yerk", Domain: "personal", Remote: "x", DefaultReplica: "main"}
	_, err = r.CreateReplica(context.Background(), p, "feat", project.CreateReplicaOptions{})
	if err == nil {
		t.Fatal("expected main missing error")
	}
	if !strings.Contains(err.Error(), "main replica") || !strings.Contains(err.Error(), "materialize") {
		t.Fatalf("got %v", err)
	}
}

func TestCreateReplicaRefusePresent(t *testing.T) {
	isolateState(t)
	root := t.TempDir()
	feat := filepath.Join(root, "devel", "yerk", "feat")
	if err := os.MkdirAll(filepath.Join(feat, ".git"), 0o755); err != nil {
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
	p := config.Project{Name: "yerk", Domain: "personal", Remote: "x", DefaultReplica: "main"}
	_, err = r.CreateReplica(context.Background(), p, "feat", project.CreateReplicaOptions{Method: project.MethodClone})
	if err == nil {
		t.Fatal("expected already present error")
	}
	if !strings.Contains(err.Error(), "already present") {
		t.Fatalf("got %v", err)
	}
}

func TestStatusPresenceOnly(t *testing.T) {
	isolateState(t)
	root := t.TempDir()
	yerkWS := filepath.Join(root, "devel", "yerk")
	yerkReplica := filepath.Join(yerkWS, "main")
	if err := os.MkdirAll(filepath.Join(yerkReplica, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Projects: hostProjects(root,
			"personal/yerk", "devel/yerk",
			"personal/bimhaw", "devel/bimhaw",
		),
	}
	r, err := project.NewResolver(cfg, fakeGit{
		probe: gitcmd.ProbeResult{Clean: true, Branch: "main", NoUpstream: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	cat := config.Catalog{Projects: []config.Project{
		{Name: "yerk", Domain: "personal", Remote: "x", DefaultReplica: "main"},
		{Name: "bimhaw", Domain: "personal", Remote: "y", DefaultReplica: "main"},
	}}
	rows, err := r.Status(context.Background(), cat.Projects, project.StatusOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows %d", len(rows))
	}
	if rows[0].Kind != api.KindReplicaStatus || rows[0].APIVersion != api.APIVersion {
		t.Fatalf("type meta: %+v", rows[0])
	}
	if rows[0].Project != "yerk" || rows[0].Presence != api.PresencePresent || rows[0].Change != "-" {
		t.Fatalf("yerk: %+v", rows[0])
	}
	if rows[0].Path != yerkReplica {
		t.Fatalf("yerk path: got %q want %q", rows[0].Path, yerkReplica)
	}
	if rows[1].Project != "bimhaw" || rows[1].Presence != api.PresenceMissing {
		t.Fatalf("bimhaw: %+v", rows[1])
	}
	wantMissing := filepath.Join(root, "devel", "bimhaw", "main")
	if rows[1].Path != wantMissing {
		t.Fatalf("bimhaw path: got %q want %q", rows[1].Path, wantMissing)
	}

	rows, err = r.Status(context.Background(), cat.Projects, project.StatusOptions{Git: true})
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Change != "clean no-upstream" {
		t.Fatalf("change %q", rows[0].Change)
	}
	if rows[0].Branch != "main" {
		t.Fatalf("branch %q", rows[0].Branch)
	}
}

func TestProjectStatus(t *testing.T) {
	isolateState(t)
	root := t.TempDir()
	yerkWS := filepath.Join(root, "devel", "yerk")
	yerkReplica := filepath.Join(yerkWS, "main")
	if err := os.MkdirAll(filepath.Join(yerkReplica, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Projects:  hostProjects(root, "personal/yerk", "devel/yerk"),
	}
	r, err := project.NewResolver(cfg, fakeGit{
		probe: gitcmd.ProbeResult{Clean: true, Branch: "main", NoUpstream: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{
		Name: "yerk", Domain: "personal", Remote: "x",
		DefaultReplica: "main", Tags: []string{"devel"},
	}
	st, err := r.ProjectStatus(context.Background(), p, project.StatusOptions{Git: true})
	if err != nil {
		t.Fatal(err)
	}
	if st.Kind != api.KindProjectStatus || st.Name != "yerk" {
		t.Fatalf("project status: %+v", st)
	}
	if st.WorkspacePath != yerkWS {
		t.Fatalf("workspace path: got %q want %q", st.WorkspacePath, yerkWS)
	}
	if st.WorkspacePresence != api.PresencePresent {
		t.Fatalf("workspace presence: %s", st.WorkspacePresence)
	}
	if st.DefaultReplica == nil {
		t.Fatal("expected default replica summary")
	}
	if st.DefaultReplica.Name != "main" || st.DefaultReplica.Path != yerkReplica {
		t.Fatalf("default replica: %+v", st.DefaultReplica)
	}
	if st.DefaultReplica.Presence != api.PresencePresent {
		t.Fatalf("presence: %s", st.DefaultReplica.Presence)
	}
	if st.DefaultReplica.Change != "clean no-upstream" {
		t.Fatalf("change: %q", st.DefaultReplica.Change)
	}
	if st.Overall == nil || st.Overall.Presence != "all-present" {
		t.Fatalf("overall: %+v", st.Overall)
	}
	if st.Overall.Change != "clean" && st.Overall.Change != "no-upstream" && st.Overall.Change != "clean no-upstream" {
		// rollup drops lone clean when other tokens exist; no-upstream alone is fine
		if st.Overall.Change != "no-upstream" {
			t.Fatalf("overall change: %q", st.Overall.Change)
		}
	}
	if len(st.Replicas) != 1 || st.Replicas[0].Replica != "main" {
		t.Fatalf("replicas: %+v", st.Replicas)
	}

	list, err := r.ProjectStatuses(context.Background(), []config.Project{p}, project.StatusOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].DefaultReplica.Change != "-" {
		t.Fatalf("presence-only list: %+v", list)
	}
}

func TestProjectStatusMultipleReplicas(t *testing.T) {
	isolateState(t)
	root := t.TempDir()
	ws := filepath.Join(root, "devel", "yerk")
	mainP := filepath.Join(ws, "main")
	featP := filepath.Join(ws, "feat")
	for _, d := range []string{mainP, featP} {
		if err := os.MkdirAll(filepath.Join(d, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Non-git sibling should be ignored.
	if err := os.MkdirAll(filepath.Join(ws, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Projects:  hostProjects(root, "personal/yerk", "devel/yerk"),
	}
	r, err := project.NewResolver(cfg, fakeGit{
		probe: gitcmd.ProbeResult{Dirty: true, Branch: "x", NoUpstream: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal", Remote: "x", DefaultReplica: "main"}
	st, err := r.ProjectStatus(context.Background(), p, project.StatusOptions{Git: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Replicas) != 2 {
		t.Fatalf("want 2 replicas, got %+v", st.Replicas)
	}
	if st.Overall == nil || st.Overall.Presence != "all-present" || st.Overall.PresentCount != 2 {
		t.Fatalf("overall %+v", st.Overall)
	}
	// Dirty should surface in overall; clean should not dominate.
	if !strings.Contains(st.Overall.Change, "dirty") {
		t.Fatalf("overall change should surface dirty: %q", st.Overall.Change)
	}
	if strings.Contains(st.Overall.Change, "clean") {
		t.Fatalf("overall should not keep clean when dirty: %q", st.Overall.Change)
	}
}

func TestReplicaStatusNamed(t *testing.T) {
	isolateState(t)
	root := t.TempDir()
	feature := filepath.Join(root, "devel", "yerk", "feature")
	if err := os.MkdirAll(filepath.Join(feature, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Projects:  hostProjects(root, "personal/yerk", "devel/yerk"),
	}
	r, err := project.NewResolver(cfg, fakeGit{
		probe: gitcmd.ProbeResult{Dirty: true, Branch: "feature", NoUpstream: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal", Remote: "x", DefaultReplica: "main"}
	row, err := r.ReplicaStatus(context.Background(), p, "feature", project.StatusOptions{Git: true})
	if err != nil {
		t.Fatal(err)
	}
	if row.Replica != "feature" || row.Path != feature {
		t.Fatalf("%+v", row)
	}
	if row.Presence != api.PresencePresent || row.Change != "dirty no-upstream" {
		t.Fatalf("%+v", row)
	}
}

func TestStatusRequiresProjectPath(t *testing.T) {
	isolateState(t)
	cfg := config.Config{Workspace: config.Workspace{Style: "workspace-dir"}}
	r, err := project.NewResolver(cfg, fakeGit{})
	if err != nil {
		t.Fatal(err)
	}
	cat := config.Catalog{Projects: []config.Project{
		{Name: "yerk", Domain: "personal", Remote: "x", DefaultReplica: "main"},
	}}
	_, err = r.Status(context.Background(), cat.Projects, project.StatusOptions{})
	if err == nil || !strings.Contains(err.Error(), "[[projects]]") {
		t.Fatalf("expected missing [[projects]] error, got %v", err)
	}
}

func TestDefaultReplicaOverride(t *testing.T) {
	t.Parallel()
	cfg := config.Config{Workspace: config.Workspace{Style: "workspace-dir"}}
	r, err := project.NewResolver(cfg, fakeGit{branch: "develop"})
	if err != nil {
		t.Fatal(err)
	}
	name, err := r.DefaultReplicaName(context.Background(), config.Project{
		Name: "x", DefaultReplica: "trunk", Remote: "url",
	}, true)
	if err != nil || name != "trunk" {
		t.Fatalf("%q %v", name, err)
	}
	name, err = r.DefaultReplicaName(context.Background(), config.Project{
		Name: "x", Remote: "url",
	}, true)
	if err != nil || name != "develop" {
		t.Fatalf("%q %v", name, err)
	}
	name, err = r.DefaultReplicaName(context.Background(), config.Project{Name: "x"}, false)
	if err != nil || name != "main" {
		t.Fatalf("%q %v", name, err)
	}
}
