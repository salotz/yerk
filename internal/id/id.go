// Package id parses and formats yerk project/replica identifiers and URIs.
//
// Two spellings only (ADR 012):
//
//	bare:  personal/wumpus[/main]
//	URI:   yerk://personal/wumpus[/main]
//
// Short single-segment names expand against a catalog key list when unique.
package id

import (
	"fmt"
	"strings"
)

// Scheme is the canonical URI scheme (with hierarchical // form).
const Scheme = "yerk"

// Ref is a parsed project or replica reference.
type Ref struct {
	Domain  string
	Project string
	Replica string // empty ⇒ project-only
}

// IsReplica reports whether the ref includes a replica distinguisher.
func (r Ref) IsReplica() bool {
	return r.Replica != ""
}

// Bare returns the bare identifier (domain/project[/replica]).
func (r Ref) Bare() string {
	if r.Domain == "" || r.Project == "" {
		return ""
	}
	if r.Replica == "" {
		return r.Domain + "/" + r.Project
	}
	return r.Domain + "/" + r.Project + "/" + r.Replica
}

// URI returns the canonical yerk:// URI for this ref.
func (r Ref) URI() string {
	b := r.Bare()
	if b == "" {
		return ""
	}
	return Scheme + "://" + b
}

// ProjectRef returns a copy without the replica segment.
func (r Ref) ProjectRef() Ref {
	return Ref{Domain: r.Domain, Project: r.Project}
}

// ProjectURI returns yerk://domain/project for domain+name.
func ProjectURI(domain, project string) string {
	return Ref{Domain: domain, Project: project}.URI()
}

// ReplicaURI returns yerk://domain/project/replica.
func ReplicaURI(domain, project, replica string) string {
	return Ref{Domain: domain, Project: project, Replica: replica}.URI()
}

// ProjectKey is the minimum catalog identity needed to expand short names.
type ProjectKey struct {
	Domain string
	Name   string
}

// Parse parses a bare identifier or canonical URI without catalog expansion.
// Single-segment short names are returned with empty Domain (caller must Expand).
// Two- or three-segment forms require an explicit domain.
func Parse(s string) (Ref, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Ref{}, fmt.Errorf("empty identifier")
	}
	if strings.ContainsAny(s, "?#") {
		return Ref{}, fmt.Errorf("identifier %q: query and fragment not allowed", s)
	}

	bare := s
	if strings.Contains(s, "://") {
		scheme, rest, ok := strings.Cut(s, "://")
		if !ok || scheme != Scheme {
			return Ref{}, fmt.Errorf("identifier %q: only %s:// URIs are accepted", s, Scheme)
		}
		if rest == "" {
			return Ref{}, fmt.Errorf("identifier %q: empty path after %s://", s, Scheme)
		}
		if strings.HasPrefix(rest, "/") {
			return Ref{}, fmt.Errorf("identifier %q: unexpected leading slash after %s://", s, Scheme)
		}
		bare = rest
	} else if strings.HasPrefix(s, Scheme+":") && !strings.HasPrefix(s, Scheme+"://") {
		return Ref{}, fmt.Errorf("identifier %q: use %s://… (opaque %s: form is not accepted)", s, Scheme, Scheme)
	}

	if strings.HasPrefix(bare, "/") || strings.HasSuffix(bare, "/") {
		return Ref{}, fmt.Errorf("identifier %q: leading/trailing slash not allowed", s)
	}
	if strings.Contains(bare, "//") {
		return Ref{}, fmt.Errorf("identifier %q: empty path segment", s)
	}

	parts := strings.Split(bare, "/")
	if len(parts) == 0 || len(parts) > 3 {
		return Ref{}, fmt.Errorf("identifier %q: want domain/project[/replica] (at most 3 segments)", s)
	}
	for i, p := range parts {
		if err := validateSegment(p); err != nil {
			return Ref{}, fmt.Errorf("identifier %q: segment %d: %w", s, i+1, err)
		}
	}

	switch len(parts) {
	case 1:
		// Short project name; domain filled by Expand.
		return Ref{Project: parts[0]}, nil
	case 2:
		return Ref{Domain: parts[0], Project: parts[1]}, nil
	default:
		return Ref{Domain: parts[0], Project: parts[1], Replica: parts[2]}, nil
	}
}

