// Package version holds build-time identity for the yerk binary.
package version

// Values may be overridden at link time via -ldflags, for example:
//
//	-X github.com/salotz/yerk/internal/version.Version=0.1.0
//	-X github.com/salotz/yerk/internal/version.Commit=abc1234
//	-X github.com/salotz/yerk/internal/version.BuildDate=2026-09-25T00:00:00Z
var (
	// Version is the semantic version string.
	Version = "0.0.0-dev"
	// Commit is the short git commit hash, when known.
	Commit = "unknown"
	// BuildDate is the UTC build timestamp, when known.
	BuildDate = "unknown"
)

// String returns a single-line identity suitable for CLI output.
func String() string {
	return Version + " (commit " + Commit + ", built " + BuildDate + ")"
}
