// Package presence classifies whether an expected replica path exists and
// looks like a usable git checkout.
package presence

import (
	"os"
	"path/filepath"
)

// Status is the on-disk presence of a replica path.
type Status string

const (
	// Missing: path does not exist.
	Missing Status = "missing"
	// Present: path exists and is a usable git checkout.
	Present Status = "present"
	// Invalid: path exists but is not a usable git checkout.
	Invalid Status = "invalid"
)

// Classify returns presence for path.
func Classify(path string) Status {
	if path == "" {
		return Missing
	}
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Missing
		}
		// Unreadable etc. — treat as invalid so the operator sees a row.
		return Invalid
	}
	if !fi.IsDir() {
		return Invalid
	}
	if isGitCheckout(path) {
		return Present
	}
	return Invalid
}

func isGitCheckout(dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	fi, err := os.Lstat(gitPath)
	if err != nil {
		return false
	}
	// Regular repo: .git directory. Worktree / linked: .git file.
	if fi.IsDir() || fi.Mode().IsRegular() {
		return true
	}
	// Symlink to gitdir also OK.
	if fi.Mode()&os.ModeSymlink != 0 {
		return true
	}
	return false
}
