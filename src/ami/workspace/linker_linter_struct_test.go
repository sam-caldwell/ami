package workspace

import (
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func testLinker_YAMLKeysPresent(t *testing.T) {
	l := Linker{Options: []string{"Optimize: 0"}}
	b, err := yaml.Marshal(l)
	if err != nil {
		t.Fatalf("yaml: %v", err)
	}
	if !strings.Contains(string(b), "options:") {
		t.Fatalf("expected options key; got: %s", string(b))
	}
}

func testLinter_YAMLKeysPresent(t *testing.T) {
	l := Linter{Options: []string{"strict"}, Rules: map[string]string{"W_IMPORT_ORDER": "warn"}, Suppress: map[string][]string{"./legacy": {"W_IDENT_UNDERSCORE"}}}
	b, err := yaml.Marshal(l)
	if err != nil {
		t.Fatalf("yaml: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "options:") {
		t.Fatalf("expected options key; got: %s", s)
	}
	if !strings.Contains(s, "rules:") {
		t.Fatalf("expected rules key; got: %s", s)
	}
}
