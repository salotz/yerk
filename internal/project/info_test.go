package project_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/project"
)

func TestProjectAndReplicaInfo(t *testing.T) {
	root := t.TempDir()
	ws := filepath.Join(root, "devel", "yerk")
	rep := filepath.Join(ws, "main")
	if err := os.MkdirAll(filepath.Join(rep, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Projects: hostProjects(root,
			"personal/yerk", "devel/yerk",
		),
	}
	r, err := project.NewResolver(cfg, fakeGit{})
	if err != nil {
		t.Fatal(err)
	}
	p := config.Project{
		Name: "yerk", Domain: "personal", Remote: "git@example.com:a/yerk.git",
		DefaultReplica: "main", Tags: []string{"devel"},
	}

	pi, err := r.ProjectInfo(p)
	if err != nil {
		t.Fatal(err)
	}
	if pi.Kind != api.KindProjectInfo || pi.APIVersion != api.APIVersion {
		t.Fatalf("meta: %+v", pi)
	}
	if pi.URI != "yerk://personal/yerk" {
		t.Fatalf("uri %q", pi.URI)
	}
	if pi.WorkspacePath != ws {
		t.Fatalf("ws %q want %q", pi.WorkspacePath, ws)
	}
	if pi.WorkspacePresence != api.PresencePresent {
		t.Fatalf("presence %q", pi.WorkspacePresence)
	}
	if pi.Placement == nil || pi.Placement.Style != "workspace-dir" {
		t.Fatalf("placement %+v", pi.Placement)
	}

	ri, err := r.ReplicaInfo(p, "main")
	if err != nil {
		t.Fatal(err)
	}
	if ri.URI != "yerk://personal/yerk/main" || ri.Path != rep {
		t.Fatalf("replica %+v", ri)
	}
	if ri.Presence != api.PresencePresent {
		t.Fatalf("presence %q", ri.Presence)
	}
	if _, err := r.ReplicaInfo(p, ""); err == nil {
		t.Fatal("empty replica should error")
	}
}

func TestLookupPathWalkUp(t *testing.T) {
	root := t.TempDir()
	ws := filepath.Join(root, "devel", "yerk")
	rep := filepath.Join(ws, "main")
	deep := filepath.Join(rep, "internal", "cli")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(rep, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	otherWS := filepath.Join(root, "devel", "other")
	if err := os.MkdirAll(otherWS, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Projects: hostProjects(root,
			"personal/yerk", "devel/yerk",
			"personal/other", "devel/other",
		),
	}
	r, err := project.NewResolver(cfg, fakeGit{})
	if err != nil {
		t.Fatal(err)
	}
	projects := []config.Project{
		{Name: "yerk", Domain: "personal", Remote: "x", DefaultReplica: "main"},
		{Name: "other", Domain: "personal", Remote: "y"},
	}

	// Deep path under replica → replica info.
	got, err := r.LookupPath(projects, deep, project.LookupAny)
	if err != nil {
		t.Fatal(err)
	}
	if got.Replica == nil || got.Project != nil {
		t.Fatalf("want replica: %+v", got)
	}
	if got.Replica.Replica != "main" || got.Replica.MatchedPath != deep {
		t.Fatalf("replica %+v", got.Replica)
	}

	// Workspace root → project info.
	got, err = r.LookupPath(projects, ws, project.LookupAny)
	if err != nil {
		t.Fatal(err)
	}
	if got.Project == nil || got.Replica != nil {
		t.Fatalf("want project: %+v", got)
	}
	if got.Project.Name != "yerk" {
		t.Fatalf("name %q", got.Project.Name)
	}

	// project lookup under replica still returns project.
	got, err = r.LookupPath(projects, deep, project.LookupProject)
	if err != nil {
		t.Fatal(err)
	}
	if got.Project == nil || got.Project.Name != "yerk" {
		t.Fatalf("project lookup: %+v", got)
	}

	// replica lookup on workspace root errors.
	if _, err := r.LookupPath(projects, ws, project.LookupReplica); err == nil {
		t.Fatal("replica lookup on workspace should error")
	} else if !strings.Contains(err.Error(), "not under a known replica") {
		t.Fatalf("got %v", err)
	}

	// Unknown path.
	if _, err := r.LookupPath(projects, filepath.Join(root, "nowhere"), project.LookupAny); err == nil {
		t.Fatal("unknown path should error")
	}
}

func TestLookupLongestRootWins(t *testing.T) {
	root := t.TempDir()
	// Nested workspaces: outer and inner projects.
	outer := filepath.Join(root, "tree")
	inner := filepath.Join(outer, "inner")
	if err := os.MkdirAll(filepath.Join(inner, "main", "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Workspace: config.Workspace{Style: "workspace-dir"},
		Projects: []config.HostProject{
			{Domain: "personal", Name: "outer", Path: outer},
			{Domain: "personal", Name: "inner", Path: inner},
		},
	}
	r, err := project.NewResolver(cfg, fakeGit{})
	if err != nil {
		t.Fatal(err)
	}
	projects := []config.Project{
		{Name: "outer", Domain: "personal"},
		{Name: "inner", Domain: "personal"},
	}
	path := filepath.Join(inner, "main", "src")
	got, err := r.LookupPath(projects, path, project.LookupAny)
	if err != nil {
		t.Fatal(err)
	}
	if got.Replica == nil || got.Replica.Project != "inner" {
		t.Fatalf("want inner replica, got %+v", got)
	}
}
