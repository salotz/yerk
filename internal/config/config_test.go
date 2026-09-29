package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/config"
)

func TestLoadMissingFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Workspace.Style != "workspace-dir" {
		t.Fatalf("style: got %q", cfg.Workspace.Style)
	}

	cat, err := config.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Projects) != 0 {
		t.Fatalf("expected empty catalog, got %d", len(cat.Projects))
	}
}

func TestLoadSplitFiles(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	catPath := filepath.Join(dir, "catalog.toml")
	if err := os.WriteFile(cfgPath, []byte(`
[workspace]
style = "project-dir"

[domains]
personal = "/tree/personal"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(catPath, []byte(`
tags = ["devel"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@example.com:salotz/yerk.git"
path = "devel/yerk"
tags = ["devel"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")
	t.Setenv("YERK__WORKSPACE_STYLE", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Workspace.Style != "project-dir" {
		t.Fatalf("config: %+v", cfg.Workspace)
	}
	if cfg.Domains["personal"] != "/tree/personal" {
		t.Fatalf("domains: %+v", cfg.Domains)
	}

	cat, err := config.LoadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Projects) != 1 || cat.Projects[0].Name != "yerk" {
		t.Fatalf("catalog: %+v", cat.Projects)
	}
	p, ok := cat.Find("yerk")
	if !ok || p.Domain != "personal" || p.Path != "devel/yerk" {
		t.Fatalf("find: %+v ok=%v", p, ok)
	}
}

func TestDomainRootExpand(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Domains: map[string]string{
		"personal": "~/tree/personal",
	}}
	root, ok, err := cfg.DomainRoot("personal")
	if err != nil || !ok {
		t.Fatalf("DomainRoot: %v ok=%v", err, ok)
	}
	want := filepath.Join(home, "tree/personal")
	if root != want {
		t.Fatalf("got %q want %q", root, want)
	}
	_, ok, err = cfg.DomainRoot("missing")
	if err != nil || ok {
		t.Fatalf("missing domain: ok=%v err=%v", ok, err)
	}
}

func TestEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(cfgPath, []byte(`
[workspace]
style = "workspace-dir"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__WORKSPACE_STYLE", "project-dir")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Workspace.Style != "project-dir" {
		t.Fatalf("style env overlay: %+v", cfg.Workspace)
	}
}

func TestFilterTag(t *testing.T) {
	t.Parallel()
	cat := config.Catalog{
		Tags: []string{"devel", "work"},
		Projects: []config.Project{
			{Name: "a", Tags: []string{"devel"}},
			{Name: "b", Tags: []string{"work"}},
			{Name: "c", Tags: []string{"devel", "work"}},
		},
	}
	got := cat.FilterTag("devel")
	if len(got) != 2 {
		t.Fatalf("got %d", len(got))
	}

	sel, err := cat.SelectByTag("devel")
	if err != nil {
		t.Fatal(err)
	}
	if len(sel) != 2 {
		t.Fatalf("SelectByTag got %d", len(sel))
	}
	_, err = cat.SelectByTag("nope")
	if err == nil || !strings.Contains(err.Error(), "unknown tag") {
		t.Fatalf("SelectByTag unknown: %v", err)
	}
	_, err = cat.SelectByTag("")
	if err == nil {
		t.Fatal("empty tag should error")
	}
	lonely := config.Catalog{
		Tags:     []string{"lonely"},
		Projects: []config.Project{{Name: "a"}},
	}
	got, err = lonely.SelectByTag("lonely")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty match, got %+v", got)
	}
}

func TestCatalogValidateTags(t *testing.T) {
	t.Parallel()

	ok := config.Catalog{
		Tags: []string{"devel", "work"},
		Projects: []config.Project{
			{Name: "a", Tags: []string{"devel"}},
			{Name: "b"},
		},
	}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid catalog: %v", err)
	}

	emptyVocabOK := config.Catalog{
		Tags:     nil,
		Projects: []config.Project{{Name: "a"}},
	}
	if err := emptyVocabOK.Validate(); err != nil {
		t.Fatalf("empty tags + untagged projects: %v", err)
	}

	cases := []struct {
		name string
		cat  config.Catalog
		want string
	}{
		{
			name: "undeclared project tag",
			cat: config.Catalog{
				Tags:     []string{"devel"},
				Projects: []config.Project{{Name: "a", Tags: []string{"work"}}},
			},
			want: "not declared",
		},
		{
			name: "empty vocab with project tag",
			cat: config.Catalog{
				Projects: []config.Project{{Name: "a", Tags: []string{"devel"}}},
			},
			want: "not declared",
		},
		{
			name: "duplicate catalog tag",
			cat:  config.Catalog{Tags: []string{"devel", "devel"}},
			want: "duplicate",
		},
		{
			name: "duplicate project tag",
			cat: config.Catalog{
				Tags:     []string{"devel"},
				Projects: []config.Project{{Name: "a", Tags: []string{"devel", "devel"}}},
			},
			want: "duplicate",
		},
		{
			name: "blank catalog tag",
			cat:  config.Catalog{Tags: []string{"  "}},
			want: "empty tag",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cat.Validate()
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v want substring %q", err, tc.want)
			}
		})
	}
}

func TestLoadCatalogRejectsUndeclaredTag(t *testing.T) {
	dir := t.TempDir()
	catPath := filepath.Join(dir, "catalog.toml")
	if err := os.WriteFile(catPath, []byte(`
[[projects]]
name = "yerk"
remote = "git@example.com:salotz/yerk.git"
path = "/tmp/yerk"
tags = ["devel"]
`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CATALOG", "")
	_, err := config.LoadCatalog()
	if err == nil {
		t.Fatal("expected load error for undeclared tag")
	}
	if !strings.Contains(err.Error(), "not declared") {
		t.Fatalf("got %v", err)
	}
}
