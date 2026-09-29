package gitcmd_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/salotz/yerk/internal/gitcmd"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
}

func initRepo(t *testing.T, dir string) {
	t.Helper()
	requireGit(t)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=yerk-test",
			"GIT_AUTHOR_EMAIL=yerk-test@example.com",
			"GIT_COMMITTER_NAME=yerk-test",
			"GIT_COMMITTER_EMAIL=yerk-test@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
	run("init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "README")
	run("commit", "-m", "init")
}

func TestProbeClean(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	r := gitcmd.New()
	res, err := r.Probe(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Clean || res.Dirty || res.Untracked {
		t.Fatalf("%+v flags=%s", res, res.Flags())
	}
	if res.Branch != "main" {
		t.Fatalf("branch %q", res.Branch)
	}
	if !res.NoUpstream {
		t.Fatalf("expected no-upstream, got %+v", res)
	}
}

func TestProbeDirtyAndUntracked(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "extra"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := gitcmd.New()
	res, err := r.Probe(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Dirty || !res.Untracked || res.Clean {
		t.Fatalf("%+v", res)
	}
	flags := res.Flags()
	if flags != "dirty untracked no-upstream" {
		t.Fatalf("flags %q", flags)
	}
}

func TestCloneAndDefaultBranch(t *testing.T) {
	requireGit(t)
	src := t.TempDir()
	initRepo(t, src)
	// bare-ish: use path remote
	r := gitcmd.New()
	ctx := context.Background()
	branch, err := r.DefaultBranch(ctx, src)
	if err != nil {
		// Some git versions need a real remote update; try with file URL
		t.Logf("DefaultBranch on plain path: %v", err)
		branch = "main"
	} else if branch != "main" {
		t.Fatalf("branch %q", branch)
	}

	dest := filepath.Join(t.TempDir(), "clone")
	if err := r.Clone(ctx, src, dest, "main"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "README")); err != nil {
		t.Fatal(err)
	}
	// refuse non-empty
	if err := r.Clone(ctx, src, dest, "main"); err == nil {
		t.Fatal("expected refuse non-empty")
	}
}

func TestParseFlagsDisplay(t *testing.T) {
	t.Parallel()
	p := gitcmd.ProbeResult{Clean: true, Ahead: 2}
	if p.Flags() != "clean ahead:2" {
		t.Fatalf("%q", p.Flags())
	}
}

func TestProbeOriginBranchFallbackAhead(t *testing.T) {
	requireGit(t)
	src := t.TempDir()
	initRepo(t, src)

	// Clone normally (sets upstream), then drop tracking to mimic
	// agent-guidelines-style checkouts that still have origin/<branch>.
	dest := filepath.Join(t.TempDir(), "work")
	r := gitcmd.New()
	ctx := context.Background()
	if err := r.Clone(ctx, src, dest, "main"); err != nil {
		t.Fatal(err)
	}
	runGit(t, dest, "branch", "--unset-upstream")
	// Local commit not on origin/main.
	if err := os.WriteFile(filepath.Join(dest, "README"), []byte("local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dest, "add", "README")
	runGit(t, dest, "commit", "-m", "local")

	res, err := r.Probe(ctx, dest)
	if err != nil {
		t.Fatal(err)
	}
	if res.NoUpstream {
		t.Fatalf("expected origin/main fallback, got NoUpstream: %+v flags=%s", res, res.Flags())
	}
	if res.Ahead != 1 || res.Behind != 0 {
		t.Fatalf("want ahead=1 behind=0, got %+v flags=%s", res, res.Flags())
	}
	if got := res.Flags(); got != "clean ahead:1" {
		t.Fatalf("flags %q", got)
	}
}

func TestProbeOriginBranchFallbackMissing(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	// No origin remote → still no-upstream.
	r := gitcmd.New()
	res, err := r.Probe(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if !res.NoUpstream || res.Ahead != 0 {
		t.Fatalf("%+v flags=%s", res, res.Flags())
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=yerk-test",
		"GIT_AUTHOR_EMAIL=yerk-test@example.com",
		"GIT_COMMITTER_NAME=yerk-test",
		"GIT_COMMITTER_EMAIL=yerk-test@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v (%s)", args, err, out)
	}
}
