package presence_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/salotz/yerk/internal/presence"
)

func TestMissing(t *testing.T) {
	t.Parallel()
	got := presence.Classify(filepath.Join(t.TempDir(), "nope"))
	if got != presence.Missing {
		t.Fatalf("got %s", got)
	}
}

func TestInvalidEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	got := presence.Classify(dir)
	if got != presence.Invalid {
		t.Fatalf("got %s", got)
	}
}

func TestPresentGitDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := presence.Classify(dir)
	if got != presence.Present {
		t.Fatalf("got %s", got)
	}
}

func TestPresentGitFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: /tmp/somewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := presence.Classify(dir)
	if got != presence.Present {
		t.Fatalf("got %s", got)
	}
}

func TestInvalidFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "file")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := presence.Classify(path)
	if got != presence.Invalid {
		t.Fatalf("got %s", got)
	}
}
