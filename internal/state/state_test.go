package state_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/state"
)

func TestProjectStateRoundTrip(t *testing.T) {
	root := t.TempDir()
	t.Setenv("YERK__STATE_DIR", root)

	written, err := state.BindStyle("personal", "wumpus", "workspace-dir")
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("expected first bind to write")
	}
	path, err := state.ProjectFile("personal", "wumpus")
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filepath.Join(root, "projects", "personal", "wumpus", "state.json")
	if path != wantPath {
		t.Fatalf("path %q want %q", path, wantPath)
	}
	st, ok, err := state.LoadProject("personal", "wumpus")
	if err != nil || !ok {
		t.Fatalf("load: ok=%v err=%v", ok, err)
	}
	if st.APIVersion != api.APIVersion {
		t.Fatalf("apiVersion %q", st.APIVersion)
	}
	if st.WorkspaceStyle != "workspace-dir" {
		t.Fatalf("style %q", st.WorkspaceStyle)
	}
	if st.BoundAt == "" || st.UpdatedAt == "" {
		t.Fatalf("timestamps missing: %+v", st)
	}

	written, err = state.BindStyle("personal", "wumpus", "project-dir")
	if err != nil {
		t.Fatal(err)
	}
	if written {
		t.Fatal("second bind should be no-op")
	}
	st2, _, err := state.LoadProject("personal", "wumpus")
	if err != nil {
		t.Fatal(err)
	}
	if st2.WorkspaceStyle != "workspace-dir" {
		t.Fatalf("style changed on no-op bind: %q", st2.WorkspaceStyle)
	}
}

func TestUpdateStyleOverwrite(t *testing.T) {
	root := t.TempDir()
	t.Setenv("YERK__STATE_DIR", root)

	ur, err := state.UpdateStyle("personal", "wumpus", "workspace-dir")
	if err != nil {
		t.Fatal(err)
	}
	if !ur.Created || !ur.Changed || ur.Style != "workspace-dir" {
		t.Fatalf("%+v", ur)
	}
	ur, err = state.UpdateStyle("personal", "wumpus", "project-dir")
	if err != nil {
		t.Fatal(err)
	}
	if ur.Created || !ur.Changed || ur.Previous != "workspace-dir" || ur.Style != "project-dir" {
		t.Fatalf("%+v", ur)
	}
	st, ok, err := state.LoadProject("personal", "wumpus")
	if err != nil || !ok {
		t.Fatalf("load: %v ok=%v", err, ok)
	}
	if st.WorkspaceStyle != "project-dir" {
		t.Fatalf("style %q", st.WorkspaceStyle)
	}
	if st.BoundAt == "" {
		t.Fatal("BoundAt should be preserved/set")
	}
	// No-op same style.
	ur, err = state.UpdateStyle("personal", "wumpus", "project-dir")
	if err != nil {
		t.Fatal(err)
	}
	if ur.Changed || ur.Created {
		t.Fatalf("same style should be unchanged: %+v", ur)
	}
}

func TestDirUsesXDGStateHome(t *testing.T) {
	base := t.TempDir()
	t.Setenv("YERK__STATE_DIR", "")
	t.Setenv("XDG_STATE_HOME", base)
	// Clear explicit override if empty string still set — use unset
	os.Unsetenv("YERK__STATE_DIR")
	dir, err := state.Dir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(base, "yerk")
	if dir != want {
		t.Fatalf("got %q want %q", dir, want)
	}
}
