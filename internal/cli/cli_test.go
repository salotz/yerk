package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/cli"
)

func TestStatusEmptyCatalog(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")

	var out, errBuf bytes.Buffer
	streams := cli.IO{Out: &out, Err: &errBuf}
	if err := cli.Execute(context.Background(), streams, []string{"status"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No projects") {
		t.Fatalf("out=%q", out.String())
	}
}

func TestStatusAndPathWithFixtures(t *testing.T) {
	dir := t.TempDir()
	// Domain root + relative catalog path → project workspace; replica under style.
	domainRoot := filepath.Join(dir, "personal")
	yerkWS := filepath.Join(domainRoot, "devel", "yerk")
	yerkReplica := filepath.Join(yerkWS, "main")
	missingReplica := filepath.Join(domainRoot, "devel", "missing-one", "main")
	if err := os.MkdirAll(filepath.Join(yerkReplica, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := []byte(`[workspace]
style = "workspace-dir"

[domains]
personal = "` + domainRoot + `"
`)
	cat := []byte(`
tags = ["devel"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@example.com:salotz/yerk.git"
default_replica = "main"
path = "devel/yerk"
tags = ["devel"]

[[projects]]
name = "missing-one"
domain = "personal"
remote = "git@example.com:x/y.git"
default_replica = "main"
path = "devel/missing-one"
`)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")
	t.Setenv("YERK__WORKSPACE_STYLE", "")

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	if err := cli.Execute(context.Background(), streams, []string{"status"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "yerk") || !strings.Contains(s, "present") {
		t.Fatalf("status: %s", s)
	}
	if !strings.Contains(s, "missing-one") || !strings.Contains(s, "missing") {
		t.Fatalf("status missing: %s", s)
	}
	if !strings.Contains(s, yerkReplica) {
		t.Fatalf("status should show replica path %s\n%s", yerkReplica, s)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"path", "yerk"}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != yerkWS {
		t.Fatalf("path project got %q want workspace %q", out.String(), yerkWS)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"path", "yerk", "main"}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != yerkReplica {
		t.Fatalf("path replica got %q want %q", out.String(), yerkReplica)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"workspace", "ensure"}); err == nil {
		t.Fatal("workspace ensure with no args should require --all or project names")
	}

	out.Reset()
	missingWS := filepath.Join(domainRoot, "devel", "missing-one")
	if err := cli.Execute(context.Background(), streams, []string{"workspace", "ensure", "missing-one"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(missingWS); err != nil {
		t.Fatalf("workspace ensure should create project workspace: %v", err)
	}
	if _, err := os.Stat(missingReplica); !os.IsNotExist(err) {
		t.Fatalf("workspace ensure must not create replica leaf %s (err=%v)", missingReplica, err)
	}
	if !strings.Contains(out.String(), missingWS) {
		t.Fatalf("ensure output should show workspace path\n%s", out.String())
	}

	out.Reset()
	// Fresh missing project for --all; yerk workspace already exists from fixtures.
	if err := cli.Execute(context.Background(), streams, []string{"workspace", "ensure", "--all"}); err != nil {
		t.Fatal(err)
	}
	sAll := out.String()
	if !strings.Contains(sAll, yerkWS) || !strings.Contains(sAll, missingWS) {
		t.Fatalf("ensure --all should list each workspace\n%s", sAll)
	}
	if err := cli.Execute(context.Background(), streams, []string{"workspace", "ensure", "--all", "yerk"}); err == nil {
		t.Fatal("ensure --all with project names should error")
	}
}

func TestStatusUnknownTag(t *testing.T) {
	dir := t.TempDir()
	cat := []byte(`
tags = ["devel"]

[[projects]]
name = "yerk"
remote = "git@example.com:salotz/yerk.git"
path = "/tmp/yerk"
tags = ["devel"]
`)
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	err := cli.Execute(context.Background(), streams, []string{"status", "--tag", "nope"})
	if err == nil {
		t.Fatal("expected error for undeclared --tag")
	}
	if !strings.Contains(err.Error(), "unknown tag") {
		t.Fatalf("got %v", err)
	}
}

func TestStatusTagFilter(t *testing.T) {
	dir := t.TempDir()
	domainRoot := filepath.Join(dir, "personal")
	yerkWS := filepath.Join(domainRoot, "devel", "yerk")
	yerkReplica := filepath.Join(yerkWS, "main")
	if err := os.MkdirAll(filepath.Join(yerkReplica, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := []byte(`[workspace]
style = "workspace-dir"

[domains]
personal = "` + domainRoot + `"
`)
	cat := []byte(`
tags = ["devel", "work"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@example.com:salotz/yerk.git"
default_replica = "main"
path = "devel/yerk"
tags = ["devel"]

[[projects]]
name = "office"
domain = "personal"
remote = "git@example.com:x/office.git"
default_replica = "main"
path = "devel/office"
tags = ["work"]
`)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")
	t.Setenv("YERK__WORKSPACE_STYLE", "")

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	if err := cli.Execute(context.Background(), streams, []string{"status", "--tag", "devel"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "filter.tag=devel") {
		t.Fatalf("expected filter.tag line\n%s", s)
	}
	if !strings.Contains(s, "yerk") || !strings.Contains(s, "present") {
		t.Fatalf("expected yerk row\n%s", s)
	}
	if strings.Contains(s, "office") {
		t.Fatalf("office should be filtered out\n%s", s)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"status", "--tag", "work"}); err != nil {
		t.Fatal(err)
	}
	s = out.String()
	if !strings.Contains(s, "office") || !strings.Contains(s, "filter.tag=work") {
		t.Fatalf("expected office under work\n%s", s)
	}
	if strings.Contains(s, "yerk") {
		t.Fatalf("yerk should be filtered out for work\n%s", s)
	}
}

func TestRootHelpListsPrimaryEnvAndPointer(t *testing.T) {
	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	if err := cli.Execute(context.Background(), streams, []string{"--help"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, name := range []string{
		"YERK__CONFIG_DIR", "YERK__CONFIG", "YERK__CATALOG",
		"YERK__WORKSPACE_STYLE",
		"help envvars",
	} {
		if !strings.Contains(s, name) {
			t.Fatalf("root --help missing %s\n%s", name, s)
		}
	}
	if strings.Contains(s, "PRJX_CACHE_HOME") {
		t.Fatalf("root --help should not dump full PRJX list\n%s", s)
	}
}

func TestEnvvarsCommandLiveValues(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")
	t.Setenv("YERK__WORKSPACE_STYLE", "")

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	if err := cli.Execute(context.Background(), streams, []string{"envvars"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "NAME") || !strings.Contains(s, "VALUE") {
		t.Fatalf("expected live table headers\n%s", s)
	}
	if !strings.Contains(s, "YERK__CONFIG_DIR") || !strings.Contains(s, dir) {
		t.Fatalf("expected live YERK__CONFIG_DIR value\n%s", s)
	}
	// Live dump is not the documentation page
	if strings.Contains(s, "Command-scoped") {
		t.Fatalf("yerk envvars should print values, not full docs\n%s", s)
	}
}

func TestHelpEnvvarsTopic(t *testing.T) {
	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	if err := cli.Execute(context.Background(), streams, []string{"help", "envvars"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "YERK__CONFIG") || !strings.Contains(s, "policy=warn") {
		t.Fatalf("yerk help envvars incomplete\n%s", s)
	}
	if !strings.Contains(s, "type=string") || !strings.Contains(s, "default:") {
		t.Fatalf("yerk help envvars should show type/default\n%s", s)
	}
	if !strings.Contains(s, "Command-scoped") || !strings.Contains(s, "Global") {
		t.Fatalf("yerk help envvars should be documentation\n%s", s)
	}
	if !strings.Contains(s, "PRJX (RFC 28)") {
		t.Fatalf("yerk help envvars should point at PRJX without listing every name\n%s", s)
	}
	if strings.Contains(s, "PRJX_ROOT") {
		t.Fatalf("yerk help envvars should not dump PRJX_* names\n%s", s)
	}
}

func TestStatusHelpListsPrimaryToolEnv(t *testing.T) {
	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	if err := cli.Execute(context.Background(), streams, []string{"status", "--help"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, name := range []string{"YERK__CONFIG", "YERK__CATALOG", "YERK__WORKSPACE_STYLE", "PATH", "help envvars"} {
		if !strings.Contains(s, name) {
			t.Fatalf("status --help missing %s\n%s", name, s)
		}
	}
}

func TestConfigAndCatalogHelpers(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	if err := cli.Execute(context.Background(), streams, []string{"config", "path"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "config.toml") {
		t.Fatalf("%q", out.String())
	}
	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"config", "show"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if strings.Contains(s, "workspace.root") {
		t.Fatalf("config show should not mention workspace.root\n%s", s)
	}
	if !strings.Contains(s, "workspace.style") {
		t.Fatalf("config show missing style\n%s", s)
	}
	if !strings.Contains(s, "domains:") {
		t.Fatalf("config show missing domains\n%s", s)
	}

	// Seed a small catalog and assert table show (no sample project tags).
	catBody := `
tags = []

[[projects]]
name = "alpha"
domain = "personal"
remote = "git@example.com:a/alpha.git"
path = "devel/alpha"
default_replica = "main"
`
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), []byte(catBody), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"catalog", "show"}); err != nil {
		t.Fatal(err)
	}
	cs := out.String()
	for _, want := range []string{"NAME", "DOMAIN", "PATH", "REMOTE", "alpha", "personal", "devel/alpha", "main", "tags:"} {
		if !strings.Contains(cs, want) {
			t.Fatalf("catalog show missing %q\n%s", want, cs)
		}
	}
	if strings.Contains(cs, "  - alpha") {
		t.Fatalf("catalog show should be tabular, not a bullet list\n%s", cs)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"catalog", "--help"}); err != nil {
		t.Fatal(err)
	}
	help := out.String()
	if strings.Contains(help, "  example ") {
		t.Fatalf("catalog help must not list example subcommand\n%s", help)
	}
	if !strings.Contains(help, "path") || !strings.Contains(help, "show") {
		t.Fatalf("catalog help should list path and show\n%s", help)
	}
	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"catalog", "example"}); err == nil {
		t.Fatal("catalog example should error (unknown command)")
	}
	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"config", "example"}); err == nil {
		t.Fatal("config example should error (unknown command)")
	}
}

