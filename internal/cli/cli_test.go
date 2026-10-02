package cli_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salotz/yerk/internal/cli"
)

// setYerkHostEnv points config + state at an isolated temp tree for a test.
func setYerkHostEnv(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("YERK__CONFIG_DIR", dir)
	t.Setenv("YERK__CONFIG", "")
	t.Setenv("YERK__CATALOG", "")
	t.Setenv("YERK__STATE_DIR", filepath.Join(dir, ".state"))
	t.Setenv("YERK__WORKSPACE_STYLE", "")
}

// writeHostConfig writes config.toml + catalog.toml for isolated host fixtures.
func writeHostConfig(t *testing.T, dir, cfgBody, catBody string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(cfgBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), []byte(catBody), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestStatusEmptyCatalog(t *testing.T) {
	dir := t.TempDir()
	setYerkHostEnv(t, dir)

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
	yerkWS := filepath.Join(dir, "devel", "yerk")
	yerkReplica := filepath.Join(yerkWS, "main")
	missingWS := filepath.Join(dir, "devel", "missing-one")
	missingReplica := filepath.Join(missingWS, "main")
	if err := os.MkdirAll(filepath.Join(yerkReplica, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := []byte(`[workspace]
style = "workspace-dir"

[[projects]]
name = "yerk"
domain = "personal"
path = "` + yerkWS + `"

[[projects]]
name = "missing-one"
domain = "personal"
path = "` + missingWS + `"
`)
	cat := []byte(`
tags = ["devel"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@example.com:salotz/yerk.git"
default_replica = "main"
tags = ["devel"]

[[projects]]
name = "missing-one"
domain = "personal"
remote = "git@example.com:x/y.git"
default_replica = "main"
`)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	setYerkHostEnv(t, dir)

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	// Default status is project-scoped; --presence-only avoids git probe noise
	// when fixtures are not full git repos (only .git dir marker).
	if err := cli.Execute(context.Background(), streams, []string{"status", "--presence-only"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "yerk") || !strings.Contains(s, "present") {
		t.Fatalf("status: %s", s)
	}
	if !strings.Contains(s, "missing-one") || !strings.Contains(s, "missing") {
		t.Fatalf("status missing: %s", s)
	}
	if !strings.Contains(s, yerkWS) {
		t.Fatalf("status should show project workspace path %s\n%s", yerkWS, s)
	}
	// Project view: no REPLICA column; PATH is workspace.
	if strings.Contains(s, "REPLICA") {
		t.Fatalf("project status must not include REPLICA column\n%s", s)
	}
	if !strings.Contains(s, "NAME") || !strings.Contains(s, "PRESENCE") || !strings.Contains(s, "PATH") {
		t.Fatalf("expected project table headers\n%s", s)
	}
	if strings.Contains(s, yerkReplica) && !strings.Contains(s, yerkWS) {
		t.Fatalf("unexpected: replica path without workspace\n%s", s)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"status", "yerk", "--presence-only"}); err != nil {
		t.Fatal(err)
	}
	sOne := out.String()
	if !strings.Contains(sOne, "yerk") || !strings.Contains(sOne, yerkWS) {
		t.Fatalf("status one project: %s", sOne)
	}
	if strings.Contains(sOne, "missing-one") {
		t.Fatalf("single project should not list others\n%s", sOne)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"status", "yerk", "main", "--presence-only"}); err != nil {
		t.Fatal(err)
	}
	sRep := out.String()
	if !strings.Contains(sRep, "yerk") || !strings.Contains(sRep, "main") {
		t.Fatalf("status replica: %s", sRep)
	}
	if !strings.Contains(sRep, yerkReplica) {
		t.Fatalf("replica status should show checkout path %s\n%s", yerkReplica, sRep)
	}
	if !strings.Contains(sRep, "REPLICA") {
		t.Fatalf("replica status should include REPLICA column\n%s", sRep)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"status", "nope"}); err == nil {
		t.Fatal("status unknown project should error")
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

func TestPathAmbiguousShortName(t *testing.T) {
	dir := t.TempDir()
	wsPersonal := filepath.Join(dir, "personal", "devel", "wumpus")
	wsWork := filepath.Join(dir, "work", "devel", "wumpus")
	cfg := `[workspace]
style = "workspace-dir"

[[projects]]
name = "wumpus"
domain = "personal"
path = "` + wsPersonal + `"

[[projects]]
name = "wumpus"
domain = "work"
path = "` + wsWork + `"
`
	cat := `
[[projects]]
name = "wumpus"
domain = "personal"
remote = "git@example.com:a/w.git"

[[projects]]
name = "wumpus"
domain = "work"
remote = "git@example.com:b/w.git"
`
	writeHostConfig(t, dir, cfg, cat)
	setYerkHostEnv(t, dir)

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	if err := cli.Execute(context.Background(), streams, []string{"path", "wumpus"}); err == nil {
		t.Fatal("expected ambiguous short name error")
	} else if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("got %v", err)
	}
	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"path", "work/wumpus"}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != wsWork {
		t.Fatalf("got %q want %q", out.String(), wsWork)
	}
}

func TestStatusUnknownTag(t *testing.T) {
	dir := t.TempDir()
	cat := []byte(`
tags = ["devel"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@example.com:salotz/yerk.git"
tags = ["devel"]
`)
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	setYerkHostEnv(t, dir)

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
	officeWS := filepath.Join(domainRoot, "devel", "office")
	if err := os.MkdirAll(filepath.Join(yerkReplica, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := []byte(`[workspace]
style = "workspace-dir"

[[projects]]
name = "yerk"
domain = "personal"
path = "` + yerkWS + `"

[[projects]]
name = "office"
domain = "personal"
path = "` + officeWS + `"
`)
	cat := []byte(`
tags = ["devel", "work"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@example.com:salotz/yerk.git"
default_replica = "main"
tags = ["devel"]

[[projects]]
name = "office"
domain = "personal"
remote = "git@example.com:x/office.git"
default_replica = "main"
tags = ["work"]
`)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	setYerkHostEnv(t, dir)

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	if err := cli.Execute(context.Background(), streams, []string{"status", "--tag", "devel", "--presence-only"}); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "yerk") || strings.Contains(s, "office") {
		t.Fatalf("tag devel filter: %s", s)
	}
	if !strings.Contains(s, yerkWS) {
		t.Fatalf("expected workspace path in project view\n%s", s)
	}
	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"status", "--tag", "nope"}); err == nil {
		t.Fatal("unknown tag should error")
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
		"YERK__CONFIG_DIR", "YERK__CONFIG", "YERK__CATALOG", "YERK__STATE_DIR",
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
	setYerkHostEnv(t, dir)

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

func TestBareParentCommandsShowHelp(t *testing.T) {
	for _, args := range [][]string{
		{"config"},
		{"catalog"},
		{"workspace"},
	} {
		t.Run(args[0], func(t *testing.T) {
			var out bytes.Buffer
			streams := cli.IO{Out: &out, Err: &out}
			if err := cli.Execute(context.Background(), streams, args); err != nil {
				t.Fatalf("%v args=%v out=%s", err, args, out.String())
			}
			s := out.String()
			if !strings.Contains(s, "Usage:") {
				t.Fatalf("expected help Usage for %v\n%s", args, s)
			}
			// Should not be the old hard error.
			if strings.Contains(s, "subcommand required") {
				t.Fatalf("unexpected error text for %v\n%s", args, s)
			}
		})
	}
}

func TestConfigAndCatalogHelpers(t *testing.T) {
	dir := t.TempDir()
	setYerkHostEnv(t, dir)

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
	if !strings.Contains(s, "projects:") {
		t.Fatalf("config show missing projects\n%s", s)
	}

	// Seed a small catalog and assert table show (no sample project tags).
	catBody := `
tags = []

[[projects]]
name = "alpha"
domain = "personal"
remote = "git@example.com:a/alpha.git"
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
	for _, want := range []string{"NAME", "DOMAIN", "REMOTE", "alpha", "personal", "main", "tags:"} {
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

func TestMaterializeSelectionErrors(t *testing.T) {
	dir := t.TempDir()
	cat := []byte(`
tags = ["devel", "work"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@example.com:salotz/yerk.git"
tags = ["devel"]
`)
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	setYerkHostEnv(t, dir)

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"bare", []string{"materialize"}, "name a project"},
		{"all+name", []string{"materialize", "--all", "yerk"}, "not a combination"},
		{"tag+name", []string{"materialize", "--tag", "devel", "yerk"}, "not a combination"},
		{"all+tag", []string{"materialize", "--all", "--tag", "devel"}, "not a combination"},
		{"unknown-tag", []string{"materialize", "--tag", "nope"}, "unknown tag"},
		{"empty-tag", []string{"materialize", "--tag", "work"}, "no projects matched tag"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := cli.Execute(context.Background(), streams, tc.args)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v want substring %q", err, tc.want)
			}
		})
	}
}

func TestMaterializeAllEmptyCatalog(t *testing.T) {
	dir := t.TempDir()
	setYerkHostEnv(t, dir)
	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	err := cli.Execute(context.Background(), streams, []string{"materialize", "--all"})
	if err == nil {
		t.Fatal("expected empty catalog error")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("got %v", err)
	}
}

func TestMaterializeBulkTagAndAll(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	domainRoot := filepath.Join(dir, "personal")
	// Two bare repos as remotes.
	remoteA := filepath.Join(dir, "remotes", "a.git")
	remoteB := filepath.Join(dir, "remotes", "b.git")
	initFileRemote(t, remoteA)
	initFileRemote(t, remoteB)

	alphaWS := filepath.Join(domainRoot, "devel", "alpha")
	betaWS := filepath.Join(domainRoot, "devel", "beta")
	cfg := []byte(`[workspace]
style = "workspace-dir"

[[projects]]
name = "alpha"
domain = "personal"
path = "` + alphaWS + `"

[[projects]]
name = "beta"
domain = "personal"
path = "` + betaWS + `"
`)
	cat := []byte(`
tags = ["devel", "work"]

[[projects]]
name = "alpha"
domain = "personal"
remote = "` + remoteA + `"
default_replica = "main"
tags = ["devel"]

[[projects]]
name = "beta"
domain = "personal"
remote = "` + remoteB + `"
default_replica = "main"
tags = ["work"]
`)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	setYerkHostEnv(t, dir)

	var out, errBuf bytes.Buffer
	streams := cli.IO{Out: &out, Err: &errBuf}

	if err := cli.Execute(context.Background(), streams, []string{"materialize", "--tag", "devel"}); err != nil {
		t.Fatalf("materialize --tag: %v\nerr=%s\nout=%s", err, errBuf.String(), out.String())
	}
	alphaPath := filepath.Join(domainRoot, "devel", "alpha", "main")
	if strings.TrimSpace(out.String()) != alphaPath {
		t.Fatalf("materialize --tag out=%q want %q", out.String(), alphaPath)
	}
	if _, err := os.Stat(filepath.Join(alphaPath, ".git")); err != nil {
		t.Fatalf("alpha checkout missing: %v", err)
	}
	betaPath := filepath.Join(domainRoot, "devel", "beta", "main")
	if _, err := os.Stat(betaPath); !os.IsNotExist(err) {
		t.Fatalf("beta should not be materialized by --tag devel")
	}

	out.Reset()
	errBuf.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"materialize", "--tag", "work"}); err != nil {
		t.Fatalf("materialize --tag work: %v\n%s", err, errBuf.String())
	}
	if !strings.Contains(out.String(), betaPath) {
		t.Fatalf("expected beta path\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(betaPath, ".git")); err != nil {
		t.Fatalf("beta checkout missing: %v", err)
	}

	// --all on already-materialized catalog should succeed (already present).
	out.Reset()
	errBuf.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"materialize", "--all"}); err != nil {
		t.Fatalf("materialize --all when present should be ok: %v\n%s", err, errBuf.String())
	}
	sAll := out.String()
	if !strings.Contains(sAll, alphaPath) || !strings.Contains(sAll, betaPath) {
		t.Fatalf("expected both paths on already-present --all\n%s", sAll)
	}
	if !strings.Contains(errBuf.String(), "already present") {
		t.Fatalf("expected already present messages on stderr\n%s", errBuf.String())
	}
}

func TestMaterializeAlreadyPresentSingle(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	domainRoot := filepath.Join(dir, "personal")
	remote := filepath.Join(dir, "remotes", "a.git")
	initFileRemote(t, remote)

	alphaWS := filepath.Join(domainRoot, "devel", "alpha")
	betaWS := filepath.Join(domainRoot, "devel", "beta")
	cfg := []byte(`[workspace]
style = "workspace-dir"

[[projects]]
name = "alpha"
domain = "personal"
path = "` + alphaWS + `"

[[projects]]
name = "beta"
domain = "personal"
path = "` + betaWS + `"
`)
	cat := []byte(`
tags = []

[[projects]]
name = "alpha"
domain = "personal"
remote = "` + remote + `"
default_replica = "main"
`)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	setYerkHostEnv(t, dir)

	var out, errBuf bytes.Buffer
	streams := cli.IO{Out: &out, Err: &errBuf}
	if err := cli.Execute(context.Background(), streams, []string{"materialize", "alpha"}); err != nil {
		t.Fatalf("first materialize: %v\n%s", err, errBuf.String())
	}
	path := filepath.Join(domainRoot, "devel", "alpha", "main")
	out.Reset()
	errBuf.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"materialize", "alpha"}); err != nil {
		t.Fatalf("second materialize should succeed: %v\n%s", err, errBuf.String())
	}
	if strings.TrimSpace(out.String()) != path {
		t.Fatalf("out=%q want %q", out.String(), path)
	}
	if !strings.Contains(errBuf.String(), "already present") {
		t.Fatalf("stderr=%q", errBuf.String())
	}
	if strings.Contains(errBuf.String(), "cloning ") {
		t.Fatalf("should not re-materialize\n%s", errBuf.String())
	}
}

func TestMaterializeInvalidPathErrors(t *testing.T) {
	dir := t.TempDir()
	domainRoot := filepath.Join(dir, "personal")
	// Path exists as a plain directory (no .git) → invalid presence.
	invalid := filepath.Join(domainRoot, "devel", "alpha", "main")
	if err := os.MkdirAll(invalid, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(invalid, "not-git"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	alphaWS := filepath.Join(domainRoot, "devel", "alpha")
	betaWS := filepath.Join(domainRoot, "devel", "beta")
	cfg := []byte(`[workspace]
style = "workspace-dir"

[[projects]]
name = "alpha"
domain = "personal"
path = "` + alphaWS + `"

[[projects]]
name = "beta"
domain = "personal"
path = "` + betaWS + `"
`)
	cat := []byte(`
tags = []

[[projects]]
name = "alpha"
domain = "personal"
remote = "git@example.com:x/alpha.git"
default_replica = "main"
`)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	setYerkHostEnv(t, dir)

	var out bytes.Buffer
	streams := cli.IO{Out: &out, Err: &out}
	err := cli.Execute(context.Background(), streams, []string{"materialize", "alpha"})
	if err == nil {
		t.Fatal("expected error for invalid non-checkout path")
	}
	if !strings.Contains(err.Error(), "not a usable git checkout") {
		t.Fatalf("got %v", err)
	}
}

func TestMaterializeAllFresh(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	domainRoot := filepath.Join(dir, "personal")
	remoteA := filepath.Join(dir, "remotes", "a.git")
	remoteB := filepath.Join(dir, "remotes", "b.git")
	initFileRemote(t, remoteA)
	initFileRemote(t, remoteB)

	alphaWS := filepath.Join(domainRoot, "devel", "alpha")
	betaWS := filepath.Join(domainRoot, "devel", "beta")
	cfg := []byte(`[workspace]
style = "workspace-dir"

[[projects]]
name = "alpha"
domain = "personal"
path = "` + alphaWS + `"

[[projects]]
name = "beta"
domain = "personal"
path = "` + betaWS + `"
`)
	cat := []byte(`
tags = []

[[projects]]
name = "alpha"
domain = "personal"
remote = "` + remoteA + `"
default_replica = "main"

[[projects]]
name = "beta"
domain = "personal"
remote = "` + remoteB + `"
default_replica = "main"
`)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "catalog.toml"), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	setYerkHostEnv(t, dir)

	var out, errBuf bytes.Buffer
	streams := cli.IO{Out: &out, Err: &errBuf}
	if err := cli.Execute(context.Background(), streams, []string{"materialize", "--all"}); err != nil {
		t.Fatalf("materialize --all: %v\n%s", err, errBuf.String())
	}
	s := out.String()
	alphaPath := filepath.Join(domainRoot, "devel", "alpha", "main")
	betaPath := filepath.Join(domainRoot, "devel", "beta", "main")
	if !strings.Contains(s, alphaPath) || !strings.Contains(s, betaPath) {
		t.Fatalf("materialize --all paths:\n%s", s)
	}
	for _, p := range []string{alphaPath, betaPath} {
		if _, err := os.Stat(filepath.Join(p, ".git")); err != nil {
			t.Fatalf("%s missing .git: %v", p, err)
		}
	}
}

func TestGetAndLookup(t *testing.T) {
	dir := t.TempDir()
	yerkWS := filepath.Join(dir, "devel", "yerk")
	yerkReplica := filepath.Join(yerkWS, "main")
	deep := filepath.Join(yerkReplica, "pkg", "x")
	if err := os.MkdirAll(filepath.Join(yerkReplica, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `[workspace]
style = "workspace-dir"

[[projects]]
name = "yerk"
domain = "personal"
path = "` + yerkWS + `"
`
	cat := `
tags = ["devel"]

[[projects]]
name = "yerk"
domain = "personal"
remote = "git@example.com:salotz/yerk.git"
default_replica = "main"
tags = ["devel"]
`
	writeHostConfig(t, dir, cfg, cat)
	setYerkHostEnv(t, dir)

	var out, errBuf bytes.Buffer
	streams := cli.IO{Out: &out, Err: &errBuf}

	if err := cli.Execute(context.Background(), streams, []string{"get", "personal/yerk"}); err != nil {
		t.Fatalf("get project: %v\n%s", err, errBuf.String())
	}
	s := out.String()
	if !strings.Contains(s, "kind:\tProjectInfo") || !strings.Contains(s, "yerk://personal/yerk") {
		t.Fatalf("get project human:\n%s", s)
	}
	if !strings.Contains(s, yerkWS) || !strings.Contains(s, "placement.style:\tworkspace-dir") {
		t.Fatalf("get project fields:\n%s", s)
	}

	out.Reset()
	errBuf.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"get", "yerk/main", "--output", "json"}); err != nil {
		t.Fatalf("get replica json: %v", err)
	}
	js := out.String()
	if !strings.Contains(js, `"kind": "ReplicaInfo"`) || !strings.Contains(js, `"uri": "yerk://personal/yerk/main"`) {
		t.Fatalf("json:\n%s", js)
	}
	if !strings.Contains(js, yerkReplica) {
		t.Fatalf("json path:\n%s", js)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"project", "get", "yerk/main"}); err == nil {
		t.Fatal("project get with replica should error")
	}

	out.Reset()
	errBuf.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"replica", "get", "yerk", "main"}); err != nil {
		t.Fatalf("replica get: %v", err)
	}
	if !strings.Contains(out.String(), "kind:\tReplicaInfo") {
		t.Fatalf("replica get:\n%s", out.String())
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"lookup", deep}); err != nil {
		t.Fatalf("lookup: %v", err)
	}
	lu := out.String()
	if !strings.Contains(lu, "kind:\tReplicaInfo") || !strings.Contains(lu, "matchedPath:\t"+deep) {
		t.Fatalf("lookup deep:\n%s", lu)
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"project", "lookup", deep}); err != nil {
		t.Fatalf("project lookup: %v", err)
	}
	if !strings.Contains(out.String(), "kind:\tProjectInfo") {
		t.Fatalf("project lookup:\n%s", out.String())
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"replica", "lookup", yerkWS}); err == nil {
		t.Fatal("replica lookup workspace should error")
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"lookup", filepath.Join(dir, "nope")}); err == nil {
		t.Fatal("lookup unknown should error")
	}

	out.Reset()
	if err := cli.Execute(context.Background(), streams, []string{"get", "yerk", "--output", "yaml"}); err == nil {
		t.Fatal("unsupported output should error")
	}
}

// initFileRemote creates a non-bare git repo usable as a local file remote.
func initFileRemote(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
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
