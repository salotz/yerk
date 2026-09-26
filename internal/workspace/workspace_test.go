package workspace_test

import (
	"testing"

	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/workspace"
)

func TestReplicaDirWorkspaceStyle(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Workspace{
		Root:  "/tree/personal/devel",
		Style: workspace.StyleWorkspaceDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := layout.ReplicaDir("yerk", "main")
	if err != nil {
		t.Fatal(err)
	}
	want := "/tree/personal/devel/yerk/main"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReplicaDirProjectStyle(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Workspace{
		Root:  "/tree/personal/devel",
		Style: workspace.StyleProjectDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := layout.ReplicaDir("yerk", "main")
	if err != nil {
		t.Fatal(err)
	}
	want := "/tree/personal/devel/yerk__main"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestMissingRoot(t *testing.T) {
	t.Parallel()
	layout, err := workspace.NewLayout(config.Workspace{Style: workspace.StyleWorkspaceDir})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := layout.ReplicaDir("yerk", "main"); err == nil {
		t.Fatal("expected error for empty root")
	}
}
