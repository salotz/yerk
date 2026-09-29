package api_test

import (
	"testing"

	"github.com/salotz/yerk/internal/api"
	"github.com/salotz/yerk/internal/config"
	"github.com/salotz/yerk/internal/presence"
)

func TestProjectFromConfig(t *testing.T) {
	t.Parallel()
	p := api.ProjectFromConfig(config.Project{
		Name:           "yerk",
		Domain:         "personal",
		Remote:         "git@example.com:salotz/yerk.git",
		Path:           "devel/yerk",
		DefaultReplica: "main",
		Tags:           []string{"devel"},
	})
	if p.APIVersion != api.APIVersion || p.Kind != api.KindProject {
		t.Fatalf("type meta: %+v", p)
	}
	if p.Name != "yerk" || p.Domain != "personal" || p.Path != "devel/yerk" {
		t.Fatalf("identity: %+v", p)
	}
	if p.DefaultReplica != "main" || p.Remote == "" {
		t.Fatalf("placement: %+v", p)
	}
	if len(p.Tags) != 1 || p.Tags[0] != "devel" {
		t.Fatalf("tags: %+v", p.Tags)
	}
	// Defensive copy: mutating source tags must not affect resource.
	src := config.Project{Name: "x", Tags: []string{"devel"}}
	out := api.ProjectFromConfig(src)
	src.Tags[0] = "mutated"
	if out.Tags[0] != "devel" {
		t.Fatalf("expected tag copy, got %q", out.Tags[0])
	}
}

func TestCatalogFromConfig(t *testing.T) {
	t.Parallel()
	c := api.CatalogFromConfig(config.Catalog{
		Tags: []string{"devel", "work"},
		Projects: []config.Project{
			{Name: "a", Remote: "r1", Tags: []string{"devel"}},
			{Name: "b", Remote: "r2"},
		},
	})
	if c.Kind != api.KindCatalog || c.APIVersion != api.APIVersion {
		t.Fatalf("type meta: %+v", c)
	}
	if len(c.Tags) != 2 || len(c.Projects) != 2 {
		t.Fatalf("sizes tags=%d projects=%d", len(c.Tags), len(c.Projects))
	}
	if c.Projects[0].Kind != api.KindProject || c.Projects[0].Name != "a" {
		t.Fatalf("project0: %+v", c.Projects[0])
	}
}

func TestPresenceFrom(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   presence.Status
		want api.Presence
	}{
		{presence.Missing, api.PresenceMissing},
		{presence.Present, api.PresencePresent},
		{presence.Invalid, api.PresenceInvalid},
	}
	for _, tc := range cases {
		if got := api.PresenceFrom(tc.in); got != tc.want {
			t.Fatalf("%s: got %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestReplicaStatusSummary(t *testing.T) {
	t.Parallel()
	r := api.NewReplicaStatus()
	r.Project = "yerk"
	r.Replica = "main"
	r.Path = "/tmp/yerk/main"
	r.Presence = api.PresencePresent
	r.Change = "clean"
	r.Branch = "main"
	s := r.Summary()
	if s.Name != "main" || s.Path != r.Path || s.Presence != api.PresencePresent {
		t.Fatalf("summary: %+v", s)
	}
	if s.Change != "clean" || s.Branch != "main" {
		t.Fatalf("summary probe: %+v", s)
	}
}
