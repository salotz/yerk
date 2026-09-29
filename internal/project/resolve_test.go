package project_test

import (
	"context"
	"os"
	"path/filepath"
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

func (f fakeGit) Probe(ctx context.Context, repoPath string) (gitcmd.ProbeResult, error) {
	if f.err != nil {
		return gitcmd.ProbeResult{}, f.err
	}
	return f.probe, nil
}

func TestStatusPresenceOnly(t *testing.T) {
	root := t.TempDir()
	domainRoot := filepath.Join(root, "personal")
	yerkWS := filepath.Join(domainRoot, "devel", "yerk")
	yerkReplica := filepath.Join(yerkWS, "main")
	if err := os.MkdirAll(filepath.Join(yerkReplica, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Domains:   map[string]string{"personal": domainRoot},
	}
	r, err := project.NewResolver(cfg, fakeGit{
		probe: gitcmd.ProbeResult{Clean: true, Branch: "main", NoUpstream: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	cat := config.Catalog{Projects: []config.Project{
		{Name: "yerk", Domain: "personal", Remote: "x", Path: "devel/yerk", DefaultReplica: "main"},
		{Name: "bimhaw", Domain: "personal", Remote: "y", Path: "devel/bimhaw", DefaultReplica: "main"},
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
	wantMissing := filepath.Join(domainRoot, "devel", "bimhaw", "main")
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
	root := t.TempDir()
	domainRoot := filepath.Join(root, "personal")
	yerkWS := filepath.Join(domainRoot, "devel", "yerk")
	yerkReplica := filepath.Join(yerkWS, "main")
	if err := os.MkdirAll(filepath.Join(yerkReplica, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Domains:   map[string]string{"personal": domainRoot},
	}
	r, err := project.NewResolver(cfg, fakeGit{
		probe: gitcmd.ProbeResult{Clean: true, Branch: "main", NoUpstream: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{
		Name: "yerk", Domain: "personal", Remote: "x",
		Path: "devel/yerk", DefaultReplica: "main", Tags: []string{"devel"},
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

	list, err := r.ProjectStatuses(context.Background(), []config.Project{p}, project.StatusOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].DefaultReplica.Change != "-" {
		t.Fatalf("presence-only list: %+v", list)
	}
}

func TestReplicaStatusNamed(t *testing.T) {
	root := t.TempDir()
	domainRoot := filepath.Join(root, "personal")
	feature := filepath.Join(domainRoot, "devel", "yerk", "feature")
	if err := os.MkdirAll(filepath.Join(feature, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Domains:   map[string]string{"personal": domainRoot},
	}
	r, err := project.NewResolver(cfg, fakeGit{
		probe: gitcmd.ProbeResult{Dirty: true, Branch: "feature", NoUpstream: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{Name: "yerk", Domain: "personal", Remote: "x", Path: "devel/yerk", DefaultReplica: "main"}
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

func TestStatusRequiresPath(t *testing.T) {
	cfg := config.Config{Workspace: config.Workspace{Style: "workspace-dir"}}
	r, err := project.NewResolver(cfg, fakeGit{})
	if err != nil {
		t.Fatal(err)
	}
	cat := config.Catalog{Projects: []config.Project{
		{Name: "yerk", Remote: "x", DefaultReplica: "main"},
	}}
	_, err = r.Status(context.Background(), cat.Projects, project.StatusOptions{})
	if err == nil {
		t.Fatal("expected missing path error")
	}
}

func TestStatusRelativeNeedsDomain(t *testing.T) {
	cfg := config.Config{Workspace: config.Workspace{Style: "workspace-dir"}}
	r, err := project.NewResolver(cfg, fakeGit{})
	if err != nil {
		t.Fatal(err)
	}
	cat := config.Catalog{Projects: []config.Project{
		{Name: "yerk", Domain: "personal", Remote: "x", Path: "devel/yerk", DefaultReplica: "main"},
	}}
	_, err = r.Status(context.Background(), cat.Projects, project.StatusOptions{})
	if err == nil {
		t.Fatal("expected missing domain root error")
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
