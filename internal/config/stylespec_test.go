package config_test

import (
	"testing"

	"github.com/salotz/yerk/internal/config"
)

func TestParseStyleValueString(t *testing.T) {
	t.Parallel()
	s, err := config.ParseStyleValue("name-tags")
	if err != nil || s.Name() != "name-tags" || s.HasParams() {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestParseStyleValueTable(t *testing.T) {
	t.Parallel()
	s, err := config.ParseStyleValue(map[string]any{
		"style":       "name-tags",
		"main_dir":    "~/.bimker",
		"replica_dir": "~/tree/x",
	})
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "name-tags" || s.MainDir != "~/.bimker" || s.ReplicaDir != "~/tree/x" {
		t.Fatalf("%+v", s)
	}
}

func TestParseStyleValueTableRequiresStyle(t *testing.T) {
	t.Parallel()
	_, err := config.ParseStyleValue(map[string]any{"main_dir": "/x"})
	if err == nil {
		t.Fatal("expected error")
	}
}
