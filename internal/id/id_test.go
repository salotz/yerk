package id_test

import (
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/id"
)

func TestParseURIAndBare(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in      string
		domain  string
		project string
		replica string
	}{
		{"personal/wumpus", "personal", "wumpus", ""},
		{"personal/wumpus/main", "personal", "wumpus", "main"},
		{"yerk://personal/wumpus", "personal", "wumpus", ""},
		{"yerk://personal/wumpus/main", "personal", "wumpus", "main"},
		{"wumpus", "", "wumpus", ""},
		{"My-Proj_1.0", "", "My-Proj_1.0", ""},
	}
	for _, tc := range cases {
		ref, err := id.Parse(tc.in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.in, err)
		}
		if ref.Domain != tc.domain || ref.Project != tc.project || ref.Replica != tc.replica {
			t.Fatalf("Parse(%q)=%+v want domain=%q project=%q replica=%q", tc.in, ref, tc.domain, tc.project, tc.replica)
		}
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()
	bads := []string{
		"",
		"yerk:personal/wumpus",
		"http://personal/wumpus",
		"yerk://personal/wumpus?x=1",
		"yerk://personal/wumpus#frag",
		"/personal/wumpus",
		"personal/wumpus/",
		"personal//wumpus",
		"a/b/c/d",
		"has space/x",
		"yerk://",
		"yerk:///personal/wumpus",
	}
	for _, s := range bads {
		if _, err := id.Parse(s); err == nil {
			t.Fatalf("Parse(%q): expected error", s)
		}
	}
}

func TestRefURIBare(t *testing.T) {
	t.Parallel()
	r := id.Ref{Domain: "personal", Project: "wumpus", Replica: "main"}
	if r.Bare() != "personal/wumpus/main" {
		t.Fatalf("Bare: %q", r.Bare())
	}
	if r.URI() != "yerk://personal/wumpus/main" {
		t.Fatalf("URI: %q", r.URI())
	}
	if id.ProjectURI("personal", "wumpus") != "yerk://personal/wumpus" {
		t.Fatal("ProjectURI")
	}
	if id.ReplicaURI("personal", "wumpus", "main") != "yerk://personal/wumpus/main" {
		t.Fatal("ReplicaURI")
	}
	if !r.IsReplica() || r.ProjectRef().IsReplica() {
		t.Fatal("IsReplica / ProjectRef")
	}
}

func TestExpandUniqueAndAmbiguous(t *testing.T) {
	t.Parallel()
	keys := []id.ProjectKey{
		{Domain: "personal", Name: "wumpus"},
		{Domain: "work", Name: "wumpus"},
		{Domain: "personal", Name: "yerk"},
	}

	ref, err := id.Expand("yerk", keys)
	if err != nil {
		t.Fatal(err)
	}
	if ref.URI() != "yerk://personal/yerk" {
		t.Fatalf("got %+v", ref)
	}

	ref, err = id.Expand("personal/wumpus/main", keys)
	if err != nil {
		t.Fatal(err)
	}
	if ref.URI() != "yerk://personal/wumpus/main" || !ref.IsReplica() {
		t.Fatalf("got %+v", ref)
	}

	_, err = id.Expand("wumpus", keys)
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("want ambiguous, got %v", err)
	}
	if !strings.Contains(err.Error(), "personal/wumpus") || !strings.Contains(err.Error(), "work/wumpus") {
		t.Fatalf("want candidates listed: %v", err)
	}

	_, err = id.Expand("missing", keys)
	if err == nil || !strings.Contains(err.Error(), "not in catalog") {
		t.Fatalf("want missing, got %v", err)
	}

	_, err = id.Expand("other/yerk", keys)
	if err == nil || !strings.Contains(err.Error(), "not in catalog") {
		t.Fatalf("want missing qualified, got %v", err)
	}

	// Unique short name + replica.
	ref, err = id.Expand("yerk/main", keys)
	if err != nil {
		t.Fatal(err)
	}
	if ref.URI() != "yerk://personal/yerk/main" || ref.Replica != "main" {
		t.Fatalf("short+replica: %+v", ref)
	}

	// Ambiguous short + replica still errors.
	_, err = id.Expand("wumpus/main", keys)
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("want ambiguous short+replica, got %v", err)
	}
}

func TestExpandURI(t *testing.T) {
	t.Parallel()
	keys := []id.ProjectKey{{Domain: "personal", Name: "yerk"}}
	ref, err := id.Expand("yerk://personal/yerk", keys)
	if err != nil {
		t.Fatal(err)
	}
	if ref.Bare() != "personal/yerk" {
		t.Fatalf("%+v", ref)
	}
}