// Expand parses input and fills domain for short project names using keys.
//
// Forms:
//   - one segment: unique short project name → domain/name
//   - two segments: domain/project if that pair exists; else unique short
//     name + replica (e.g. wumpus/main when only one wumpus)
//   - three segments or yerk://…: fully qualified; must exist when keys ≠ nil
func Expand(input string, keys []ProjectKey) (Ref, error) {
	ref, err := Parse(input)
	if err != nil {
		return Ref{}, err
	}

	// Fully qualified three-segment, or URI/bare with domain already set from
	// three segments: Domain+Project+Replica from Parse.
	// Two-segment Parse always sets Domain=parts[0], Project=parts[1].
	if ref.Domain != "" && ref.Replica != "" {
		if keys != nil && !hasProject(keys, ref.Domain, ref.Project) {
			return Ref{}, fmt.Errorf("project %q not in catalog", ref.BareProject())
		}
		return ref, nil
	}

	if ref.Domain == "" {
		// One-segment short project name.
		return expandShortProject(ref.Project, "", keys)
	}

	// Two segments from Parse: domain/project. Prefer exact catalog pair;
	// else treat as shortName/replica when the first segment is a unique name.
	if keys != nil && hasProject(keys, ref.Domain, ref.Project) {
		return ref, nil
	}
	shortName, replica := ref.Domain, ref.Project
	expanded, err := expandShortProject(shortName, replica, keys)
	if err == nil {
		return expanded, nil
	}
	// Prefer a clear "not in catalog" when exact pair was intended.
	if keys != nil {
		// If short name is ambiguous or missing, surface that; if short name
		// isn't a project name at all, report missing domain/project.
		matches := findByName(keys, shortName)
		if len(matches) == 0 {
			return Ref{}, fmt.Errorf("project %q not in catalog", shortName+"/"+replica)
		}
		return Ref{}, err
	}
	return ref, nil
}

func expandShortProject(name, replica string, keys []ProjectKey) (Ref, error) {
	matches := findByName(keys, name)
	switch len(matches) {
	case 0:
		return Ref{}, fmt.Errorf("project %q not in catalog", name)
	case 1:
		return Ref{Domain: matches[0].Domain, Project: matches[0].Name, Replica: replica}, nil
	default:
		cands := make([]string, 0, len(matches))
		for _, m := range matches {
			cands = append(cands, m.Domain+"/"+m.Name)
		}
		return Ref{}, fmt.Errorf("ambiguous project name %q; candidates: %s", name, strings.Join(cands, ", "))
	}
}

// BareProject returns domain/project without replica.
func (r Ref) BareProject() string {
	if r.Domain == "" || r.Project == "" {
		return r.Project
	}
	return r.Domain + "/" + r.Project
}

func validateSegment(s string) error {
	if s == "" {
		return fmt.Errorf("empty")
	}
	if strings.TrimSpace(s) != s {
		return fmt.Errorf("%q has leading/trailing space", s)
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '.':
		default:
			return fmt.Errorf("invalid character %q in %q", string(r), s)
		}
	}
	return nil
}

func findByName(keys []ProjectKey, name string) []ProjectKey {
	var out []ProjectKey
	for _, k := range keys {
		if k.Name == name {
			out = append(out, k)
		}
	}
	return out
}

func hasProject(keys []ProjectKey, domain, name string) bool {
	for _, k := range keys {
		if k.Domain == domain && k.Name == name {
			return true
		}
	}
	return false
}
