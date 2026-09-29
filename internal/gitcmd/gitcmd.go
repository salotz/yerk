// Package gitcmd is the git subprocess adapter for yerk.
//
// All host git interaction should go through Runner so tests can fake it and
// a future go-git backend can replace the implementation.
package gitcmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner talks to git. The zero value is not usable; use New or a test fake.
type Runner interface {
	// DefaultBranch returns the remote HEAD branch short name (e.g. "main").
	DefaultBranch(ctx context.Context, remote string) (string, error)
	// Clone clones remote into dest at branch. dest must not exist or be empty.
	Clone(ctx context.Context, remote, dest, branch string) error
	// Probe returns change-status flags for a present checkout.
	Probe(ctx context.Context, repoPath string) (ProbeResult, error)
}

// ProbeResult holds change-status flags for one replica.
type ProbeResult struct {
	Branch     string
	Clean      bool
	Dirty      bool
	Untracked  bool
	Ahead      int
	Behind     int
	NoUpstream bool
	SyncUnknown bool
	Error      string // non-empty if probe partially failed
}

// Flags returns space-separated change flags for display.
func (p ProbeResult) Flags() string {
	if p.Error != "" && !p.Dirty && !p.Untracked && p.Clean {
		return "error"
	}
	var parts []string
	if p.Dirty {
		parts = append(parts, "dirty")
	}
	if p.Untracked {
		parts = append(parts, "untracked")
	}
	if !p.Dirty && !p.Untracked && p.Error == "" {
		parts = append(parts, "clean")
	}
	if p.NoUpstream {
		parts = append(parts, "no-upstream")
	} else if p.SyncUnknown {
		parts = append(parts, "sync-unknown")
	} else {
		if p.Ahead > 0 {
			parts = append(parts, fmt.Sprintf("ahead:%d", p.Ahead))
		}
		if p.Behind > 0 {
			parts = append(parts, fmt.Sprintf("behind:%d", p.Behind))
		}
	}
	if p.Error != "" {
		parts = append(parts, "error")
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

// CLI implements Runner via the git binary on PATH.
type CLI struct {
	// Git is the executable name or path (default "git").
	Git string
	// Env optional extra env; nil means inherit.
	Env []string
}

// New returns a CLI runner using "git" on PATH.
func New() *CLI {
	return &CLI{Git: "git"}
}

func (c *CLI) gitBin() string {
	if c.Git == "" {
		return "git"
	}
	return c.Git
}

func (c *CLI) cmd(ctx context.Context, dir string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.gitBin(), args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if c.Env != nil {
		cmd.Env = c.Env
	}
	return cmd
}

func (c *CLI) run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := c.cmd(ctx, dir, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return out, fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return out, nil
}

// DefaultBranch resolves remote HEAD short name.
func (c *CLI) DefaultBranch(ctx context.Context, remote string) (string, error) {
	out, err := c.run(ctx, "", "ls-remote", "--symref", remote, "HEAD")
	if err != nil {
		return "", err
	}
	// Example first line:
	// ref: refs/heads/main\tHEAD
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ref:") {
			// ref: refs/heads/main\tHEAD
			rest := strings.TrimSpace(strings.TrimPrefix(line, "ref:"))
			fields := strings.Fields(rest)
			if len(fields) == 0 {
				continue
			}
			ref := fields[0]
			const prefix = "refs/heads/"
			if strings.HasPrefix(ref, prefix) {
				return strings.TrimPrefix(ref, prefix), nil
			}
		}
	}
	return "", fmt.Errorf("could not parse default branch from ls-remote for %s", remote)
}

// Clone clones remote into dest at branch.
func (c *CLI) Clone(ctx context.Context, remote, dest, branch string) error {
	if err := ensureDestOK(dest); err != nil {
		return err
	}
	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	args := []string{"clone"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, "--", remote, dest)
	_, err := c.run(ctx, "", args...)
	return err
}

func ensureDestOK(dest string) error {
	fi, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("clone destination exists and is not a directory: %s", dest)
	}
	entries, err := os.ReadDir(dest)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("clone destination is not empty: %s", dest)
	}
	// Empty dir: remove so git clone can create it.
	if err := os.Remove(dest); err != nil {
		return fmt.Errorf("prepare empty clone dest: %w", err)
	}
	return nil
}

// Probe inspects local dirty/untracked and upstream ahead/behind.
func (c *CLI) Probe(ctx context.Context, repoPath string) (ProbeResult, error) {
	var res ProbeResult

	branch, err := c.run(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	if branch == "HEAD" {
		res.Branch = "DETACHED"
	} else {
		res.Branch = branch
	}

	// Porcelain v1 is enough and widely available.
	porcelain, err := c.run(ctx, repoPath, "status", "--porcelain")
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}
	res.Dirty, res.Untracked = parsePorcelain(porcelain)
	res.Clean = !res.Dirty && !res.Untracked

	// Sync comparison ref:
	// 1) @{upstream} when branch tracking is set
	// 2) else origin/<branch> when that remote-tracking ref exists (MVP fallback)
	// 3) else no-upstream
	compareRef, ok := c.syncCompareRef(ctx, repoPath, res.Branch)
	if !ok {
		res.NoUpstream = true
		return res, nil
	}

	counts, err := c.run(ctx, repoPath, "rev-list", "--left-right", "--count", "HEAD..."+compareRef)
	if err != nil {
		res.SyncUnknown = true
		return res, nil
	}
	// "ahead behind" as two integers
	fields := strings.Fields(counts)
	if len(fields) != 2 {
		res.SyncUnknown = true
		return res, nil
	}
	var ahead, behind int
	if _, err := fmt.Sscanf(fields[0], "%d", &ahead); err != nil {
		res.SyncUnknown = true
		return res, nil
	}
	if _, err := fmt.Sscanf(fields[1], "%d", &behind); err != nil {
		res.SyncUnknown = true
		return res, nil
	}
	res.Ahead = ahead
	res.Behind = behind
	return res, nil
}

// syncCompareRef picks the ref for ahead/behind. Prefer @{upstream}; fall back
// to origin/<branch> when tracking is unset but that remote-tracking branch exists.
func (c *CLI) syncCompareRef(ctx context.Context, repoPath, branch string) (string, bool) {
	if _, err := c.run(ctx, repoPath, "rev-parse", "--abbrev-ref", "@{upstream}"); err == nil {
		return "@{upstream}", true
	}
	if branch == "" || branch == "DETACHED" || branch == "HEAD" {
		return "", false
	}
	// MVP: hard-code remote name "origin" (git clone default).
	cand := "origin/" + branch
	if _, err := c.run(ctx, repoPath, "rev-parse", "--verify", cand); err != nil {
		return "", false
	}
	return cand, true
}

func parsePorcelain(out string) (dirty, untracked bool) {
	if out == "" {
		return false, false
	}
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		// XY PATH — untracked is "?? "
		if strings.HasPrefix(line, "??") || strings.HasPrefix(line, "!!") {
			untracked = true
			continue
		}
		if len(line) >= 2 {
			dirty = true
		}
	}
	return dirty, untracked
}
